# Gilbert Workflow File Specification

This document specifies the Gilbert workflow file format, the input schema model, and the expression language available inside workflow values.

The workflow file is a YAML document named `gilbert.yaml` or `gilbert.yml`. It declares executable tasks, reusable mixins, plugins, constants, and input definitions.

Reference material:

- `src/schema.d.ts` describes the YAML shape in TypeScript syntax for documentation.
- `src/eval_context.d.ts` describes variables available to expression evaluation.
- `src/expressions.md` and `src/types.md` describe expression and input-type behavior.
- `src/gilbert.yml` is an example workflow.

## Document Shape

```yaml
version: "2"
include: []
plugins: {}
const: {}
inputs: {}
mixins: {}
tasks: {}
```

All top-level sections except `version` are optional unless a task needs them. Unknown fields are invalid when the loader validates a structured block.

| Field | Required | Type | Expressions | Description |
| --- | --- | --- | --- | --- |
| `version` | Yes | string or number | No | Workflow format version. The supported value is `2`. |
| `include` |  | list of strings | No | Other workflow files to load and merge into the current workflow. Paths are resolved relative to the file that declares them. |
| `plugins` |  | map of string to string | No | Plugin imports. The map key is the action namespace exposed by the plugin. |
| `const` |  | map of scalar values | No | Static values exposed to expressions as `consts.*`. |
| `inputs` |  | map of [input definitions](#input-definition) | Per definition | Global input definitions. |
| `mixins` |  | map of [mixin definitions](#tasks-and-mixins) | Per mixin field | Reusable job groups that can be called from tasks or other mixins. |
| `tasks` |  | map of [task definitions](#tasks-and-mixins) | Per task field | Runnable entry points. A task can be executed with `gilbert run <task-name>`. |

## Includes

`include` is a list of YAML file paths:

```yaml
include:
  - ./common.yml
```

Each path must point to a file, cannot point to the current file, and is resolved relative to the including file. Included workflows are merged into the current workflow. Duplicate definitions are rejected for sections that require unique names.

## Plugins

`plugins` maps a namespace to an import URL:

```yaml
plugins:
  docker: github://go-gilbert/gilbert-contrib/docker
  local: ./plugins/local
```

Actions exported by the plugin are referenced with the namespace prefix:

```yaml
tasks:
  image:
    steps:
      - action: docker/build
        with:
          context: .
```

## Constants

`const` defines scalar values available to `${{ ... }}` expressions through the `consts` context:

```yaml
const:
  release_channel: stable
  publish: true
```

Constants are literal YAML scalars. Expressions are not expanded inside `const`.

## Inputs

Inputs declare values accepted by the workflow, tasks, and mixins.

```yaml
inputs:
  release:
    type: bool
    default: false

  timeout:
    type: duration
    default: 10s
    binding:
      env: JOB_TIMEOUT
      flag: timeout
```

Task inputs can be set from command-line flags, environment variables, defaults, or explicit `with` values. Mixin inputs are passed explicitly from the calling job through `with`; command-line bindings do not apply to mixins.

### Input Definition

| Field | Required | Type | Expressions | Description |
| --- | --- | --- | --- | --- |
| `type` | Yes | [input type](#input-types) | No | One of [`string`](#string), [`int`](#int), [`bool`](#bool), [`float`](#float), [`date`](#date), [`duration`](#duration), or [`list`](#list). |
| `items` | when `type: list` | [type schema](#type-schema) | No | Invalid otherwise. Defines the list element type. |
| `dateFormat` |  | string | No | Date parsing layout. Only valid for [`date`](#date). Defaults to Go `time.RFC3339`. |
| `default` |  | value or [expression](#expression-language) | Yes | Fallback value when no value is provided. Must match the input type after expansion. |
| `optional` |  | bool | No | Allows an omitted input to resolve to the type's zero value. Has no effect when `default` is set. Boolean inputs are optional by default. |
| `binding.env` |  | string | No | Environment variable used when no explicit input or flag is provided. |
| `binding.flag` |  | string | No | Command-line flag name. Defaults to the input name when omitted. |
| `binding.delimiter` |  | string | No | Intended delimiter for parsing list values from flags or environment variables. Current loader support is not implemented. |

### Input Types

| Type | YAML representation | Runtime value |
| --- | --- | --- |
| [`string`](#string) | string | string |
| [`int`](#int) | number | signed 64-bit integer |
| [`float`](#float) | number | 64-bit floating-point number |
| [`bool`](#bool) | boolean | boolean |
| [`date`](#date) | string | parsed with `dateFormat` or RFC3339 |
| [`duration`](#duration) | string | parsed by Go `time.ParseDuration`, for example `300ms`, `10s`, `5m` |
| [`list`](#list) | YAML sequence or delimited input value | list of typed elements |

List inputs use an `items` schema:

```yaml
inputs:
  ports:
    type: list
    items:
      type: int
```

Nested complex item types are not supported. A list item type can be scalar, but not another list or dictionary.

#### String

`string` is text. YAML strings may be plain, quoted, or block scalars.

```yaml
name:
  type: string
  default: Gilbert
```

#### Int

`int` is a signed 64-bit integer.

```yaml
retries:
  type: int
  default: 3
```

#### Float

`float` is a 64-bit floating-point number.

```yaml
threshold:
  type: float
  default: 0.75
```

#### Bool

`bool` is a boolean value. Boolean inputs are optional by default.

```yaml
release:
  type: bool
  default: false
```

#### Date

`date` is a date/time value parsed from a string. By default, values use Go `time.RFC3339`. A custom Go time layout can be set with `dateFormat`.

```yaml
publishedAt:
  type: date
  dateFormat: "2006-01-02 15:04:05"
```

#### Duration

`duration` is a time duration parsed from a string with Go `time.ParseDuration`. Supported examples include `300ms`, `10s`, `5m`, `2h45m`, and `1.5h`.

Duration values are also used by workflow fields described as a [duration string](#duration), such as job `delay` and `timeout`.

```yaml
timeout:
  type: duration
  default: 10s
```

#### List

`list` is an array of typed elements. List inputs must declare an `items` [type schema](#type-schema).

```yaml
ports:
  type: list
  items:
    type: int
```

### Type Schema

A type schema is the object form used by `items` to describe a list element:

```yaml
items:
  type: int
```

| Field | Required | Type | Expressions | Description |
| --- | --- | --- | --- | --- |
| `type` | Yes | [input type](#input-types) | No | Element type. Complex nested types such as [`list`](#list) are not supported. |
| `dateFormat` |  | string | No | Date parsing layout. Only valid when `type` is [`date`](#date). |

## Tasks And Mixins

Tasks and mixins share the same job-group shape:

```yaml
tasks:
  build:
    inputs: {}
    env: {}
    working-directory: .
    steps: []

mixins:
  hello:
    inputs: {}
    env: {}
    working-directory: .
    steps: []
```

| Field | Required | Type | Expressions | Description |
| --- | --- | --- | --- | --- |
| `inputs` |  | map of [input definitions](#input-definition) | Per definition | Values accepted by this task or mixin. |
| `env` |  | map | No | Environment values scoped to the task or mixin. |
| `working-directory` |  | string | No | Working directory for jobs in the group. Relative paths resolve from the current workflow working directory. |
| `steps` | Yes | list of [jobs](#jobs) | Per job field | Jobs executed in order unless a job is asynchronous. |

Tasks are command-line entry points. Mixins are reusable job groups and cannot be run directly from the command line.

Comments immediately above task, mixin, and input keys may be collected as documentation by tooling.

## Jobs

A job is one step in a task, mixin, or event hook. It must target exactly one executable unit.

```yaml
steps:
  - action: debug/echo
    with:
      message: Hello

  - mixin: hello
    with:
      name: Bob

  - task: test:subtask
    with:
      message: nested task call
```

### Job Target

Exactly one target field must be set.

| Field | Required | Type | Expressions | Description |
| --- | --- | --- | --- | --- |
| `action` | one target | string | No | Action name. Built-in and plugin actions use namespace-style names such as `debug/echo` or `go/build`. |
| `mixin` | one target | string | No | Mixin name to execute. |
| `task` | one target | string | No | Task name to execute. |

### Job Fields

| Field | Required | Type | Expressions | Description |
| --- | --- | --- | --- | --- |
| `async` |  | bool | No | When true, the runner starts the next job without waiting for this job to finish. |
| `delay` |  | [duration string](#duration) | No | Time to wait before starting the job. |
| `working-directory` |  | string | No | Working directory for this job. |
| `timeout` |  | [duration string](#duration) | No | Maximum execution time for this job. |
| `continue-on-error` |  | bool | No | When true, task execution continues after this job fails. |
| `if` |  | [expression](#expression-language) | Required expression | Conditional expression. If it evaluates to false, the job is skipped. Static literal values are invalid here. |
| `env` |  | map | No | Environment values scoped to this job. |
| `strategy` |  | [strategy object](#matrix-strategy) | Per strategy field | Matrix execution strategy. |
| `with` |  | map | Yes, recursively in values | Arguments passed to the action, mixin, or task. Map keys are literal. |
| `on` |  | map of string to job list | Per hook job field | Event hooks fired by this job or by the runner. Hook names are literal. |

## Matrix Strategy

`strategy` expands one job into multiple job runs by computing the Cartesian product of matrix values.

```yaml
steps:
  - action: go/build
    strategy:
      max-parallel: 2
      matrix:
        os: [windows, linux, darwin]
        arch: [amd64, arm64]
      exclude:
        - os: darwin
          arch: amd64
    with:
      os: ${{ matrix.os }}
      arch: ${{ matrix.arch }}
```

| Field | Required | Type | Expressions | Description |
| --- | --- | --- | --- | --- |
| `max-parallel` |  | positive integer | No | Maximum number of matrix jobs to run concurrently. Defaults to `1`. |
| `matrix` | Yes | ordered map of string to list or expression returning list | Yes, in values | Each key becomes available under `matrix.<key>` during each job run. Matrix keys are literal. |
| `exclude` |  | list of scalar maps | No | Excludes matrix rows that match all specified key/value pairs in an exclude rule. |

`matrix` values may be literal arrays or expressions that evaluate to arrays. Empty matrix arrays are ignored with a warning. `exclude` values and the referenced matrix values must be scalar and comparable.

## Event Hooks

Jobs may define `on` hooks. Each hook maps an event name to jobs that run when the event is emitted.

```yaml
steps:
  - action: debug/signal
    with:
      signal: changed
    on:
      changed:
        - action: debug/echo
          with:
            message: "event: ${{ event.name }}"
      error:
        - action: debug/echo
          with:
            message: "job failed"
```

Actions may define their own hook names. The runner also supports an `error` hook for handling job failures.

Jobs inside a hook receive the `event` expression context.

## Expression Language

Gilbert expands dynamic expressions in YAML string values where the schema allows lazy values. Expressions are parsed before execution and evaluated at runtime.

There are three expression forms:

| Form | Example | Result type |
| --- | --- | --- |
| Shell expression | `$(git describe --tags --abbrev=0)` | string |
| Value expression | `${{ inputs.release ? "release" : "debug" }}` | value returned by Expr |
| Mixed expression | `'some string, ${{ "expres" + "sion" }}, and $(echo shell)'` | string |

When a string contains multiple literal or expression parts, it is a mixed expression. The parts are concatenated and the final value is a string:

```yaml
message: "$(whoami)'s workspace is ${{ project.workspaceDir }}"
```

Values returned from `${{ ... }}` inside a mixed expression are cast to strings. String, number, and boolean values can be cast. Objects and lists cannot be cast and cause expression evaluation to fail.

When a string contains exactly one value expression, the expression may return a non-string value:

```yaml
cgo: ${{ inputs.release }}
```

Terminology note: this document uses "value expression" for `${{ ... }}` because the expression can return typed values when it is the whole YAML value. Other reasonable names are "Expr expression", "language expression", and "runtime expression"; "value expression" is the least awkward and describes the behavior most directly.

### Shell Expressions

Shell expressions use `$(...)`:

```yaml
version:
  type: string
  default: $(git describe --tags --abbrev=0)
```

The body is executed by the system shell in the current expression scope. The working directory is `project.workDir`, and the command receives the current environment. Command evaluation has a fixed timeout of 30 seconds.

Shell expressions can contain value expressions in their body:

```yaml
value: $(echo "${{ inputs.name }}")
```

### Value Expressions

Value expressions use `${{ ... }}` and are evaluated by the [Expr](https://expr-lang.org/) language:

```yaml
if: ${{ inputs.version != "invalid" }}
```

Expr syntax supports operators, conditionals, function calls, selectors, indexing, and built-ins provided by Expr. See the Expr [language definition](https://expr-lang.org/docs/language-definition) and [functions](https://expr-lang.org/docs/functions) documentation for the syntax and standard functions available from Expr.

Gilbert supplies the evaluation context described below.

Nested Gilbert expressions are not allowed inside the body of a `${{ ... }}` expression.

### Mixed Expressions

Mixed expressions combine literal text with one or more shell or value expressions:

```yaml
message: 'some string, ${{ "expres" + "sion" }}, and $(echo shell)'
```

Mixed expressions always produce a string. Shell expression results are already strings. Value expression results are converted to strings when possible; conversion fails for objects and lists.

### Expression Context

| Name | Type | Available when | Description |
| --- | --- | --- | --- |
| [`project`](#project-context) | object | always | Project paths and current working directory. |
| `env` | map of string to string | always | System environment variables. |
| `consts` | map | when constants exist | Values from the workflow `const` block. |
| `inputs` | map | inside task, mixin, job, and default evaluation scopes | Resolved input values. |
| `matrix` | map | inside matrix job runs | Current matrix row values. |
| `event` | map | inside jobs declared under `on` hooks | Event metadata emitted by an action or runner hook. |

Input and constant lookups are inherited through nested scopes. A child scope can see values from its parent scope unless it overrides the same key.

### Project Context

`project` contains paths for the current workflow execution:

| Name | Type | Description |
| --- | --- | --- |
| `project.workDir` | string | Current job working directory. |
| `project.workspaceDir` | string | Directory containing the workflow file. |
| `project.workflowFile` | string | Path to the current workflow file. |

### Expression Placement

Each schema table includes an `Expressions` column. The general rules are:

- `No` means the value is literal and expressions are not evaluated.
- `Yes` means the field accepts shell expressions, value expressions, and mixed expressions.
- `Required` means the field must be an expression and cannot be a static literal.
- `Per ... field` means the object itself is structural, and expression support is determined by each child property.
- `recursively in values` means expressions are allowed in nested map and list values, but map keys remain literal.

Common dynamic value positions include:

- input `default`
- job `if`
- job `with` values
- matrix values
- nested arrays and maps inside dynamic values

## JSON Schema Representation

The workflow schema can be represented as JSON Schema for structural validation, but not as a complete replacement for the current TypeScript documentation and Go loader rules.

A JSON Schema can model:

- top-level sections and required `version`
- task, mixin, job, strategy, and input object shapes
- allowed scalar input type names
- `oneOf` rules for `action`, `mixin`, and `task` targets
- `if`/`then` rules such as requiring `items` when `type` is `list`
- `patternProperties` for name-to-definition maps
- duration and expression fields as strings

A JSON Schema cannot fully model, without custom keywords or runtime validation:

- parsing Go duration strings with exact `time.ParseDuration` behavior
- parsing dates with a user-provided Go layout in `dateFormat`
- validating Expr syntax inside `${{ ... }}`
- validating shell command expressions inside `$(...)`
- distinguishing all literal-only fields from expression-capable fields by runtime parser behavior
- enforcing include-file existence, recursive include checks, and merge conflicts
- validating action-specific `with` schemas from plugins
- checking matrix exclude comparability and whether exclude keys are present in the matrix
- collecting comments as task/input documentation

Recommended approach: generate or maintain a JSON Schema for editor assistance and early structural diagnostics, while keeping the TypeScript schema and Go loader as the normative source for semantic validation. Use custom schema extensions such as `x-gilbert-expression`, `x-gilbert-const`, and `x-gilbert-duration` if tooling needs to preserve Gilbert-specific semantics.
