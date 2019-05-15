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
<h4>Linux, macOS and FreeBSD</h4>
<p>
    Run this command to install Gilbert on your system:
```bash
curl https://raw.githubusercontent.com/go-gilbert/gilbert/master/install.sh | sh
```

Default installation path is `$GOPATH/bin`
<h4>Windows</h4>
<p>
    You can download release binaries from <a href="https://github.com/go-gilbert/gilbert/releases" target="_blank">releases page</a> or install using PowerShell script:
</p>
```powershell
Invoke-Expression (Invoke-Webrequest 'https://raw.githubusercontent.com/go-gilbert/gilbert/master/install.ps1' -UseBasicParsing).Content
```
<p><b>Note:</b> You should run <code>Set-ExecutionPolicy Bypass</code> in PowerShell to be able to execute installation script.</p>


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
    This command will generate a sample file with <code>build</code>,<code>clean</code> and <code>cover</code> tasks:
</p>

```yaml
version: "1.0"
vars:
  appVersion: 1.0.0
tasks:
  build:
  - description: Build project
    action: build
  clean:
  - if: file ./vendor
    description: Remove vendor files
    action: shell
    params:
      command: rm -rf ./vendor
  cover:
  - description: Check project coverage
    action: cover
    params:
      reportCoverage: true
      threshold: 60
      packages:
      - ./...
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

See <a href="../commands/#run-task">command docs</a> for more information.
</p>

{{<doc-section id="next-steps" label="Next Steps" >}}

We recommend to take a look at theese links to get more information about Gilbert:

#### Documentation

* [File Syntax](../syntax)
* [Built-in Actions](../actions)

#### Examples

* [Sample project](ttps://github.com/x1unix/demo-go-plugins)

#### Tools

* [Plugin for Visual Studio Code](https://marketplace.visualstudio.com/items?itemName=x1unix.gilbert) 