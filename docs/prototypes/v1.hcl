// See: https://developer.hashicorp.com/terraform/language/expressions/version-constraints
version = 2

imports = [
  "./semver/module.hcl",
  "./golangci-lint/module.hcl",
  "github://username/foobar"
]

vars {
  pkgName = "app"
}

param "configFile" "Path to app config file" {
  default = "./config.json"
  required = false

  flag {
    name = "cfg"
    short = "c"
  }

  validate = [
    fileExists(value)
  ]
}

param "buildDir" "Output build directory" {
  default = path_join(project.workDir, "target")
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
    default = "${shell('git describe --abbrev=0 --tags')}-snapshot"
    validate = [
      regex_match('/(?:^v([0-9]+).([0-9]+).([0-9]+))(-([a-z]+)(.[0-9])?)?$/i', value),
      semver_match(value)
    ]
  }

  vars {
    outputDir = path_join(globals.buildDir, "${params.os}-${params.arch}")
    archiveName = "myApp_${params.os}-${params.arch}.${params.os == 'windows' ? 'zip' : 'tar.gz'}"
  }

  action "fs:remove" {
    dir = vars.outputDir
    when = [
      file_exists(vars.outputDir)
    ]
  }

  action "go:build" "Building application for ${params.os} ${params.arch}" {
    package = "./cmd/..."
    outputFile = vars.outputDir
  }

  action "archive:compress" "Creating archive..." {
    id = 'archive'
#    keepResult = true
#    type = 'tar.gz'
    outputFile = path_join(vars.outputDir, vars.archiveName)
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
      os: 'linux', arch: ['amd64', '386', 'arm64', 'arm'],
    },
    {
      os: 'windows', arch: ['amd64', '386', 'arm64', 'arm'],
    },
    {
      os: 'darwin', arch: ['amd64', 'arm64'],
    },
  ]

  task "build:platform" {
    os = matrix.os
    arch = matrix.arch
  }
}

task "start" "Run with live reload" {
  signal "reload" {}

  action "fs:watch" {
    path = "./..."
    notifySignal = signals.reload
  }

  action "go:run" {
    package = "./cmd/app"
    args = "--config ${params.configFile}"
    async = true
    signals {
      start = signals.reload
    }
  }

}