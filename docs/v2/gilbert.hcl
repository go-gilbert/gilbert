version = 2

imports = [
  "./docs/common.yml"
]

plugins = [
  "go://./docs/actions"
]

vars {
  appVersion = "1.0.0"
  buildDir = "${PROJECT}/build"
  libDir = "${buildDir}/lib"
  serverDir = "./server"
  watcherAddr = "localhost:4800"
  libs = "${split(shell("ls -1 ./sources | xargs -0 -n 1 basename"), "\n")}"
  packages = [
      "./server",
      "./sources/..."
  ]
}

mixin "rebuild" {
  task "build" {}
  action "live-reload:trigger" {
      address = "${watcherAddr}"
  }

  task "start" {}
}

task "build" {
  task "clean" {}
  task "copy-assets" {}
  task "build-libs" {}
  action "go:build" "build server" {
    source = "${serverDir}"
    outputPath = "${buildDir}/server"

    replace "main.version" {
        value = "${appVersion}"
    }

    replace "main.commit" {
      value = "${shell("git log --format=%H -n 1")}"
    }
  }
}

// Dynamic blocks - under discussion
task "build:libs" {
  // 1. Templates variant (kinda sucks)

  /** range $libs */
  mixin "build-lib" {
    vars {
      name = "/**.*/"
    }
  }
  /** endfor */

  // 2. Dynamic block extension (see: https://github.com/hashicorp/hcl/tree/hcl2/ext/dynblock)
  // (less sucks)
  dynamic "mixin" {
    for_each = vars.libs
    iterator = lib
    content {
      vars {
        name = "${lib}"
      }
    }
  }
}

task "cover" {
  action "go:cover" {
    threshold = 40
    reportCoverage = 40
    packages = vars.packages
  }
}

task "watch" {
  action "live-reload:start-server" {
    async = true
    timeout = 1500
    address = vars.watcher_addr
  }

  action "watch" {
    path = "./server/..."
    run {
        mixin "rebuild" {}
    }
  }
}

task "copy-assets" {
  if "exists('${buildDir}')" {
    action "shell" "create build directory" {
      command = "mkdir ${buildDir}"
    }
  }

  action "file:copy" "copy config file" {
    from = "${serverDir}/config.json"
    to = "${buildDir}/config.json"
  }
}

task "start" {
  if "exists('${buildDir}')" {
    action "start" {
      command = "${buildDir}"
      args = []
      workDir = "${buildDir}"
    }
  }
}

task "clean" {
  if "exists('$buildDir')" {
    action "dir:remove" "clean build directory" {
      dir = "${buildDir}"
    }
  }
}
