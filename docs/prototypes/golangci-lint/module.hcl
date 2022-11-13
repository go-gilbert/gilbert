version = 2

module "golangci-lint" {
  vars {
    binName = "golangci-lint"
  }

  // parameters are resolved from import URL query parameters.
  // For example, this is from "./golangci-lint?version=1.0.0"
  param "version" {
    default = "latest"
  }

  install {
    skip = [
      command_exists(module.vars.binName) || module_command_exists(module.vars.binName)
    ]

#   "skip" is alternative to "when" block
#    when = [
#     !command_exists("golangci-lint")
#    ]

    dynamic "mixin" {
      content {
        name = array_includes(platform.os, ['windows', 'linux', 'darwin']) ? "download:${platform.os}" : fatal(
          "Plugin golangci-lint is not supported on ${platform.os}."
        )
      }
    }

    mixin "download" {
      when = [
        array_includes(platform.os, ['windows', 'linux', 'darwin']) ||
        error("Plugin golangci-lint is not supported on ${platform.os}.")
      ]
    }

#   Dynamic allows building dynamic blocks.
#   This is replacement of code below.
#   See: https://developer.hashicorp.com/terraform/language/expressions/dynamic-blocks
#
#    mixin "download:windows" {
#      when = [
#        platform.os == "windows"
#      ]
#    }
#
#    mixin "download:linux" {
#      when = [
#        platform.os == "linux"
#      ]
#    }
#
#    mixin "download:darwin" {
#      when = [
#        platform.os == "darwin"
#      ]
#    }
  }

  function "getDownloadUrl" {
    param "version" {}
    // See: https://github.com/google/cel-go
    expression = first(
      regex_filter(
        jsonpath(
          http_request("https://api.github.com/repos/golangci/golangci-lint/releases/${params.version}"),
          "$.assets[*]['browser_download_url']"
        ),
        "/${platform.os}-${platform.arch}/i"
      )
    ) || error("failed to find download link")
  }

  mixin "download" {
    vars {
      tmpdir = mktempdir()
    }

    defer {
        // Cleanup
        action "fs:remove" {
          path = vars.tempdir
        }
    }

    action "http:download" "Downloading golangci-lint..." {
      id = "download"
      keepResult = true
      url = getDownloadUrl(module.params.version)
      destination = vars.tmpdir
    }

    action "archive:extract" {
      id = "extract"
      keepResult = true
      fileName = jobs.download.result.fileName
      destination = basename(jobs.download.result.fileName)
    }

#    action "fs:copy" {
#      source = path_join(jobs.extract.result.destination, "golangci-lint.exe")
#      destination = path_join(module.binDir, "golangci-lint.exe")
#    }

    action "module:installBinaries" {
      source = path_join(jobs.extract.result.destination)
      name = vars.binName
      names = [
        vars.binName
      ]
    }
  }

  task "run" {
    param config {
      default = pathJoin(project.workDir, ".golangci.yml")
    }

    param "timeout" {
      type = "duration"
      default = "1m"
    }

    param "analyzeTests" {
      default = true
      flag = "analyze-tests"
    }

    param "outFormat" {
      default = "colored-line-number"
      flag = "out-format"
      validate = [
        arrayIncludes(
          value,
          "colored-line-number", "line-number", "json", "tab", "checkstyle", "code-climate", "html", "junit-xml", "github-actions"
        )
      ]
    }

    param "path" {
      default = "./..."
    }

    action "module:execBinary" {
      binary = vars.binName
      args = array_concat(params.path, mapParamsToArgs(params))
    }
  }
}