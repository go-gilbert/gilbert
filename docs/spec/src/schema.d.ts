/**
 * @schema
 * This file is a YAML schema description using TypeScript type syntax.
 * It is NOT part of the application codebase (which is in Go).
 * Do not generate TypeScript code from this file.
 *
 * Documentation conventions:
 *
 * - See @./expressions.md for expression language documentation.
 * - "@const" annotation states that property should be a scalar value and expressions are not supported.
 * - Properties that only allow expressions, use `Expression<T>` type, where `T` is a return type of an expression.
 */

/**
 * A Yaml string type which a value parseable with Go's time.ParseDuration.
 */
type DurationString = string

/**
 * YAML string with expression, execution of which returns a type T.
 * The `T` type carries a purely documentation role.
 *
 * Expression syntax is described in @./expressions.md
 *
 * Although most values can be a type of an expression string,
 * this type states that a value should be an expression and cannot be a static value.
 */
type Expression<T> = string

/**
 * Primitive value or an expression which can return a scalar value.
 */
type ScalarOrExpression = Expression<string | number | boolean> | string | number | boolean

interface JobAction {
  /**
   * Name of action to be executed.
   */
  action: string
}

interface JobMixin {
  /**
   * Name of mixin to be executed.
   */
  mixin: string
}

interface JobTask {
  /**
   * Name of task to be executed.
   */
  mixin: string
}

/**
 * Contains fields common to all job types.
 */
interface JobBase {
  /**
   * Whether to execute job concurrently.
   * When true, a next job will start, without waiting current one to finish.
   */
  async?: boolean

  /**
   * Duration of time to wait before starting a job.
   */
  delay?: DurationString

  /**
   * Path to a working directory where job will be executed.
   */
  ["working-directory"]?: string

  /**
   * Job execution timeout duration
   */
  timeout?: DurationString

  /**
   * Whether to allow job to fail.
   */
  ["continue-on-error"]?: boolean

  /**
   * Whether to run a job.
   *
   * If expression returns false - job is skipped.
   */
  if?: Expression<boolean>;

  /**
   * Custom environment variables to set for a job.
   *
   * Example:
   *
   * ```yaml
   * env:
   *  FOO: "bar"
   *  key: "${{inputs.myinput}}"
   * ```
   */
  env?: Record<string, ScalarOrExpression>

  /**
   * Enables matrix strategy.
   * Used to define variables whose combinations create separate job runs.
   * For example, node: [18, 20] and os: [ubuntu-latest, macos-latest] creates four runs.
   *
   * NOTE: "@const" annotation means that value cannot be an expression and should be a literal YAML value.
   */
  strategy?: {
    /**
     * Maximum number of matrix jobs to run concurrently.
     *
     * @const
     * @default 1
     */
    ["max-parallel"]?: number

    /**
     * Matrix of different job configurations.
     * The variables defined in a matrix, become properties in `matrix` context.
     *
     * Example:
     *
     * ```yaml
     * matrix:
     *  os: [darwin, linux, windows]
     *  arch: [amd64, arm64]
     * ```
     *
     * Matrix will populate `${{matrix.os}}` and `${{matrix.arch}}` properties to every job execution.
     *
     * @const
     */
    matrix: Record<string, Expression<any[]> | any[]>

    /**
     * List of combinations to exclude from job matrix.
     *
     * @const
     */
    exclude?: Record<string, any>[]
  }

  /**
   * Input parameters for action, mixin or task.
   *
   * Tasks and mixins declare expected parameters in "inputs" block.
   */
  with: Record<string, any>

  /**
   * Defines a set of jobs to run when certain hook is called.
   * Hooks are events produced by certain actions.
   *
   * For example, the `fs/watch` action has `changed` hook which is called when filesystem contents are changed.
   *
   * Besides that, runner itself has a builtin `error` hook to handle job execution failures (e.g. resource cleanup).
   */
  on?: Record<string, Job[]>
}

type JobType = JobAction | JobMixin | JobTask

/**
 * Job is a single execution unit of a task or mixin.
 * Job can start an action, run mixin or call a different task.
 *
 * TODO: improve description.
 */
type Job = JobType & JobBase

interface ListInputDefinition {
  /**
   * @const
   */
  type: 'list'

  /**
   * Defines a type of element of a list.
   *
   * Should only be used when `type` is set to `list`.
   *
   * @const
   */
  items: 'string' | 'int' | 'bool' | 'date' | 'duration' | 'float'
}

interface ScalarInputDefinition {
  /**
   * Input value type.
   *
   * Besides standard scalar values, it supports special types that are parceable from command-line flag or string value:
   *  - `duration`: time duration is nanoseconds. String value is parsed using Go's `time.ParseDuration` function.
   *  - `date`: Date and time. Parsed from a format defined in `dateFormat` field.
   *
   * @const
   */
  type: 'string' | 'int' | 'bool' | 'date' | 'duration' | 'float'
  items: never

  /**
   * Date format used to parse a command-line flag value.
   *
   * Has effect only when `type` is `date`.
   *
   * @const
   */
  dateFormat?: string
}

interface InputDefinitionCommon {
  /**
   * Fallback value used when input value not specified.
   *
   * Value type should correspond to `type`.
   * Note: use `optional` to use an empty value by default.
   */
  default?: Expression<any> | any

  /**
   * Whether a value is not required.
   *
   * When enabled and input value is unspecified - sets an empty value for input.
   *
   * Has no effect when `default` is set.
   *
   * @const
   */
  optional?: boolean

  /**
   * Binding property controls how input value is parsed from command-line
   * flags and environment variables.
   *
   * Works only for tasks. Has no effect in mixins.
   *
   * @const
   */
  binding?: {
    /**
     * Name of environment variable.
     *
     * If input value is unspecified and command-line isn't present,
     * value of a given environment variable will be used (if present).
     *
     * @const
     */
    env?: string

    /**
     * Overrides a name of command-line flag assigned to input.
     *
     * By default, each input gets a command-line flag with the same name as an input.
     *
     * @const
     */
    flag?: string

    /**
     * Character to be used to split a string into a list.
     *
     * Used when parsing a value from environment variable or command-line.
     *
     * Has no effect if input type is not `list`.
     *
     * @const
     */
    delimiter?: string
  }
}

/**
 * Defines task or mixin input parameter.
 * Parameters are mapped to command-line flags.
 *
 * Input parameter description can be documented with a comment block below input block.
 */
type InputDefinition = InputDefinitionCommon & (ListInputDefinition | ScalarInputDefinition)

/**
 * Mixins allow to decople a set of jobs into a reusable block.
 */
interface Mixin {
  /**
   * List of input parameters accepted by mixin.
   *
   * NOTE: unlike tasks, mixins don't support using command-line flags.
   * Input parameters for mixins should be explicitly passed via `with:` parameter in a job.
   *
   * @const
   */
  inputs?: Record<string, InputDefinition>

  /**
   * Custom environment variables to use for a mixin.
   *
   * Example:
   *
   * ```yaml
   * env:
   *  FOO: "bar"
   *  key: "${{inputs.myinput}}"
   * ```
   */
  env?: Record<string, ScalarOrExpression>

  /**
   * Sequence of jobs that will be executed when mixin is called.
   *
   * @const
   */
  steps: Job[]

  /**
   * Override a working directory for a mixin.
   * @const
   */
  ["working-directory"]?: string
}

interface Task {
  /**
   * List of input parameters accepted by a task.
   *
   * Input values can be passed to a task via command-line flags.
   *
   * @const
   */
  inputs?: Record<string, InputDefinition>

  /**
   * Custom environment variables to use for a task.
   *
   * Example:
   *
   * ```yaml
   * env:
   *  FOO: "bar"
   *  key: "${{inputs.myinput}}"
   * ```
   */
  env?: Record<string, ScalarOrExpression>

  /**
   * Sequence of jobs that will be executed when task is started.
   *
   * @const
   */
  steps: Job[]

  /**
   * Override a working directory for a task.
   * @const
   */
  ["working-directory"]?: string
}

interface WorkflowFile {
  /**
   * Workflow file version. Should be `2`.
   * @const
   */
  version: 2

  /**
   * List of other workflow files to include and merge.
   *
   * @const
   */
  include?: string[]

  /**
   * Environment variables to apply to all tasks.
   */
  env?: Record<string, ScalarOrExpression>

  /**
   * Set of plugins to import.
   *
   * Key is a namespace and value is import URL.
   *
   * For example, plugin with a following import, will expose its actions via `mydocker/` prefix:
   * 
   * ```yaml
   * plugins:
   *  mydocker: github://go-gilbert/gilbert-contrib/docker
   * tasks:
   *  foo:
   *    steps:
   *      - action: mydocker/run
   *        with:
   *          ...
   * ```
   *
   * @const
   */
  plugins?: Record<string, string>

  /**
   * List of predefined variables to be used in expressions.
   *
   * Values defined in this block are available as `${{consts.*}}`
   *
   * Value should not be an expression.
   *
   * @const
   */
  const?: Record<string, any>

  /**
   * Global input parameters.
   *
   * @see [Task.inputs]
   * @const
   */
  inputs?: Record<string, InputDefinition>

  /**
   * Map of task name and its definition.
   *
   * Task description can be documented via a comment block below its name:
   *
   * ```yaml
   * tasks:
   *  # Build the project
   *  build:
   *    steps:
   *      ...
   * ```
   *
   * Documented task description will be present in `gilbert list` output.
   *
   * Task can be called by name using `gilbert run <taskname>`.
   *
   * Use `gilbert list` to display a list of available tasks.
   * Use `gilbert run <taskname> --help` see task description and its parameters.
   */
  tasks?: Record<string, Task>

  /**
   * Mixins are reusable pieces of pipeline which can accept inputs.
   *
   * Unlike tasks, they cannot be called from command-line.
   */
  mixins?: Record<string, Mixin>
}
