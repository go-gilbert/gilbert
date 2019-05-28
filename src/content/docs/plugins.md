+++
title = "Plugins"
description = "Extend functionality with third-party plugins"
weight = 40
draft = false
toc = true
bref = "Usage of third-party plugins to use custom actions"
+++

{{<doc-section id="plugin-import" label="Import a plugin" >}}
Before using a plugin, it should be imported into your `gilbert.yaml` file.
Each import declaration should be in URL format

Plugin will be download automatically at first start and you will be able to use all actions that it exports.

```yaml
plugins:
  - github://github.com/go-gilbert/gilbert-plugin-example # import URL

tasks:
  hello-world:  # each plugin action should be in format 'plugin-name:action-name'
    - action: 'example-plugin:hello-world'
      params:
        message: 'hello world'
```

{{<doc-section id="import-sources" label="Import sources" >}}

Each plugin import URL starts with import handler as schema (e.g.: `github://`).
There are a few supported import sources:

#### Local file

Import plugin locally by file path.

```yaml
plugins:
  - file:///home/root/path/to/plugin.so
```

#### Web

Downloads plugin file from specified URL. Supported schemas are `http` and `https`.

```yaml
plugins:
  - http://example.com/storage/my_plugin.so
  - https://example.com/storage/my_plugin2.so
```

#### Local package

Builds local Go package as Gilbert plugin

```yaml
plugins:
  - go://./mypkg
  - go://{{ GOPATH }}/src/github.com/user/package?rebuild=true
```

**Optional URL parameters:**

* `rebuild` - rebuild local package each time. Useful for local plugin development.

See [plugin development docs](../plugin-development/) for more information.

#### GitHub

Plugins that are hosted on GitHub, can be downloaded by using special `github` handler.
Handler finds specified repo and downloads latest or specified plugin release.

Plugin artifact should be present at repo's **Releases** page.
See [GitHub publishing](../plugin-development/#plugin-deployment) for more info.

GitHub Enterprize and token auth are also supported.


```yaml
plugins:
  - github://github.com/owner/repo_name?version=v1.0.0&token=AUTH_TOKEN
```

**Optional URL parameters:**

* `version` - Release tag to download (default is `latest`)
* `token` - Your personal GitHub auth token

##### GitHub Enterprise

```yaml
plugins:
  - github://company.domain.com:8888/custom_path/owner/repo_name?version=v1.0.0&token=AUTH_TOKEN
```

To use custom GitHub host, just replace `github.com` to your GitHub Enterprise instance path.

Path can contain hostname, port and path.

**Optional URL parameters:**

* `version` - Release tag to download (default is `latest`)
* `token` - Your personal GitHub auth token
* `protocol` - Protocol to use (`http` or `https`). `https` is default value.

{{<doc-section id="explore-plugins" label="Explore plugins" >}}

You can explore third-party plugins by searching on GitHub by <code>[gilbert-plugin](https://github.com/topics/gilbert-plugin)</code> topic.

Also you can explore our [plugin development docs](../plugin-development) and create a plugin on your own.