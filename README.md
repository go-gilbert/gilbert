<p align="center"><img src="docs/assets/gilbert.png" width="256"></p>

# Gilbert

> Build toolchain and task runner for Go

This project is currently in early development stage. Bug reports and pull requests are welcome.

## Features

**Gilbert** is task runner that aims to provide declarative way to define and run tasks like in other projects like _Gradle_, _Maven_ and etc.

All tasks are declared in *gilbert file* (`gilbert.yaml`). Example of the file you can find [here](https://github.com/x1unix/gilbert/blob/master/gilbert.yaml).

## Installation

`go get -u github.com/x1unix/gilbert`

## Usage

Run `gilbert init` to create a sample `gilbert.yaml` file with basic build task.

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
  - [ ] Scaffolding
  - [ ] Third-party plugins
    - [ ] Plugin support
    - [ ] Windows support
