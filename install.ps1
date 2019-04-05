# gilbert installer
# usage: (in powershell)
#  Invoke-Expression (Invoke-Webrequest <my location> -UseBasicParsing).Content

param (
    [string]$version = "latest",
    [bool]$quiet = $false
)

& {
    $global:PKG_URL="github.com/x1unix/gilbert"
    $global:ISSUE_URL="https://$global:PKG_URL/issues"

    Function Panic($msg){
        Write-Err "ERROR: $msg`n"
        Write-Err "`nIf you feel that this issue is an installation script failure,`nfeel free to create an issue here: $global:ISSUE_URL`n"
        Throw "Installation Failed"
    }

    Function Write-Err($msg) {
        Write-Host "$msg" -BackgroundColor Black -ForegroundColor Red
    }

    Function Write-Success($msg) {
        Write-Host $msg -ForegroundColor Green
    }

    Function ToolInstalled($name) {
        [bool](Get-Command -Name $name -ErrorAction SilentlyContinue)
    }

    Function CheckGoEnv {
        param([string] $Name)
        if (-not (Test-Path env:$Name)) {
            Write-Debug "$Name is not defined" -Debug
            $val = (go env $Name)
            [Environment]::SetEnvironmentVariable($Name, $val)
        }
    }

    Function DefaultInstallationPath {
        $pf = ([Environment]::GetFolderPath([System.Environment+SpecialFolder]::ProgramFiles))
        return Join-Path -Path $pf -ChildPath "gilbert"
    }

    Function CheckEnvironment {
        CheckGoEnv -Name GOROOT
        CheckGoEnv -Name GOPATH
    }

    Function TempDownloadPath {
        $tempDir = [System.IO.Path]::GetTempPath()
        return (Join-Path -Path "$tempDir" -ChildPath "gilbert.exe")
    }

    Function GetDownloadURL {
        if ($ENV:PROCESSOR_ARCHITECTURE -eq "AMD64") {
            $arch = "amd64"
        } else {
            $arch = "386"
        }

        return "https://$global:PKG_URL/releases/latest/download/gilbert_windows-$arch.exe"
    }

    Function AskBuildPrompt {
        Write-Output "`nAlthough binary download was failed, installation script can try to build Gilbert on your machine"
        $reply = Read-Host -Prompt "Continue?[y/n]"
        if ( $reply -match "[yY]" ) {
            return $true
        }

        return $false
    }

    Function TryBuild {
        if (! (ToolInstalled "git")) {
            Throw "Git is not installed"
        }

        if (! (ToolInstalled "dep")) {
            Throw "Go Dep is not installed"
        }

        Write-Host " - Downloading package '$global:PKG_URL'..."
        go get -d $global:PKG_URL
        if ($LastExitCode -ne 0) {
            Throw "failed to get '$global:PKG_URL' (error $LastExitCode)"
        }

        Write-Host " - Installing dependencies..."
        $buildDir = "$env:GOPATH\src\$global:PKG_URL"

        $p = Start-Process "dep" -WorkingDirectory "$buildDir" -ArgumentList "ensure"
        if ($p.ExitCode -ne 0) {
            Throw "failed to get package dependencies (error $($p.ExitCode))"
        }

        Write-Host " - Building Gilbert"
        go install $global:PKG_URL
        if ($LastExitCode -ne 0) {
            Throw "failed to build Gilbert (error $LastExitCode)"
        }

        Write-Success "Gilbert successfully built and installed to '$env:GOPATH\\bin'"
    }

    Function Main {
        $HasGo = (ToolInstalled "go")
        if ($HasGo) {
            $InstallationPath = (Join-Path -Path $env:GOPATH -ChildPath "bin")
        } else {
            $InstallationPath = (DefaultInstallationPath)
            Write-Warning "Go installation not found"
        }

        $outFile = (TempDownloadPath)
        $downloadUrl = (GetDownloadUrl)
        try {
            [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
            Write-Host "Downloading from '$downloadUrl' ..."
            Invoke-WebRequest $downloadUrl -OutFile "$outFile"

            Write-Debug "Download finished ($outFile)"
            Write-Debug "Installation directory: '$InstallationPath'"
            if (!(Test-Path -Path "$InstallationPath")) {
                Write-Debug "Installation directory not exists, creating a new one..."
                try {
                    New-Item -Force -ItemType directory -Path "$InstallationPath"
                } catch {
                    Panic "Failed to create installation path '$InstallationPath': $_"
                }
            }

            try {
                Write-Debug "Moving downloaded file to destination"
                Move-Item -Path "$outFile" -Destination $InstallationPath -Force
            } catch {
                Panic "Failed to copy file to destination path '$InstallationPath': $_"
            }

            Write-Success "Gilbert successfully installed to '$InstallationPath'"
        } catch {
            Write-Debug "Dowload failed: $_"
            # If download failed, but user has Go installation
            # prompt build option prompt
            if ($HasGo) {
                Write-Err "Failed to download Gilbert binary: $_"
                if ($quiet -Or (AskBuildPrompt)) {
                    try {
                        TryBuild
                    } catch {
                        Panic "Failed to build Gilbert locally, $_"
                    }
                }
            }

            Panic "Failed to download Gilbert: $_"
        }
    }

    if($PSVersionTable.PSVersion.Major -lt 3){
        $mj=$PSVersionTable.PSVersion.Major
        $mn=$PSVersionTable.PSVersion.Minor
        Write-Err "ERROR: This script requires PowerShell version 3 and above`nYour version is v$mj.$mn`n"
        throw 'Unsupported PowerShell Version'
    }

    Main
}