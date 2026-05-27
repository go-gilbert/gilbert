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
consts:
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

- **Expression:**: `"$(whoami)'s home dir is ${{env.HOME}}"`
- **Result:**: `"root's home dir is /root"`

## Available Context Variables

### `project`

Contains
