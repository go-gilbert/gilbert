+++
title = "Syntax"
description = "Creating tasks and pipelines in gilbert.yaml file"
weight = 20
draft = false
toc = true
bref = "This article covers all information about gilbert.yaml syntax and step-by-step task definition guide"
+++

<h3 class="section-head" id="intro"><a href="#intro">Manifest file</a></h3>
<p>
    <code>gilbert.yaml</code> it's a <a href="https://en.wikipedia.org/wiki/YAML" target="_blank">yaml</a> file that contains all information about your tasks and tells <b>Gilbert</b> what to do when you try to run specific task.
</p>
<p>
    This file should be located at the root folder of your codebase, or at least at the same location where you call <code>gilbert</code> command.
</p>
<p>
    To create a sample file, use <code>gilbert init</code> command.
</p>

<h3 class="section-head" id="h-structure"><a href="#h-strucure">Structure</a></h3>
<p>
    <code>gilbert.yaml</code> consists of tasks, and tasks include a list of jobs to be done and rules.
</p>
<p>
    Here is an annotated example of <code>gilbert.yaml</code> file.
</p>
```yaml
version: 1.0
imports:
  - ./misc/mixins.yaml
vars:
  appVersion: 1.0.0
tasks:
  build:
  - description: Build project
    action: build
  clean:
  - if: file ./vendor
    description: Remove vendor files
    action: shell
    params:
      command: rm -rf ./vendor
```
<p>
    <b>Sections:</b>
    <ul>
        <li>
            <code>version</code> - file schema version
        </li>
        <li>
          <code>imports</code> - list of files to import. Imported files should share the same syntax as <code>gilbert.yaml</code> file.
        </li>
        <li>
            <code>vars</code> - list of global variables that available for all tasks and jobs
        </li>
        <li>
            <code>mixins</code> - mixins. See <a href="#h-mixins">mixins</a> for more info
        </li>
        <li>
            <code>tasks</code> - contains a list of tasks. Task can be called by <code>gilbert run task_name</code>
        </li>
    </ul>
</p>

<h3 class="section-head" id="variables"><a href="#variables">Variables</a></h3>
<p>
    *Gilbert* allows to keep variables in the manifest file and use them in tasks and jobs.
</p>
<p>
    There are 2 type of variables:
    <ul>
        <li>
            <b>Global</b> - defined in <code>vars</code> section in root. Available everythere.
        </li>
        <li>
            <b>Local</b> - defined in specific job and visible only in scope of job.
        </li>
    </ul>
</p>
<p>
  <b>Tip:</b> variable values could be set or changed using <a href="../commands/#run-task">command flags</a>
</p>

<h4>Prefefined variables</h4>
<p>
  By default, there are a few predefined variables in global scope:

  - `PROJECT` - Path to the folder where `gilbert.yaml` is located.
  - `BUILD` - Alias to `${PROJECT}/build`, can be useful as default build output directory.
  - `GOPATH` - Go path environment variable
</p>

<h3 class="section-head" id="h-templates"><a href="#h-templates">String templates</a></h3>
<p>
    All variable values and some other params can contain not only static value, but also template expression.
    <br />
    String template can contain a value of any variable (`{{ var_name }}`) or a value of some shell command (`{% whoami %}`) or both.
</p>
<p>
    <b>Example:</b><br />
    <pre><code class="code">
version: 1.0
vars:
    foo: "{% go version %} is installed on {{ GOROOT }}"
    </code></pre>
</p>
<p>
    The value of variable <i>foo</i> will be:<br />
    <code>go version go1.10.1 linux/amd64 is installed on /usr/local/go</code>
</p>

<h3 class="section-head" id="tasks"><a href="#tasks">Tasks</a></h3>
<p>
    Each task is located in <code>tasks</code> section and contains from a sequence of
    jobs that should be ran when task was called.
</p>
<h4>Job definiton</h4>
<p>
    Each job should, contains action to execure, variables and action arguments.
    <br />
    Here is a full example of task with a few jobs. Most of parameters are <i>optional</i>.
</p>
<p>
```yaml
tasks:
    build_project:
    - action: build                     # name of action to perform, required!
      description: "build the project"  # step description, optional
      delay: 500                        # delay before step start in milliseconds, optional
      vars:
        commit: "{% git log --format=%H -n 1 %}"     # Variables for current step, optional
        foo: "bar"
      params:                                           # Arguments for action.
        variables:                                      # Those values are specific
            'main.version': "{{ commit }}"              # to each action.
            'main.stable': 'true'
    # Additional task:
    - if: 'uname -a'    # Condition for step run, contains a shell command, optional
      action: shell
      params:
        command: 'echo I am running on Unix machine'
```
</p>
<p>
    <h4>Job fields</h4>
    <table>
        <tr>
            <th>Param name</th>
            <th>Type</th>
            <th>Description</th>
        </tr>
        <tr>
            <td>`if`</td>
            <td><i>boolean</i></td>
            <td>Contains conditional shell command. If command returns non-zero exit code, step will be skipped</td>
        </tr>
        <tr>
            <td>`description`</td>
            <td><i>string</i></td>
            <td>Contains step description and makes your job run status more informative</td>
        </tr>
        <tr>
            <td class="param-required">`action`</td>
            <td><i>string</i></td>
            <td>Name of action to execute. See <a href="../actions">built-in actions</a> for more info</td>
        </tr>
        <tr>
            <td class="param-required">`mixin`</td>
            <td><i>string</i></td>
            <td>Name of the mixin to be called. <b>Cannot be together</b> with `action` in the same job.</td>
        </tr>
        <tr>
            <td>`async`</td>
            <td><i>boolean</i></td>
            <td>Run job asynchronously. Useful for executing programs like _web-servers_, etc.</td>
        </tr>
        <tr>
            <td>`delay`</td>
            <td><i>int</i></td>
            <td>Delay before step start in milliseconds</td>
        </tr>
        <tr>
            <td>`deadline`</td>
            <td><i>int</i></td>
            <td>Job execution deadline in milliseconds</td>
        </tr>
        <tr>
            <td>`vars`</td>
            <td><i>dict</i></td>
            <td>List of local variables for the step, work the same as global <code>vars</code> section.</td>
        </tr>
        <tr>
            <td class="param-optional">`params`</td>
            <td><i>dict</i></td>
            <td>Contains arguments for the action. See action docs for more info</td>
        </tr>
    </table>
</p>
<p>
  <span class="param-required"></span> - Required parameter<br />
  <span class="param-optional"></span> - Optional parameter but depends on action<br />
</p>
<h3 class="section-head" id="mixins"><a href="#mixins">Mixins</a></h3>
<p>
    Mixin is a set of jobs that can be included into task and used to reduce boilerplate code in `gilbert.yaml` file.<br />
    Also, one of the biggest differences that most of values can contain template expressions.  
</p>
<h4>Declaration</h4>
<p>
    Each mixin should be declared in `mixins` section and have the same syntax as regular jobs in `tasks` section:
</p>
```yaml
  version: 1.0
  mixins:
    hello-world:
      - action: shell
        params:
          command: 'echo "hello world"'
      - action: build
```
<h4>Calling a mixin</h4>
<p>
  Mixins are called by tasks and use job variables as parameters.<br />
  The same mixin can be called several times with different parameters in the same job.
</p>
<p>
  <b>Example:</b>
```yaml
version: 1.0
mixins:
  platform-build:
  - action: build
    description: 'build for {{os}} {{arch}}'
      vars:
        extension: '' # variable default value
      params:
        outputPath: '{{buildDir}}/myproject_{{os}}-{{arch}}{{extension}}'
        target:
          os: '{{os}}'
          arch: '{{arch}}'
  - if: 'type md5sum'
    description: 'generate checksum for {{buildDir}}/myproject_{{os}}-{{arch}}{{extension}}'
    action: shell
    vars:
      fileName: 'myproject_{{os}}-{{arch}}{{extension}}'
    params:
      workDir: '{{buildDir}}'
      command: 'md5sum {{fileName}} > {{fileName}}.md5'

tasks:
  release:
  - mixin: platform-build
    vars:
      os: windows
      arch: amd64
      ext: .exe
  - mixin: platform-build
    if: '[ $(uname -s) == "Darwin" ]'
    vars:
      os: darwin
      arch: amd64
```
</p>
<p>
  In example above, task `release` calls mixin `platform-release` and passes variables `os`, `arch` and `ext` to the mixin.
</p>
<h3 class="section-head" id="advanced-example"><a href="#advanced-example">Advanced examples</a></h3>
<p>
  You can find a good use-case example in <a href="https://github.com/go-gilbert/demo-go-plugins" target="_blank">this demo project</a>.<br />
  That repo demonstrates usage of mixins and a few built-in actions for a real-world web-server example.
</p>
<p><br /></p>
