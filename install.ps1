#Requires -Version 5.1
<#
.SYNOPSIS
    One-line installer for the opencode-console-notify plugin.

.DESCRIPTION
    Copies plugin/console-notify.js into the OpenCode plugins directory and
    registers the OpenCode.Notifier AppUserModelID in HKCU (per-user, no
    administrator rights required).

    One-line install from GitHub:

        irm https://raw.githubusercontent.com/IbatoLionDev/opencode-console-notify/main/install.ps1 | iex

    Local checkout (offline / development, supports all parameters):

        .\install.ps1 [-PluginsDir <path>] [-Uninstall]

    When the script runs from a checkout that contains plugin\console-notify.js
    the local file is used and nothing is downloaded. Otherwise the plugin
    source is fetched from the URL above.

.PARAMETER Uninstall
    Remove the installed plugin file and the AUMID registration instead of
    installing.

.PARAMETER PluginsDir
    Destination directory for console-notify.js.
    Defaults to OpenCode's Windows config directory:
    <HOME>\.config/opencode\plugins

.EXAMPLE
    .\install.ps1

.EXAMPLE
    .\install.ps1 -PluginsDir "$env:TEMP\ocn-t2\plugins"

.EXAMPLE
    .\install.ps1 -Uninstall
#>
[CmdletBinding()]
param(
    [switch] $Uninstall,
    [string] $PluginsDir = (Join-Path $HOME '.config/opencode\plugins')
)

# The whole body lives in this function so that $ErrorActionPreference,
# Set-StrictMode and the helper variables stay scoped to it. Under
# `irm ... | iex` the script text runs in the caller's session scope, and
# leaking those preferences would change how the user's shell behaves.
function Invoke-ConsoleNotifyInstall {
    [CmdletBinding()]
    param(
        [switch] $Uninstall,
        [string] $PluginsDir
    )

    $ErrorActionPreference = 'Stop'
    Set-StrictMode -Version Latest
    $ProgressPreference = 'SilentlyContinue'

    if ([string]::IsNullOrWhiteSpace($PluginsDir)) {
        throw 'PluginsDir must not be empty.'
    }

    # Same AUMID key the plugin registers on load (plugin/console-notify.js).
    $aumidKey = 'HKCU:\Software\Classes\AppUserModelId\OpenCode.Notifier'
    $pluginName = 'console-notify.js'

    # ------------------------------------------------------------------
    # Uninstall
    # ------------------------------------------------------------------
    if ($Uninstall) {
        $dest = Join-Path $PluginsDir $pluginName
        if (Test-Path -LiteralPath $dest) {
            Remove-Item -LiteralPath $dest -Force
            Write-Output "Removed plugin file: $dest"
        } else {
            Write-Output "Plugin file not present, nothing to remove: $dest"
        }

        # Only the exact AUMID key is ever removed - never its parent key.
        if (Test-Path -LiteralPath $aumidKey) {
            Remove-Item -LiteralPath $aumidKey -Recurse -Force
            Write-Output "Removed AUMID registry key: $aumidKey"
        } else {
            Write-Output "AUMID registry key not present, nothing to remove: $aumidKey"
        }

        Write-Output 'Uninstall complete. Restart OpenCode to unload the plugin.'
        return
    }

    # ------------------------------------------------------------------
    # Install
    # ------------------------------------------------------------------
    # Windows PowerShell 5.1 may not negotiate TLS 1.2 by default.
    [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

    # `irm ... | iex` has no $PSScriptRoot / $PSCommandPath of its own. Probe
    # with Get-Variable so a missing automatic variable can never throw under
    # Set-StrictMode, and treat an empty value as "not a file run".
    $scriptRoot = Get-Variable -Name PSScriptRoot -ValueOnly -ErrorAction SilentlyContinue

    $bytes = $null
    $sourceDescription = $null

    if (-not [string]::IsNullOrWhiteSpace([string]$scriptRoot)) {
        # Local checkout / dev path: uses the file on disk, no network.
        $localPlugin = Join-Path (Join-Path $scriptRoot 'plugin') $pluginName
        if (Test-Path -LiteralPath $localPlugin -PathType Leaf) {
            $bytes = [IO.File]::ReadAllBytes($localPlugin)
            $sourceDescription = "local file $localPlugin (no download)"
        }
    }

    if ($null -eq $bytes) {
        $url = 'https://raw.githubusercontent.com/IbatoLionDev/opencode-console-notify/main/plugin/console-notify.js'
        try {
            $response = Invoke-WebRequest -Uri $url -UseBasicParsing -UserAgent 'opencode-console-notify/install.ps1'
        } catch {
            throw "Failed to download the plugin source from $url : $($_.Exception.Message)"
        }
        # Keep raw bytes when available so UTF-8 content and line endings are
        # byte-identical to the committed file (SHA256 parity with the repo).
        $streamProperty = $response.PSObject.Properties['RawContentStream']
        if (($null -ne $streamProperty) -and ($null -ne $streamProperty.Value)) {
            $bytes = $streamProperty.Value.ToArray()
        } else {
            $bytes = [Text.Encoding]::UTF8.GetBytes([string]$response.Content)
        }
        $sourceDescription = "downloaded from $url"
    }

    if (-not (Test-Path -LiteralPath $PluginsDir)) {
        # Windows PowerShell 5.1 New-Item has no -LiteralPath; -Path is correct
        # here (the directory path never contains wildcard characters).
        New-Item -Path $PluginsDir -ItemType Directory -Force | Out-Null
    }

    $dest = Join-Path $PluginsDir $pluginName
    $tempFile = Join-Path $PluginsDir ("{0}.{1}.tmp" -f $pluginName, ([Guid]::NewGuid().ToString('N')))
    try {
        # Write to a temp file in the same directory, then rename into place,
        # so a running OpenCode never sees a half-written plugin file.
        [IO.File]::WriteAllBytes($tempFile, $bytes)
        Move-Item -LiteralPath $tempFile -Destination $dest -Force
    } finally {
        if (Test-Path -LiteralPath $tempFile) {
            Remove-Item -LiteralPath $tempFile -Force -ErrorAction SilentlyContinue
        }
    }

    # Register the AUMID exactly like plugin/console-notify.js does: HKCU only,
    # no admin, idempotent (the Test-Path / Get-ItemProperty guards make a
    # re-run a no-op).
    if (-not (Test-Path -LiteralPath $aumidKey)) {
        # PS 5.1 New-Item has no -LiteralPath; the key path is a fixed literal,
        # so -Path is safe.
        New-Item -Path $aumidKey -Force | Out-Null
    }
    $displayName = Get-ItemProperty -LiteralPath $aumidKey -Name DisplayName -ErrorAction SilentlyContinue
    if ($null -eq $displayName) {
        Set-ItemProperty -LiteralPath $aumidKey -Name DisplayName -Value 'OpenCode' -Type String
    }

    $hash = (Get-FileHash -LiteralPath $dest -Algorithm SHA256).Hash

    Write-Output "Installed: $dest"
    Write-Output "SHA256:    $hash"
    Write-Output "Source:    $sourceDescription"
    Write-Output "AUMID:     registered ($aumidKey, DisplayName=OpenCode)"
    Write-Output 'Restart OpenCode to load the plugin.'
}

$installError = $null
try {
    Invoke-ConsoleNotifyInstall -Uninstall:$Uninstall -PluginsDir $PluginsDir
} catch {
    $installError = $_
} finally {
    # Do not leave the helper function behind in the caller's session after
    # `irm ... | iex`.
    if (Test-Path -LiteralPath 'function:Invoke-ConsoleNotifyInstall') {
        Remove-Item -LiteralPath 'function:Invoke-ConsoleNotifyInstall' -Force
    }
}

if ($null -ne $installError) {
    # Re-throw instead of calling `exit`: `powershell -File install.ps1` turns
    # this into a non-zero exit code, while `iex` surfaces the error without
    # terminating the user's interactive session.
    throw "opencode-console-notify install failed: $($installError.Exception.Message)"
}
