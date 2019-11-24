# Gilbert 2

This document covers all problems that we faced and solutions we hope implement in Gilbert 2.

## Issues & Goals

### Configuration file (`gilbert.yaml`)

#### File format is not comfortable

**Description**

Current configuration file becomes very overloaded and unintuitive when contains
when became complicated (e.g. contains lot of mixins or tasks) [sample](gilbert.yaml).

Also, Yaml doesn't support `RawMessage` unmarshaling (unlike JSON) so `mapstructure` lib
should be used.

Yaml is human-friendly format, but not suitable for complicated contents.

**Solution**

Hashicorp's [HCL 2](https://github.com/hashicorp/hcl/tree/hcl2/) is a good candidate
for new file format. According to docs, it's a good trade-off between data format and
DSL. Kinda reminds me a groovy.

Package provides additional tools for parsing. That gives us a good potential for extensibility.

Preview version of **HCL** file can be found [here](gilbert.hcl).

### Plugins

#### Provide reliable plugin API

**Description**

Current plugin API is experimental and based on Go's `plugin` package.

The `plugin` package causes some issues with plugin usage and support:

- Plugin and Gilbert Go version should be the same (otherwise - error).
- Minor Go release also might break plugin work (e.g. can get `error` package version mismatch).
- Gilbert and all plugins should use the same versions of mutual packages (otherwise - error).
- MacOS and Linux only.

**Solution**

I've tried to write Go->CGo->Go bridge for Windows but unfortunately, plugin compiled as c-shared library
cannot call functions that use syscall underhood.

Also, `plugin` package itself has lot of limitations, so a different approach should be used.

The best solution might be to compile plugin as a separate executable and provide IPC mechanism.
A good candidate for IPC can be a named pipe or socket. Shared memory is slow and requires CGo, so not an option.

Sharing task execution context will be a big challenge.

### Actions

#### Migration action

**Description**

Add [golang-migrate](https://github.com/golang-migrate/migrate) wrapper for SQL migrations.


## Development Roadmap

* [ ] **Configuration file**
    - [ ] Design HCL-based configuration file schema (in progress)
    - [ ] Implement HCL config file parser (convert human-friendly data to task runner steps tree)
    - [ ] Integrate new config file codebase into runner
