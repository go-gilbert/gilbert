# Workflow File

Workflow file is a YAML file named `gilbert.yaml` which contains definitions of executable tasks.

## Schema

See: @./schema.d.ts

## Input Types

Gilbert introduces a set of types available for inputs. Types above are convertable from/to YAML:

| Name       | Description                                                                                   | YAML equivalent            |
| ---------- | --------------------------------------------------------------------------------------------- | -------------------------- |
| `string`   | String                                                                                        | Strings, including heredoc |
| `int`      | signed 64-bit integer                                                                         | number                     |
| `float`    | floating point number                                                                         | number                     |
| `date`     | Date and time                                                                                 | string                     |
| `duration` | Time diration in nanoseconds. String values are parsed using `time.ParseDuration` Go function | string                     |
| `list`     | Array of elements                                                                             | List block                 |

## Expressions

See @./expressions.md
