imports = [
  "./common.hcl",
  "foo" + "bar",
  plugin("github://github.com/go-gilbert/gilbert-plugin-example"),
]

project = {
  location = "."
}

app_version   = "1.0.0"
build_dir     = "${project.location}/build"
lib_dir       = "${build_dir}/lib"
server_dir    = "./server"
watcher_addr  = "localhost:4800"

//pkg_name = go.mod.module


libs = split(shell("ls -1 ./sources | xargs -0 -n 1 basename"), "\n")
packages = [
  "./server",
  "./resources/..."
]

mixin "rebuild" {
  task "go:build" {
    package = "github.com/x1unix/go-gilbert"
  }
  action "live-reload:trigger" {
    address = "${watcher_addr}"
  }
  task "start" {
    params = {
      server_bin = "${build_dir}/server"
    }
  }
}

/*** $libs := shell "ls -1 ./sources | xargs -0 -n 1 basename" | split "\n" ***/
/*** $packages := slice "./server" "./sources/..." ***/
task "watch" "start server and watch for changes" {
  action "live-reload:start-server" "starting live-reload server" {
    async = true
    address = "${watcher_addr}"
    timeout = 1500
  }

  action "watch" {
    path = "./server/..."
    mixin "rebuild" {}
  }
}

task "start" "start web server" {
  param "server_bin" "server binary path" {
    // required = true
    type = "string"
    default = "${build_dir}/server"
  }

  action "exec" {
    require = file_exists("${server_bin}")
    program = "${server_bin}"
    work_dir = "${build_dir}"
  }
}

task "copy-assets" "copies assets to build dir" {
  action "shell" "create build directory" {
    require = file_exists("${build_dir}")
    command = "mkdir ${build_dir}"
  }

  action "file:copy" "copy config file" {
    from = "${server_dir}/config.json"
    to = "${build_dir}/config.json"
  }

  action "dir:copy" "copy assets" {
    from = "${server_dir}/public"
    to = "${public_dir}"
  }
}