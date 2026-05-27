import "@typespec/json-schema";

using JsonSchema;

@jsonSchema
namespace Schemas;

/** A YAML string parseable by Go's time.ParseDuration. */
alias DurationString = string;

/**
 * YAML string containing a Gilbert expression.
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

  /** Path to a working directory where job will be executed. */
  `working-directory`?: string;

  /** Job execution timeout duration. */
  timeout?: DurationString;

  /** Whether to allow job to fail. */
  `continue-on-error`?: boolean;

  /**
   * Whether to run a job.
   * If expression returns false, the job is skipped.
   */
  `if`?: Expression;

  /** Custom environment variables to set for a job. */
  env?: AnyRecord;

  /** Matrix execution strategy. */
  strategy?: Strategy;

  /** Input parameters for action, mixin or task. */
  with: AnyRecord;

  /**
   * Defines a set of jobs to run when a certain hook is called.
   * Hooks are events produced by certain actions.
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
   */
  matrix: Record<MatrixValues>;

  /** List of combinations to exclude from job matrix. */
  exclude?: AnyRecord[];
}

/** Matrix values may be a literal array or an expression returning an array. */
alias MatrixValues = Expression | AnyValue[];

/** Job that executes an action. */
model ActionJob extends JobBase {
  /** Name of action to be executed. */
  action: string;
}

/** Job that executes a mixin. */
model MixinJob extends JobBase {
  /** Name of mixin to be executed. */
  mixin: string;
}

/** Job that executes a task. */
model TaskJob extends JobBase {
  /** Name of task to be executed. */
  task: string;
}

/**
 * Job is a single execution unit of a task or mixin.
 * Job can start an action, run a mixin, or call a different task.
 */
alias Job = ActionJob | MixinJob | TaskJob;

/** Input value types supported by task and mixin inputs. */
alias InputValueType = "string" | "int" | "bool" | "date" | "duration" | "float";

/** List item value types supported by list inputs. */
alias ListItemValueType = InputValueType | "list";

/** Command-line and environment binding for an input value. */
model InputBinding {
  /**
   * Environment variable name.
   * If input value and command-line flag are unspecified, this variable is used when present.
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
   */
  delimiter?: string;
}

/** Fields common to all input definitions. */
model InputDefinitionCommon {
  /**
   * Fallback value used when input value is not specified.
   * Value type should correspond to `type`.
   */
  default?: AnyValue;

  /**
   * Whether a value is not required.
   * Has no effect when `default` is set.
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
  type: "list";

  /**
   * Defines a type of element of a list.
   * Should only be used when `type` is set to `list`.
   */
  items: ListItemValueType;
}

/** Input definition for scalar values. */
model ScalarInputDefinition extends InputDefinitionCommon {
  /**
   * Input value type.
   *
   * Besides standard scalar values, special values are parseable from
   * command-line flags or string values.
   */
  type: InputValueType;

  /**
   * Date format used to parse a command-line flag value.
   * Has effect only when `type` is `date`.
   */
  dateFormat?: string;
}

/** Defines task or mixin input parameter. */
alias InputDefinition = ListInputDefinition | ScalarInputDefinition;

/** Mixins allow a set of jobs to be reused as a block. */
model Mixin {
  /**
   * List of input parameters accepted by mixin.
   * Unlike tasks, mixins do not support command-line flags.
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
   * Input values can be passed to a task via command-line flags.
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
model WorkflowFile {
  /** Workflow file version. Should be `2`. */
  version: 2;

  /** List of other workflow files to include and merge. */
  include?: string[];

  /**
   * Set of plugins to import.
   * Key is a namespace and value is import URL.
   */
  plugins?: Record<string>;

  /**
   * Predefined variables available to expressions as `${{ consts.* }}`.
   * Values in this block should not be expressions.
   */
  `const`?: AnyRecord;

  /** Global input parameters. */
  inputs?: Record<InputDefinition>;

  /** Map of task name and its definition. */
  tasks?: Record<Task>;

  /**
   * Mixins are reusable pieces of pipeline which can accept inputs.
   * Unlike tasks, they cannot be called from command-line.
   */
  mixins?: Record<Mixin>;
}
