+++
title = "Built-in actions"
description = "Information about built-in actions"
weight = 30
draft = false
toc = true
bref = "Gilbert contains a few core built-in actions. External actions are available through third-party plugins"
+++

<h3 class="section-head" id="build-action"><a href="#build-action">Build action</a></h3>
<p>
	Build action is abstraction over <code>go build</code> compile tool and simplifies build params pass.
	<br />
	This action can operate without configuration.
</p>
<h4>Configuration sample</h4>
```yaml
version: 1.0
tasks:
	build:
	- action: build
	  params:
	  	source: 'github.com/foo/bar' 		# default: current package
		buildMode: 'c-archive' 				# default: "default"
		outputPath: './build/foo.exe'		# default: project directory
        tags: 'foo bar baz'                 # set of build tags, separated by space
		params:
			stripDebugInfo: true			# removes debug info, default: false
			linkerFlags:					# custom linker flags, default: empty
			- '-X main.foo=bar'
		target:
			os: windows		# default: current OS
			arch: '386'		# default: current arch
		variables:			# default: empty
			'main.commit': '{% git log --format=%H -n 1 %}'	
```
<h4>Configuration params</h4>
<p>
	    <table>
        <tr>
            <th>Param name</th>
            <th>Type</th>
            <th>Description</th>
        </tr>
        <tr>
            <td>`source`</td>
            <td><i>string</i></td>
            <td>Package name to be built</td>
        </tr>
        <tr>
            <td>`buildMode`</td>
            <td><i>string</i></td>
            <td>Build mode, see `go help buildmode` for possible values</td>
        </tr>
        <tr>
            <td>`outputPath`</td>
            <td><i>string</i></td>
            <td>Artifact output path</td>
        </tr>
        <tr>
            <td>`tags`</td>
            <td><i>string</i></td>
            <td>List of Go build tags separated by space (e.g: `foo bar`)</td>
        </tr>
        <tr>
            <td>`params`</td>
            <td><i>object</i></td>
            <td>
                Additional params related to linker:
                <ul>
                    <li>`stripDebugInfo` - Remove all debug symbols</li>
                    <li>`linkerFlags` - Array of flags passed to linker</li>
                </ul>
            </td>
        </tr>
        <tr>
            <td>`target`</td>
            <td><i>object</i></td>
            <td>
                Defines build target:
                <ul>
                    <li>`os` - Target operating system (<i>default: current OS</i>)</li>
                    <li>`arch` - Target architecture (<i>default: current architecture</i>)</li>
                </ul>
            </td>
        </tr>
        <tr>
            <td>`variables`</td>
            <td><i>dict</i></td>
            <td>
                Key-value pair of variables to replace in executable by linker (`main.version` for example).<br />
                Can be useful to set application version or build commit.
            </td>
        </tr>
    </table>
</p>

<h3 class="section-head" id="shell-action"><a href="#shell-action">Shell action</a></h3>
<p>
	Shell action allows to execute shell commands. If command returns non-zero exit code, task will fail.
</p>
<h4>Configuration sample</h4>
```
version: 1.0
tasks:
  run_something:
  - action: shell
    params:
      command: 'scp root@localhost:/foo/bar ./bar'
      silent: false           # optional, default: false
      rawOutput: false        # optional, default: false
      shell: '/bin/bash'      # optional, default: /bin/sh or cmd.exe
      shellExecParam: '-c'    # optional, default: -c (or /c on windows for cmd.exe)
      workDir: '/tmp'         # optional, default: project directory
      env:                    # optional, default: use user's env vars
        LC_LANG: 'en_UTF-8'
```
<h4>Configuration params</h4>
<p>
	    <table>
        <tr>
            <th>Param name</th>
            <th>Type</th>
            <th>Description</th>
        </tr>
        <tr>
            <td class="param-required"><code>command</code></td>
            <td><i>string</i></td>
            <td>Command to run</td>
        </tr>
        <tr>
            <td><code>silent</code></td>
            <td><i>boolean</i></td>
            <td>Hide command output</td>
        </tr>
        <tr>
            <td><code>rawOutput</code></td>
            <td><i>boolean</i></td>
            <td>Do not decorate command output, can be useful if command output seems ugly</td>
        </tr>
        <tr>
            <td><code>shell</code></td>
            <td><i>string</i></td>
            <td>Shell executable, not recommended to change on <b>Windows</b></td>
        </tr>
        <tr>
            <td><code>shellExecParam</code></td>
            <td><i>string</i></td>
            <td>Shell command argument, not recommended to change</td>
        </tr>
        <tr>
            <td><code>workDir</code></td>
            <td><i>string</i></td>
            <td>Working directory</td>
        </tr>
        <tr>
            <td><code>env</code></td>
            <td><i>dict</i></td>
            <td>Custom environment variables</td>
        </tr>
    </table>
</p>
<h3 class="section-head" id="watch-action"><a href="#watch-action">Watch action</a></h3>
<p>
	Tracks file changes in specified path and restarts specified job on file/folder change.
</p>
<h4>Configuration sample</h4>
```
version: 1.0
tasks:
  watch:
  - action: watch
    params:
      path: './src/...'   # path to watch, required
      debounceTime: 300   # debounce time, optional
      ignore:
      - *.log             # list of entries to ignore, optional
      run:
        mixin: build-and-run-server # job or mixin to execute, similar to manifest job syntax. required.
```
<h4>Configuration params</h4>
<p>
	    <table>
        <tr>
            <th>Param name</th>
            <th>Type</th>
            <th>Description</th>
        </tr>
        <tr>
            <td class="param-required"><code>path</code></td>
            <td><i>string</i></td>
            <td>Path to track for changes. Use `/...` to track changes in all sub-directories.</td>
        </tr>
        <tr>
            <td class="param-required"><code>run</code></td>
            <td><i>object</i></td>
            <td>Job or mixin to run on change. See <a href="../schema/#tasks">Job definition</a> for more info</td>
        </tr>
        <tr>
            <td><code>debounceTime</code></td>
            <td><i>int</i></td>
            <td>period to postpone job execution until after wait milliseconds have elapsed since the last time it was invoked</td>
        </tr>
        <tr>
            <td><code>ignore</code></td>
            <td><i>[]string</i></td>
            <td>List of entries to ignore. All dotfiles are already included</td>
        </tr>
    </table>
</p>
<p>
  <span class="param-required"></span> - Required parameter<br />
</p>

<h3 class="section-head" id="cover-action"><a href="#cover-action">Cover action</a></h3>
<p>
    Runs package tests and checks package code coverage. Task fails if code coverage is below specified threshold.
</p>
<h4>Full configuration sample</h4>
```
version: 1.0
tasks:
  coverage:
  - action: cover
    params:
      threshold: 60.5       # minimal coverage percent
      reportCoverage: true  # show coverage report in output
      fullReport: false     # display coverage for each function
      showUncovered: false  # show list of packages without tests
      sort:
        by: 'coverage'      # sort report by package name or coverage
        desc: true          # sort ascending or descending
      packages:
      - ./controllers       # list of packages to cover
      - ./src/...
```
<h4>Configuration params</h4>
<p>
	    <table>
        <tr>
            <th>Param name</th>
            <th>Type</th>
            <th>Description</th>
        </tr>
        <tr>
            <td class="param-required"><code>packages</code></td>
            <td><i>[]string</i></td>
            <td>List of packages to check</td>
        </tr>
        <tr>
            <td class="param-required"><code>threshold</code></td>
            <td><i>double</i></td>
            <td>Minimal coverage percent</td>
        </tr>
        <tr>
            <td><code>reportCoverage</code></td>
            <td><i>boolean</i></td>
            <td>Display coverage summary</td>
        </tr>
        <tr>
            <td><code>fullReport</code></td>
            <td><i>boolean</i></td>
            <td>Display coverage for each function in package</td>
        </tr>
        <tr>
            <td><code>showUncovered</code></td>
            <td><i>boolean</i></td>
            <td>Display list of packages without tests</td>
        </tr>
        <tr>
            <td><code>sort</code></td>
            <td><i>object</i></td>
            <td>Coverage report sort</td>
        </tr>
    </table>
</p>
<p>
  <span class="param-required"></span> - Required parameter<br />
</p>

<h3 class="section-head" id="get-package-action"><a href="#get-package-action">Get-Package action</a></h3>
<p>
	Installs libraries using `go get` tool
</p>
<h4>Configuration sample</h4>
```
version: 1.0
tasks:
  watch:
  - action: get-package
    params:
      update: false       # force update, optional
      verbose: false      # debug output, optional
      downloadOnly: false # download without build, optional
      packages:
      - github.com/stretchr/testify
      - github.com/alecthomas/gometalinter
```
<h4>Configuration params</h4>
<p>
	    <table>
        <tr>
            <th>Param name</th>
            <th>Type</th>
            <th>Description</th>
        </tr>
        <tr>
            <td class="param-required"><code>packages</code></td>
            <td><i>[]string</i></td>
            <td>List of packages to install</td>
        </tr>
        <tr>
            <td><code>update</code></td>
            <td><i>boolean</i></td>
            <td>Force package update</td>
        </tr>
        <tr>
            <td><code>downloadOnly</code></td>
            <td><i>boolean</i></td>
            <td>Download libraries without compilation</td>
        </tr>
        <tr>
            <td><code>verbose</code></td>
            <td><i>boolean</i></td>
            <td>Debug output</td>
        </tr>
    </table>
</p>
<p>
  <span class="param-required"></span> - Required parameter<br />
</p>
<h3 class="section-head" id="third-party-actions"><a href="#third-party-actions">Third-party actions</a></h3>
<p>
    Third party actions could be added with custom plugins.
    See [plugins docs](../plugins) for more info.
</p>