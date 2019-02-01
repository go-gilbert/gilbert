# Gilbert

> Build toolchain and task runner for Go

This project is currently in early development stage. Bugs and pull requests are welcome.

## Roadmap

- [x] Build declaration
- [ ] Built-in features
  - [x] Compilation
  - [ ] Tests
  - [ ] Dependency installation
  - [ ] Linter integration
- [ ] Advanced
  - [ ] Manifest presets
  - [ ] Caching
  - [ ] Scaffolding
  - [ ] Third-party plugins
    - [ ] Plugin support
    - [ ] Windows support

## Features

**Gilbert** aims to provide declarative way to define and run tasks as it was implemented in other projects like _Gradle_, _Maven_ and etc.

All tasks are declared in *gilbert file* (`gilbert.yaml`). Example of the file you can find [here](https://github.com/x1unix/gilbert/blob/master/gilbert.yaml).

## Usage

To run specific task, use `gilbert run [taskname]`


