**errors**

- cli: support rendering multiline diagnostics
- cli: handle zero column in diagnostics

**loader**

- yamlloader: pass indents from *ast.LiteralNode into expr.DocumentInfo
- yamlloader: check if global input flag is reserved for builtin
- yamlloader: check if task input flag is reserved for builtin or global

**fields**

- input binding delimiter

**cobra**

- Fix showing tasks list "Available Tasks" for "gilbert --help"

**expr**

- Support indents in DocumentInfo
- Consider cel-go as it might support dynamic envs: https://github.com/google/cel-go

**uflag**

- Bug when boolean flag w/o value is near other:

```
$ ./gilbert.sh run build --items=1,2,3 --release

error: invalid argument "--log-level=debug" for "--release" flag: strconv.ParseBool: parsing "--log-level=debug": invalid syntax
```
