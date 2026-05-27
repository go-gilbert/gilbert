---
agent_note: See eval_context.d.ts for expression context variables.
---

# Gilbert Expression Syntax

The Gilbert task runner supports expanding expressions.
Expressions are defined as YAML string values.

## Expression Types

### Shell call expression

Substitutes a `$(...)` section in a string with a result of a shell command inside of it.

Example:

- **Expression:** `"$(printf 2+2=%d 4)"`
- **Result:** `"4"`

### Language Expression

TODO: find a better name for this type of expression.

Runs an [Expr language][expr] expression inside `${{...}}` and returns its value.

[expr]: https://expr-lang.org/

#### Examples

##### Basic Operations

- **Expression:** `"${{ 2+2*3 }}"`
- **Result:** `8`

> [!NOTE]
> Unlike shell expressions, this example returns a numeric value.

##### Accessing Context Variables

Assume given a following workflow file:

```yaml
const:
  is_prod: true
```

- **Expression:** `"${{consts.is_prod}}"`
- **Result:** `true`

> [!NOTE]
> See [Available Context Variables](#available-context-variables) for a list of variables available inside expressions.

### Mixed Expression

Both shell and language expresssions can be mixed together.
This type of expression always return string.

Example:

- **Expression:** `"$(whoami)'s home dir is ${{env.HOME}}"`
- **Result:** `"root's home dir is /root"`

## Available Context Variables

List of variables available in language expressions:

| Property  | Description                                                                                     | Example                |
| --------- | ----------------------------------------------------------------------------------------------- | ---------------------- |
| `project` | see [docs below](#project)                                                                      | `${{project.workDir}}` |
| `consts`  | Holds values of constants declared in `const` section of `gilbert.yaml`                         | `${{consts.MY_CONST}}` |
| `env`     | Shell environment variables                                                                     | `${{env.HOME}}`        |
| `inputs`  | Holds values for input parameters.                                                              | `${{inputs.foobar}}`   |
| `matrix`  | Matrix values given passed to a job. See `job.strategy.matrix` of `gilbert.yaml`                | `${{matrix.goos}}`     |
| `event`   | Holds event args when action fires on signal. Available only in jobs defined inside `job.on.*`. | `${{event.error}}`     |

### `project`

Contains current working directory and a location from where task runner was started.

| Property       | Description                                         |
| -------------- | --------------------------------------------------- |
| `workDir`      | Path to a current working directory.                |
| `workspaceDir` | Path to a directory where `gilbert.yaml` is located |
| `workflowFile` | Path to a current `gilbert.yaml`.                   |
