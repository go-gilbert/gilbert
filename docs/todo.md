**errors**

- expr: errors into diagnostics
- expr: handle multiline strings in errors
- yamlloader: yaml parse errors into diagnostics

**loader**

- yamlloader: check if global input flag is reserved for builtin
- yamlloader: check if task input flag is reserved for builtin or global

**fields**

- input binding delimiter

**cobra**

- Fix showing tasks list "Available Tasks" for "gilbert --help"

**uflag**

- Bug when boolean flag w/o value is near other:

```
$ ./gilbert.sh run build --items=1,2,3 --release

error: invalid argument "--log-level=debug" for "--release" flag: strconv.ParseBool: parsing "--log-level=debug": invalid syntax
```
