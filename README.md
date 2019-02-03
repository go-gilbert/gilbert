<p align="center"><img src="docs/assets/gilbert.png" width="256"></p>
<p align="center">
  <a href="https://travis-ci.org/x1unix/gilbert"><img src="https://travis-ci.org/x1unix/gilbert.svg?branch=master" alt="Build Status"></img></a>
  <a href="https://ci.appveyor.com/project/x1unix/gilbert"><img src="https://ci.appveyor.com/api/projects/status/github/x1unix/gilbert?svg=true&branch=master&passingText=Windows%20-%20OK&failingText=Windows%20-%20failed&pendingText=Windows%20-%20pending" alt="Windows Build Status"></a>
  <a href="https://goreportcard.com/report/github.com/x1unix/gilbert"><img src="https://goreportcard.com/badge/github.com/x1unix/gilbert" /></a>
  <a href="https://opensource.org/licenses/gpl-license"><img src="https://img.shields.io/badge/license-GPL-brightgreen.svg" /></a>
</p>

# Gilbert

> Build toolchain and task runner for Go

This project is currently in early development stage. Bug reports and pull requests are welcome.

## Features

**Gilbert** is task runner that aims to provide declarative way to define and run tasks like in other projects like _Gradle_, _Maven_ and etc.

All tasks are declared in *gilbert file* (`gilbert.yaml`). Example of the file you can find [here](https://github.com/x1unix/gilbert/blob/master/gilbert.yaml).

## Installation

Release binaries are available on the [releases](https://github.com/x1unix/gilbert/releases) page.

Also, you can install a development version using `go get`:

```
go get -u github.com/x1unix/gilbert
```

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
