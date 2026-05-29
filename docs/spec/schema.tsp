/** gilbert.yaml workflow file. Generated from src/schema.d.ts **/

import "@typespec/json-schema";

using JsonSchema;

namespace Schemas;

/**
 * A YAML string parseable by Go's time.ParseDuration.
 *
 * Examples: `300ms`, `10s`, `5m`, `2h45m`, `1.5h`.
 */
alias DurationString = string;

/**
 * YAML string containing a Gilbert expression.
 *
 * Gilbert supports three expression forms:
 *
 * - Shell expression: `$(...)`. The body is executed by the system shell,
 *   has a fixed 30 second timeout, and runs in `project.workDir`.
 * - Value expression: `${{ ... }}`. The body is evaluated by the Expr
 *   language. When this is the entire YAML value, it can return a typed value
 *   such as a boolean, number, string, list, or object.
 * - Mixed expression: literal text combined with one or more shell or value
 *   expressions. Mixed expressions always produce a string. Value-expression
 *   results inside mixed expressions are cast to strings; objects and lists
 *   cannot be cast.
 *
 * The type parameter from schema.d.ts is documentation-only, so the JSON
 * schema representation is a string.
 */
alias Expression = string;

/** Any YAML value. */
alias AnyValue = unknown;

/** YAML object with string keys and arbitrary YAML values. */
alias AnyRecord = Record<AnyValue>;

/**
 * Contains fields common to all job types.
 */
model JobBase {
  /**
   * Whether to execute job concurrently.
   * When true, the next job will start without waiting for the current one to finish.
   */
  async?: boolean;

  /** Duration of time to wait before starting a job. */
  delay?: DurationString;

  /**
   * Path to a working directory where job will be executed.
   * Relative paths resolve from the current workflow working directory.
   */
  `working-directory`?: string;

  /** Job execution timeout duration. */
  timeout?: DurationString;

  /** Whether to allow job to fail. */
  `continue-on-error`?: boolean;

  /**
   * Whether to run a job.
   * If expression returns false, the job is skipped.
   *
   * This field must be an expression. Static literal values are invalid.
   */
  `if`?: Expression;

  /**
   * Custom environment variables to set for a job.
   *
   * This block is literal. Expressions are not evaluated in keys or values.
   */
  env?: AnyRecord;

  /**
   * Matrix execution strategy.
   * Expands this job into the Cartesian product of matrix values.
   */
  strategy?: Strategy;

  /**
   * Input parameters for action, mixin or task.
   * Map keys are literal. Expressions are allowed recursively in values.
   */
  with: AnyRecord;

  /**
   * Defines a set of jobs to run when a certain hook is called.
   * Hooks are events produced by certain actions.
   *
   * The runner also provides a built-in `error` hook for handling job
   * failures. Jobs inside hook lists receive the `event` expression context.
   */
  on?: Record<Job[]>;
}

/** Matrix execution strategy. */
model Strategy {
  /**
   * Maximum number of matrix jobs to run concurrently.
   *
   * Default: 1.
   */
  `max-parallel`?: int32;

  /**
   * Matrix of different job configurations.
   * Variables defined in a matrix become properties in the `matrix` expression context.
   *
   * Each key becomes available as `matrix.<key>` during each matrix job run.
   * Values may be literal arrays or expressions that evaluate to arrays.
   * Empty arrays are ignored with a warning.
   */
  matrix: Record<MatrixValues>;

  /**
   * List of combinations to exclude from job matrix.
   * A rule matches only when all key/value pairs in the rule match the matrix
   * row. Exclude values and referenced matrix values must be scalar and
   * comparable.
   */
  exclude?: AnyRecord[];
}

/** Matrix values may be a literal array or an expression returning an array. */
alias MatrixValues = Expression | AnyValue[];

/** Job that executes an action. */
model ActionJob extends JobBase {
  /**
   * Name of action to be executed.
   * Built-in actions use `namespace/name` names such as `debug/echo` and
   * `go/build`. Plugin actions use the namespace declared in `plugins`, for
   * example `docker/build`.
   */
  action: string;
}

/** Job that executes a mixin. */
model MixinJob extends JobBase {
  /**
   * Name of mixin to be executed.
   * Mixin inputs must be passed explicitly with `with`; command-line input
   * bindings do not apply to mixin calls.
   */
  mixin: string;
}

/** Job that executes a task. */
model TaskJob extends JobBase {
  /**
   * Name of task to be executed.
   * Calls another task from the current workflow, enabling nested task calls.
   */
  task: string;
}

/**
 * Job is a single execution unit of a task or mixin.
 * Job can start an action, run a mixin, or call a different task.
 */
alias Job = ActionJob | MixinJob | TaskJob;

/** Input value types supported by task and mixin inputs. */
alias InputValueType = "string" | "int" | "bool" | "date" | "duration" | "float";

/**
 * List item value types supported by list inputs.
 * Nested complex types are not supported: a list item can be scalar, but not
 * another list or dictionary.
 */
alias ListItemValueType = InputValueType;

/** Command-line and environment binding for an input value. */
model InputBinding {
  /**
   * Environment variable name.
   * If input value and command-line flag are unspecified, this variable is used when present.
   *
   * Task input resolution order is: explicit `with` value, command-line flag,
   * environment variable, then `default`. Bindings apply only to task inputs,
   * not mixin inputs.
   */
  env?: string;

  /**
   * Overrides the command-line flag name assigned to an input.
   * By default, each input gets a command-line flag with the same name as the input.
   */
  flag?: string;

  /**
   * Character used to split a string into a list when parsing an environment
   * variable or command-line value.
   *
   * Current loader support is not implemented.
   */
  delimiter?: string;
}

/** Fields common to all input definitions. */
model InputDefinitionCommon {
  /**
   * Fallback value used when input value is not specified.
   * This value may be an expression. After expression expansion, the value
   * must match the declared `type`.
   */
  default?: AnyValue;

  /**
   * Whether a value is not required.
   * Has no effect when `default` is set.
   *
   * When `optional` is true and there is no default, the input resolves to the
   * declared type's zero value. Boolean inputs are optional by default.
   */
  optional?: boolean;

  /**
   * Controls how input value is parsed from command-line flags and environment variables.
   * Works only for tasks. Has no effect in mixins.
   */
  binding?: InputBinding;
}

/** Input definition for list values. */
model ListInputDefinition extends InputDefinitionCommon {
  /** Declares this input as a list value. */
  type: "list";

  /**
   * Defines a type of element of a list.
   * Should only be used when `type` is set to `list`.
   *
   * List item types must be scalar. Nested lists and dictionaries are not
   * supported.
   */
  items: ListItemValueType;
}

/** Input definition for scalar values. */
model ScalarInputDefinition extends InputDefinitionCommon {
  /**
   * Input value type.
   *
   * Besides standard scalar values, special values are parseable from
   * command-line flags or string values. `duration` values are parsed with
   * Go's time.ParseDuration. `date` values are parsed with the Go time layout
   * from `dateFormat`, or RFC3339 when `dateFormat` is omitted.
   */
  type: InputValueType;

  /**
   * Date format used to parse a command-line flag value.
   * Has effect only when `type` is `date`.
   * Defaults to Go time.RFC3339 when omitted.
   */
  dateFormat?: string;
}

/** Defines task or mixin input parameter. */
alias InputDefinition = ListInputDefinition | ScalarInputDefinition;

/**
 * Mixins allow a set of jobs to be reused as a block.
 * Mixin inputs are set only through `with` from a calling task or mixin;
 * command-line flag and environment bindings are ignored for mixins.
 */
model Mixin {
  /**
   * List of input parameters accepted by mixin.
   * Unlike tasks, mixins do not support command-line flags. Mixins can be
   * called from tasks and from other mixins, and their inputs should be passed
   * explicitly via `with`.
   */
  inputs?: Record<InputDefinition>;

  /** Custom environment variables to use for a mixin. */
  env?: AnyRecord;

  /** Sequence of jobs that will be executed when mixin is called. */
  steps: Job[];

  /** Override a working directory for a mixin. */
  `working-directory`?: string;
}

/** Task definition. */
model Task {
  /**
   * List of input parameters accepted by a task.
   * Input values can come from explicit `with` values, command-line flags,
   * environment variables, or defaults.
   */
  inputs?: Record<InputDefinition>;

  /** Custom environment variables to use for a task. */
  env?: AnyRecord;

  /** Sequence of jobs that will be executed when task is started. */
  steps: Job[];

  /** Override a working directory for a task. */
  `working-directory`?: string;
}

/** Gilbert workflow file. */
@jsonSchema
@id("gilbert.schema.json")
model WorkflowFile {
  /** Workflow file version. `2` is the only supported value; other values are invalid. */
  version: 2;

  /**
   * List of other workflow files to include and merge.
   * Paths resolve relative to the including file. A file cannot include
   * itself. Duplicate task, mixin, and input names across included files are
   * rejected.
   */
  include?: string[];

  /**
   * Set of plugins to import.
   * Key is a namespace and value is import URL.
   *
   * The namespace becomes the action prefix. For example, a plugin declared as
   * `docker:` exposes actions such as `docker/build` and `docker/run`.
   */
  plugins?: Record<string>;

  /**
   * Predefined variables available to expressions as `${{ consts.* }}`.
   * Expressions are not expanded inside this block. Values are literal YAML
   * scalars.
   */
  `const`?: AnyRecord;

  /** Global input parameters. */
  inputs?: Record<InputDefinition>;

  /**
   * Map of task name and its definition.
   * Tasks are command-line entry points and can be run with
   * `gilbert run <name>`. Comments immediately above a task key are collected
   * as the task description for list/help output.
   */
  tasks?: Record<Task>;

  /**
   * Mixins are reusable pieces of pipeline which can accept inputs.
   * Unlike tasks, they cannot be called from command-line. Mixins can be
   * called from both tasks and other mixins.
   */
  mixins?: Record<Mixin>;
}
