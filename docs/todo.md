**errors**

- cli: support rendering multiline diagnostics
- cli: handle zero column in diagnostics

**loader**

- yamlloader: pass indents from \*ast.LiteralNode into expr.DocumentInfo
- yamlloader: check if global input flag is reserved for builtin
- yamlloader: check if task input flag is reserved for builtin or global

**fields**

- input binding delimiter

**schema**

- wire `InputDefinition.binding.delimiter` to yamlloader
- support setting env

**expr**

- Support indents in DocumentInfo
- Consider cel-go as it might support dynamic envs: https://github.com/google/cel-go
