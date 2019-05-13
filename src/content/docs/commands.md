+++
title = "Commands"
description = "Command line reference"
weight = 20
draft = false
toc = true
bref = "This article covers all information about Gilbert commands"
+++

{{<doc-section id="init" label="Init" >}}
```
gilbert init
```

Creates a `gilbert.yaml` file with basic tasks.

{{<doc-section id="list-tasks" label="List tasks" >}}
```
gilbert ls
```

Lists tasks defined in `gilbert.yaml` file.

{{<doc-section id="run-task" label="Run task" >}}

```
gilbert run task_name
```
Runs task by name and returns non-zero exit code if task fails.

#### Flags

You can set or override job variables with `--var` command.

**Example:**

```
gilbert run build-app --var foo=value1 --var bar=value2
```

{{<doc-section id="maintenance" label="Maintenance" >}}
Commands above are not related to job execution and used for managing Gilbert configuration and storage.

#### Version

```
gilbert version
```

Shows application version

#### Cache management

```
gilbert clean [--all | --plugins]
```

Cleans local Gilbert storage.

Storage used to store downloaded plugins and etc. Default storage location is `~/.gilbert`.
You can override storage location with `GILBERT_HOME` environment variable.

You should specify one or all of storage types to clear. Each storage type represented as command flag:

* `--all` - purge everything
* `--plugins` - purge downloaded plugins
