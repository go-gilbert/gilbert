<p align="center"><img src="docs/assets/gilbert.png" width="256"></p>
<p align="center">
  <a href="https://travis-ci.org/go-gilbert/gilbert"><img src="https://travis-ci.org/go-gilbert/gilbert.svg?branch=master" alt="Build Status"></img></a>
  <a href="https://ci.appveyor.com/project/go-gilbert/gilbert"><img src="https://ci.appveyor.com/api/projects/status/github/go-gilbert/gilbert?svg=true&branch=master&passingText=Windows%20-%20OK&failingText=Windows%20-%20failed&pendingText=Windows%20-%20pending" alt="Windows Build Status"></a>
  <a href="https://goreportcard.com/report/github.com/go-gilbert/gilbert"><img src="https://goreportcard.com/badge/github.com/go-gilbert/gilbert" /></a>
  <a href="https://opensource.org/licenses/mit-license"><img src="https://img.shields.io/badge/license-MIT-brightgreen.svg" /></a>
</p>

# Gilbert

> Build toolchain and task runner for Go

This project is currently in early development stage. Bug reports and pull requests are welcome.

## Features

**Gilbert** is task runner that aims to provide declarative way to define and run tasks like in other projects like _Gradle_, _Maven_ and etc.

All tasks are declared in *gilbert file* (`gilbert.yaml`). Example of the file you can find [here](https://github.com/go-gilbert/gilbert/blob/master/gilbert.yaml).

## Installation

All release binaries are available on the [releases](https://github.com/go-gilbert/gilbert/releases) page.

### Unix-based OS'es

This installation is supported for **Linux**, **FreeBSD** and **MacOS**

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

Run `gilbert init` to create a sample `gilbert.yaml` file in your project's directory with basic build task.

To run specific task, use `gilbert run [taskname]`

## Roadmap

- [x] Core 
  - [x] Task runner
  - [x] Logging formatting
  - [x] String and shell expressions
  - [x] Variables
- [ ] Built-in plugins
  - [x] Build
  - [x] Shell command eval
  - [ ] Tests
  - [ ] Package managers integration
- [ ] Advanced
  - [x] Scaffolding
  - [ ] Third-party plugins
    - [ ] Plugin support
    - [ ] Windows support
