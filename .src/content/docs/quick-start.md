+++
title = "Quick Start"
description = "Up and running in under a minute"
weight = 10
draft = false
toc = true
bref = "This article covers installation of Gilbert on your local machine"
+++

<h3 class="section-head" id="installation">
    <a href="#installation">Installation</a>
</h3>
<p>
    You can download release binaries from <a href="https://github.com/x1unix/gilbert/releases" target="_blank">releases page</a> or grab the latest version using `go get`:
</p>
```
go get -u github.com/x1unix/gilbert
```
<p>
    This command will install <b>Gilbert</b> tool into `$GOPATH/bin`
</p>


<h3 class="section-head" id="project-integration">
    <a href="#project-integration">Project integration</a>
</h3>
<p>
    **Gilbert** uses <code>gilbert.yaml</code> file to store list of tasks to run in project folder.
</p>
<p>
    To generate a sample <code>gilbert.yaml</code> file, navigate to your project directory in terminal and run <code>gilbert init</code> command:
</p>
```
$ cd $GOPATH/src/github.com/user/myproject
$ gilbert init
```
<p>
    This command will generate a sample file with <code>build</code> and <code>clean</code> tasks:
</p>

```yaml
version: "1.0"
vars:
  appVersion: 1.0.0
tasks:
  build:
  - description: Build project
    plugin: build
  clean:
  - if: file ./vendor
    description: Remove vendor files
    plugin: shell
    params:
      command: rm -rf ./vendor
```

<h3 class="section-head" id="available-tasks">
    <a href="#available-tasks">Available tasks</a>
</h3>
<p>
    To get list of available tasks, run
</p>
```bash
gilbert ls
```

<h3 class="section-head" id="running-tasks">
    <a href="#running-tasks">Running tasks</a>
</h3>
<p>
    To run a task declared in `gilbert.yaml`, use `gilbert run` command.
</p>
<p>
    <b>Example:</b>
```bash
gilbert run build
```
</p>
<h3 class="section-head" id="next">
    <a href="#next">Next steps</a>
</h3>
<p>
    We recommend to read about gilbert file <a href="../schema">syntax</a> documentation for more information.
</p>
<p>
    Also, you can find a good use-case example in <a href="https://github.com/x1unix/demo-go-plugins" target="_blank">this demo project</a>.<br />
    That repo shows usage of mixins and a few built-in plugins for a real-world web-server example.
</p>