// See: https://developer.hashicorp.com/terraform/language/expressions/version-constraints
version = 1

imports = [
  "./fs/module.hcl",
  "./semver/module.hcl",
  "./golangci-lint/mod.hcl",
  "file://./golangci-lint/mod.hcl",
  "github://username/foobar",
  "${project.work_dir}/semver/mod.hcl",
  path_join(project.work_dir, "semver", "mod2.hcl")
]

vars {
  pkgName = "app"
}

param "buildDate" "Build date" {
  default = time_now()
  type = date
}

param "configFile" "Path to app config file" {
  default = "./config.json"
  required = false
  description = "foobar"

  flag {
    name = "cfg"
    short = "c"
  }

  validate = [
    fileExists(value)
  ]
}

param "buildDir" "Output build directory" {
  default = path_join(project.work_dir, "target")
}

task "lint" {
  action "golangci-lint:run" {
    path = "./..."
    analyzeTests = false
  }
}

task "build:platform" "Build project for specific platform" {
  param "os" {
    required = true
  }

  param "arch" {
    required = true
  }

  param "version" {
    default = "${shell("git describe --abbrev=0 --tags")}-snapshot"
    validate = [
      regex_match("/(?:^v([0-9]+).([0-9]+).([0-9]+))(-([a-z]+)(.[0-9])?)?$/i", value),
      semver_match(value)
    ]
  }

  vars {
    outputDir = path_join(globals.buildDir, "${params.os}-${params.arch}")
    archiveName = "myApp_${params.os}-${params.arch}.${params.os == "windows" ? "zip" : "tar.gz"}"
  }

  action "fs:remove" {
    dir = vars.outputDir
    when = [
      file_exists(vars.outputDir)
    ]
  }

  action "go:build" {
    description = "Building application for ${params.os} ${params.arch}"
    package = "./cmd/..."
    outputFile = vars.outputDir
  }

  action "archive:compress" "Creating archive..." {
    id = "archive"
    #    keepResult = true
    #    type = 'tar.gz'
    output_file = path_join(vars.outputDir, vars.archiveName)
    input = [
      vars.outputDir
    ]
  }

  action "fs:write" "Write MD5 hash" {
    fileName = "${basename(jobs.archive.result.fileName)}.md5"
    data = md5sum(file_open(jobs.archive.result.fileName))
  }

  action "log:info" {
    message = "Build complete - ${vars.outputDir}"
  }
}

task "build" "Build Project" {
  matrix = [
    {
      os: "linux", arch: ["amd64", "386", "arm64", "arm"],
    },
    {
      os: "windows", arch: ["amd64", "386", "arm64", "arm"],
    },
    {
      os: "darwin", arch: ["amd64", "arm64"],
    },
  ]

  task "build:platform" {
    os = matrix.os
    arch = matrix.arch
  }
}

task "start" "Run with live reload" {
  action "fs:watch" {
    path = "./..."
    ignore = [
      "!*.go",
      "*_test.go"
    ]

    onChange {
      action "go:run" {
        package = "./cmd/app"
        args = "--config ${params.configFile}"
        async = true
        signals {
          start = signals.reload
        }
      }
    }
  }

}