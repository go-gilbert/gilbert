+++
title = "gilbert.yaml"
description = "Creating tasks and pipelines in gilbert.yaml file"
weight = 20
draft = false
toc = true
bref = "This article covers all information about gilbert.yaml syntax and step-by-step task definition guide"
+++

<h3 class="section-head" id="h-intro"><a href="#h-intro">Manifest file</a></h3>
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
<pre class="code">
version: 1.0
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
<p>
    Here are 3 main sections:
    <ul>
        <li>
            <code>version</code> is file schema version
        </li>
        <li>
            <code>vars</code> is a list of global variables that available for all tasks and jobs
        </li>
        <li>
            <code>tasks</code> contains a list of tasks. Task can be called by <code>gilbert run task_name</code>
        </li>
    </ul>
</p>

<h3 class="section-head" id="h-variables"><a href="#h-variables">Variables</a></h3>
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
    All variables contain only string values and support templating
</p>

<h3 class="section-head" id="h-templates"><a href="#h-templates">String templates</a></h3>
<p>
    All variable values and some other params can contain not only static value, but also template value.
    <br />
    String template can contain a value of any variable (<code>{{ var_name }}</code>) or a value of some shell command (<code>{% whoami %}</code>) or both.
</p>
<p>
    <b>Example:</b><br />
    <pre class="code">
version: 1.0
vars:
    foo: "{% go version %} is installed on {{ GOROOT }}"
    </pre>
</p>
<p>
    The value of variable <i>foo</i> will be:<br />
    <code>go version go1.10.1 linux/amd64 is installed on /usr/local/go</code>
</p>
<p>
    <i>Tip - environment variables are by default included in global scope</i>
</p>

<h3 class="section-head" id="h-tasks"><a href="#h-tasks">Tasks</a></h3>
<p>
    Each task is located in <code>tasks</code> section and contains from a sequence of
    jobs that should be ran when task was called.
</p>
<h4>Job definiton</h4>
<p>
    Each job should be handled by specific plugin, contains it's own variables and arguments for handler plugin.
    <br />
    Here is a full example of task with a few jobs. Most of params are <i>optional</i>.
</p>
<p>
    <pre class="code">
tasks:
    build_project:
    - plugin: build                     # name of plugin that will handle this step, required!
      description: "build the project"  # step description, optional
      delay: 500                        # delay before step start in milliseconds, optional
      vars:
        commit_id: "{% git log --format=%H -n 1 %}"     # Variables for current step, optional
        foo: "bar"
      params:                                           # Arguments for plugin.
        variables:                                      # Those values are specific
            'main.version': '{{ commit_id }}'           # to each plugin.
            'main.stable': 'true'
    # Additional task:
    - if: "uname -a"    # Condition for step run, contains a shell command, optional
      plugin: shell
      params:
        command: 'echo "I`m running on Unix machine"'
    </pre>
</p>
<p>
    <h4>Job fields</h4>
    <table>
        <tr>
            <th>Param name</th>
            <th>Required</th>
            <th>Description</th>
        </tr>
        <tr>
            <td><code>if</code></td>
            <td><i>false</i></td>
            <td>Contains shell command. If command returns non-zero exit code, step will be skipped</td>
        </tr>
        <tr>
            <td><code>plugin</code></td>
            <td><i>true</i></td>
            <td>Name of the plugin that will handle this step. See <a href="../built-in-plugins">built-in plugins</a> for more info</td>
        </tr>
        <tr>
            <td><code>description</code></td>
            <td><i>false</i></td>
            <td>Contains step description and makes your job run status more informative</td>
        </tr>
        <tr>
            <td><code>delay</code></td>
            <td><i>false</i></td>
            <td>Delay before step start in milliseconds</td>
        </tr>
        <tr>
            <td><code>vars</code></td>
            <td><i>false</i></td>
            <td>List of local variables for the step, work the same as global <code>vars</code> section.</td>
        </tr>
        <tr>
            <td><code>params</code></td>
            <td><i>maybe</i></td>
            <td>Contains arguments for the plugin. See docs for specific plugin for more info</td>
        </tr>
    </table>
</p>
