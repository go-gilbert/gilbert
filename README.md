# Guru

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

**guru** aims to provide declarative way to define and run tasks as it was implemented in other projects like _Gradle_, _Maven_ and etc.

All tasks are declared in *guru file* (`guru.yaml`). Example of the file you can find [here](https://github.com/x1unix/guru/blob/master/guru.example.yaml).

## Usage

To run specific task, use `guru run [taskname]`


