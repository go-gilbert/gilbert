+++
title = "Built-in plugins"
description = "Information about plugins built-in Gilbert"
weight = 30
draft = false
toc = true
bref = "Gilbert contains a few core built-in plugins. External plugins functionality is work in progress"
+++

<h3 class="section-head" id="h-build-plugin"><a href="#h-build-plugin">Build plugin</a></h3>
<p>
	Build plugin is abstraction over <code>go build</code> compile tool and simplifies build params pass.
	<br />
	This plugin can operate without configuration.
</p>
<h4>Configuration sample</h4>
<pre class="code">
version: 1.0
tasks:
	build:
	- plugin: build
	  params:
	  	source: 'github.com/foo/bar' 		# default: current package
		buildMode: 'c-archive' 				# default: "default"
		outputPath: './build/foo.exe'		# default: project directory
		params:
			stripDebugInfo: true			# removes debug info, default: false
			linkerFlags:					# custom linker flags, default: empty
			- '-X main.foo=bar'
		target:
			os: windows		# default: current OS
			arch: '386'		# default: current arch
		variables:			# default: empty
			'main.commit': '{% git log --format=%H -n 1 %}'	
</pre>
<h4>Configuration params</h4>
<p>
	    <table>
        <tr>
            <th>Param name</th>
            <th>Required</th>
            <th>Description</th>
        </tr>
        <tr>
            <td><code>source</code></td>
            <td><i>false</i></td>
            <td>Package name to be built</td>
        </tr>
        <tr>
            <td><code>buildMode</code></td>
            <td><i>false</i></td>
            <td>Build mode, see <code>go help buildmode</code> for possible values</td>
        </tr>
        <tr>
            <td><code>outputPath</code></td>
            <td><i>false</i></td>
            <td>Artifact output path</td>
        </tr>
        <tr>
            <td><code>params</code></td>
            <td><i>false</i></td>
            <td>Additional params related to linker</td>
        </tr>
        <tr>
            <td><code>target</code></td>
            <td><i>false</i></td>
            <td>Defines build target. Uses values for <code>GOOS</code> and <code>GOARCH</code>.</td>
        </tr>
        <tr>
            <td><code>variables</code></td>
            <td><i>false</i></td>
            <td>Source variables that should be overwriten by linker</td>
        </tr>
    </table>
</p>

<h3 class="section-head" id="h-shell-plugin"><a href="#h-shell-plugin">Shell plugin</a></h3>
<p>
	Shell plugin allows to execute shell commands. If command returns non-zero exit code, task will fail.
</p>
<h4>Configuration sample</h4>
<pre class="code">
version: 1.0
tasks:
	run_something:
	- plugin: shell
	  params:
	  	command: 'scp root@localhost:/foo/bar ./bar'
		silent: false			# optional, default: false
		rawOutput: false		# optional, default: false
		shell: '/bin/bash'		# optional, default: /bin/sh or cmd.exe
		shellExecParam: '-c'	# optional, default: -c (or /c on windows for cmd.exe)
		workDir: '/tmp'			# optional, default: project directory
		env:					# optional, default: use user's env vars
			LC_LANG: 'en_UTF-8'
</pre>
<h4>Configuration params</h4>
<p>
	    <table>
        <tr>
            <th>Param name</th>
            <th>Required</th>
            <th>Description</th>
        </tr>
        <tr>
            <td><code>command</code></td>
            <td><i>true</i></td>
            <td>Command to run</td>
        </tr>
        <tr>
            <td><code>silent</code></td>
            <td><i>false</i></td>
            <td>Hide command output</td>
        </tr>
        <tr>
            <td><code>rawOutput</code></td>
            <td><i>false</i></td>
            <td>Do not decorate command output, can be useful if command output seems ugly</td>
        </tr>
        <tr>
            <td><code>shell</code></td>
            <td><i>false</i></td>
            <td>Shell executable, not recommended to change on <b>Windows</b></td>
        </tr>
        <tr>
            <td><code>shellExecParam</code></td>
            <td><i>false</i></td>
            <td>Shell command argument, not recommended to change</td>
        </tr>
        <tr>
            <td><code>workDir</code></td>
            <td><i>false</i></td>
            <td>Working directory</td>
        </tr>
        <tr>
            <td><code>env</code></td>
            <td><i>false</i></td>
            <td>Custom environment variables</td>
        </tr>
    </table>
</p>