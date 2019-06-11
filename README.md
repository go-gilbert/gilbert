<p align="center"><img src="docs/assets/gilbert.png" width="256"></p>
<p align="center">
  <a href="https://travis-ci.org/go-gilbert/gilbert"><img src="https://travis-ci.org/go-gilbert/gilbert.svg?branch=master" alt="Build Status"></img></a>
  <a href="https://ci.appveyor.com/project/x1unix/gilbert"><img src="https://ci.appveyor.com/api/projects/status/github/go-gilbert/gilbert?svg=true&branch=master&passingText=Windows%20-%20OK&failingText=Windows%20-%20failed&pendingText=Windows%20-%20pending" alt="Windows Build Status"></a>
  <a href="https://goreportcard.com/report/github.com/go-gilbert/gilbert"><img src="https://goreportcard.com/badge/github.com/go-gilbert/gilbert" /></a>
  <a href="https://opensource.org/licenses/mit-license"><img src="https://img.shields.io/badge/license-MIT-brightgreen.svg" /></a>
</p>

# Gilbert

> Build toolchain and task runner for Go

## Features

**Gilbert** is task runner that aims to provide declarative way to define and run tasks like in other projects like _Gradle_, _Maven_ and etc.

All tasks are declared in *gilbert file* (`gilbert.yaml`). Example of the file you can find [here](https://github.com/go-gilbert/gilbert/blob/master/gilbert.yaml).

**Full list of features:**

* [Tasks](https://go-gilbert.github.io/docs/syntax/#tasks)
  - Simple job declaration
  - Rollback and graceful shutdown
  - Evaluation conditions
  - Async and parallel jobs
  - Job timeout and deadline
  - Job and [manifest templates](https://go-gilbert.github.io/docs/syntax/#mixins)
  - [Variables](https://go-gilbert.github.io/docs/syntax/#variables) and [inline expressions](https://go-gilbert.github.io/docs/syntax/#h-templates)
* [Actions](https://go-gilbert.github.io/docs/actions/)
  - Built-in most necessary actions
    - Track file changes and re-run task on change
    - Check project test coverage with specified threshold
    - Build project
  - Plugins for custom actions
    - Get plugins from [GitHub](https://go-gilbert.github.io/docs/plugin-development/) or other sources
    - Simple [Plugin API](https://go-gilbert.github.io/docs/plugin-development/)

Read [documentation](https://go-gilbert.github.io/docs/) for more information.

## Installation

All release binaries are available on the [releases](https://github.com/go-gilbert/gilbert/releases) page.

### Linux, macOS and FreeBSD

```bash
curl https://raw.githubusercontent.com/go-gilbert/gilbert/master/install.sh | sh
```

### Windows

**Powershell**

```powershell
Invoke-Expression (Invoke-Webrequest 'https://raw.githubusercontent.com/go-gilbert/gilbert/master/install.ps1' -UseBasicParsing).Content
```

**Note**: You should run `Set-ExecutionPolicy Bypass` in PowerShell to be able to execute installation script.

## Usage

Please check out [quick start](https://go-gilbert.github.io/docs/quick-start/) guide.

### Tools

* [Plugin for Visual Studio Code](https://marketplace.visualstudio.com/items?itemName=x1unix.gilbert) 
