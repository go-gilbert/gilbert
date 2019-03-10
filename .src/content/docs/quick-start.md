+++
title = "Quick Start"
description = "Up and running in under a minute"
weight = 10
draft = false
toc = true
bref = "This article covers installation of Gilbert on your local machine"
+++

<h3 class="section-head" id="h-installation">
    <a href="#h-installation">Installation</a>
</h3>
<p>
    You can download release binaries from <a href="https://github.com/x1unix/gilbert/releases" target="_blank">releases page</a> or grab the latest version using <code>go get</code>:
</p>
<pre class="code">go get -u github.com/x1unix/gilbert</pre>
<p>
    This command will install <b>Gilbert</b> tool into <code>$GOPATH/bin</code>
</p>


<h3 class="section-head" id="h-project-integration">
    <a href="#h-project-integration">Project integration</a>
</h3>
<p>
    **Gilbert** uses <code>gilbert.yaml</code> file to store list of tasks to run in project folder.
</p>
<p>
    To generate a sample <code>gilbert.yaml</code> file, navigate to your project directory in terminal and run <code>gilbert init</code> command:
</p>
<pre class="code">
    $ cd $GOPATH/src/github.com/user/myproject
    $ gilbert init
</pre>
<p>
    This command will generate a sample file with <code>build</code> and <code>clean</code> tasks:
</p>
<pre class="code">
version: "1.0"
vars:
  appVersion: 1.0.0
tasks:
  build:
  - description: Build project
    plugin: build
  clean:
  - if: file ./vendor
    description: Remove vendor files
    plugin: shell
    params:
      command: rm -rf ./vendor
</pre>
<h3 class="section-head" id="h-running-tasks">
    <a href="#h-running-tasks">Running tasks</a>
</h3>
<p>
    To run a task declared in `gilbert.yaml`, use `gilbert run` command.
</p>
<p>
    <b>Example:</b>
    <pre class="code">
gilbert run build
    </pre>
</p>
