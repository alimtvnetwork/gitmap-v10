<#
.SYNOPSIS
    One-liner installer for gitmap CLI.

.DESCRIPTION
    Downloads the latest gitmap release from GitHub, verifies checksums,
    extracts to a local directory, and adds it to PATH.

.PARAMETER Version
    Install a specific version (e.g. v2.48.0). Default: latest.

.PARAMETER InstallDir
    Target directory. Default: $env:LOCALAPPDATA\gitmap

.PARAMETER NoPath
    Skip adding to PATH.

.PARAMETER Arch
    Force architecture (amd64, arm64). Default: auto-detect.

.EXAMPLE
    irm https://raw.githubusercontent.com/alimtvnetwork/gitmap-v9/main/gitmap/scripts/install.ps1 | iex

.EXAMPLE
    & ./install.ps1 -Version v2.48.0

.NOTES
    Repository: https://github.com/alimtvnetwork/gitmap-v9
#>

param(
    [string]$Version = "",
    [string]$InstallDir = "",
    [string]$Arch = "",
    [switch]$NoPath,
    [switch]$Uninstall,
    [switch]$NoDiscovery,
    [int]$ProbeCeiling = 30,
    # --- Uninstall safety knobs ---
    # -Force: skip the "is this really a gitmap install?" guard AND
    #         skip the keep-data prompt (defaults to keep).
    # -KeepData / -PurgeData: explicit data-folder choice; when both
    #         are absent the user is prompted interactively.
    [switch]$Force,
    [switch]$KeepData,
    [switch]$PurgeData
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

$Repo = "alimtvnetwork/gitmap-v9"
$BinaryName = "gitmap.exe"
$InstallerVersion = "1.0.0"

# ---------------------------------------------------------------------------
# Versioned repo discovery (spec/01-app/95-installer-script-find-latest-repo.md)
# ---------------------------------------------------------------------------

function Split-RepoSuffix([string]$repo) {
    if ($repo -match '^([^/]+)/(.+)-v(\d+)$') {
        return @{ Owner = $Matches[1]; Stem = $Matches[2]; N = [int]$Matches[3] }
    }
    return $null
}

function Test-RepoExists([string]$url) {
    try {
        $resp = Invoke-WebRequest -Uri $url -Method Head -TimeoutSec 5 `
            -UseBasicParsing -ErrorAction Stop
        return ($resp.StatusCode -eq 200)
    } catch {
        return $false
    }
}

function Resolve-EffectiveRepo([string]$repo, [int]$ceiling) {
    $parts = Split-RepoSuffix $repo
    if ($null -eq $parts) {
        Write-Host "  [discovery] no -v<N> suffix on '$repo'; installing baseline as-is"
        return $repo
    }

    $owner = $parts.Owner; $stem = $parts.Stem; $baseline = $parts.N
    $effective = $baseline

    Write-Host "  [discovery] baseline: $owner/$stem-v$baseline"
    Write-Host "  [discovery] probe ceiling: $ceiling"

    for ($m = $baseline + 1; $m -le $ceiling; $m++) {
        $url = "https://github.com/$owner/$stem-v$m"
        if (Test-RepoExists $url) {
            Write-Host "  [discovery] HEAD $url ... HIT"
            $effective = $m
        } else {
            Write-Host "  [discovery] HEAD $url ... MISS (fail-fast)"
            break
        }
    }

    if ($effective -eq $baseline) {
        Write-Host "  [discovery] no higher version found; using baseline -v$baseline"
        return $repo
    }

    Write-Host "  [discovery] effective: $owner/$stem-v$effective (was -v$baseline)"
    return "$owner/$stem-v$effective"
}

function Invoke-DelegatedFullInstaller([string]$effectiveRepo) {
    $delegatedUrl = "https://raw.githubusercontent.com/$effectiveRepo/main/gitmap/scripts/install.ps1"
    Write-Host "  [discovery] delegating to $delegatedUrl"

    $env:INSTALLER_DELEGATED = "1"
    try {
        $script = (Invoke-WebRequest -Uri $delegatedUrl -UseBasicParsing -TimeoutSec 15).Content
    } catch {
        Write-Host "  [discovery] [WARN] could not fetch delegated installer: $_" -ForegroundColor Yellow
        Write-Host "  [discovery] falling back to baseline installer" -ForegroundColor Yellow
        Remove-Item Env:INSTALLER_DELEGATED -ErrorAction SilentlyContinue
        return $false
    }

    $block = [ScriptBlock]::Create($script)
    $passArgs = @{ ProbeCeiling = $ProbeCeiling }
    if (-not [string]::IsNullOrWhiteSpace($Version))    { $passArgs.Version    = $Version }
    if (-not [string]::IsNullOrWhiteSpace($InstallDir)) { $passArgs.InstallDir = $InstallDir }
    if (-not [string]::IsNullOrWhiteSpace($Arch))       { $passArgs.Arch       = $Arch }
    if ($NoPath)    { $passArgs.NoPath    = $true }
    if ($Uninstall) { $passArgs.Uninstall = $true }
    if ($Force)     { $passArgs.Force     = $true }
    if ($KeepData)  { $passArgs.KeepData  = $true }
    if ($PurgeData) { $passArgs.PurgeData = $true }

    try {
        & $block @passArgs
    }
    catch {
        throw
    }
    return $true
}

if ($env:INSTALLER_DELEGATED -eq "1") {
    Write-Host "  [discovery] INSTALLER_DELEGATED=1; skipping discovery (loop guard)"
} elseif ($NoDiscovery) {
    Write-Host "  [discovery] -NoDiscovery set; skipping probe"
} elseif (-not [string]::IsNullOrWhiteSpace($Version)) {
    # Pinned-version contract (spec/07-generic-release/08-pinned-version-install-snippet.md):
    # When -Version is supplied, install EXACTLY that version from the embedded $Repo.
    # Skip versioned-repo discovery so a snippet copied from a v3.x release page
    # never silently jumps to the v4 repo's latest tag.
    Write-Host "  [discovery] -Version $Version pinned; skipping repo probe (exact-version install)"
} else {
    $effective = Resolve-EffectiveRepo $Repo $ProbeCeiling
    if ($effective -ne $Repo) {
        if (Invoke-DelegatedFullInstaller $effective) { return }
    }
}



# --- Logging helpers ---

function Write-Step([string]$msg) {
    Write-Host "  $msg" -ForegroundColor Cyan
}

function Write-OK([string]$msg) {
    Write-Host "  $msg" -ForegroundColor Green
}

function Write-Err([string]$msg) {
    Write-Host "  $msg" -ForegroundColor Red
}

function Set-InstallerExitCode([int]$exitCode) {
    $global:LASTEXITCODE = $exitCode
    [System.Environment]::ExitCode = $exitCode
}

function Write-FatalError($record, [int]$exitCode = 1) {
    Set-InstallerExitCode $exitCode
    $message = "Unknown PowerShell error"
    if ($record) {
        if ($record.Exception -and -not [string]::IsNullOrWhiteSpace($record.Exception.Message)) {
            $message = $record.Exception.Message
        }
        elseif (-not [string]::IsNullOrWhiteSpace($record.ToString())) {
            $message = $record.ToString()
        }
    }

    Write-Host ""
    Write-Err "FATAL: $message"

    if ($record) {
        if (-not [string]::IsNullOrWhiteSpace($record.ScriptStackTrace)) {
            Write-Err ""
            Write-Err "Script stack trace:"
            foreach ($line in ($record.ScriptStackTrace -split "`r?`n")) {
                if (-not [string]::IsNullOrWhiteSpace($line)) {
                    Write-Err "  $line"
                }
            }
        }

        if ($record.InvocationInfo) {
            $scriptName = $record.InvocationInfo.ScriptName
            if ([string]::IsNullOrWhiteSpace($scriptName)) {
                $scriptName = $PSCommandPath
            }

            Write-Err ""
            Write-Err "Failure context:"
            if (-not [string]::IsNullOrWhiteSpace($scriptName)) {
                Write-Err "  Script: $scriptName"
            }
            Write-Err "  Line: $($record.InvocationInfo.ScriptLineNumber)"
            if (-not [string]::IsNullOrWhiteSpace($record.InvocationInfo.Line)) {
                Write-Err "  Code: $($record.InvocationInfo.Line.Trim())"
            }
        }

        if ($record.CategoryInfo) {
            Write-Err ""
            Write-Err "CategoryInfo: $($record.CategoryInfo)"
        }

        if (-not [string]::IsNullOrWhiteSpace($record.FullyQualifiedErrorId)) {
            Write-Err "FullyQualifiedErrorId: $($record.FullyQualifiedErrorId)"
        }

        if ($record.Exception) {
            Write-Err ""
            Write-Err "Exception:"
            foreach ($line in ($record.Exception.ToString() -split "`r?`n")) {
                if (-not [string]::IsNullOrWhiteSpace($line)) {
                    Write-Err "  $line"
                }
            }
        }
    }

    Write-Host ""
    return
}

# --- Resolve install directory ---

function Resolve-InstallDir([string]$dir) {
    if ($dir -ne "") { return $dir }
    return Join-Path $env:LOCALAPPDATA "gitmap"
}

# --- Detect architecture ---

function Resolve-Arch([string]$arch) {
    if ($arch -ne "") { return $arch }

    $cpu = $env:PROCESSOR_ARCHITECTURE
    switch ($cpu) {
        "AMD64"   { return "amd64" }
        "ARM64"   { return "arm64" }
        "x86"     { return "amd64" }
        default   { return "amd64" }
    }
}

# --- Resolve version (latest or pinned) ---

function Resolve-Version([string]$version) {
    if ($version -ne "") { return $version }

    $url = "https://api.github.com/repos/$Repo/releases/latest"
    Write-Step "Fetching latest release..."
    Write-Step "  URL: $url"

    try {
        $response = Invoke-WebRequest -Uri $url -UseBasicParsing -ErrorAction Stop
        $release = $response.Content | ConvertFrom-Json
        return $release.tag_name
    }
    catch {
        $statusCode = "unknown"
        $body = ""

        if ($_.Exception.Response) {
            $statusCode = [int]$_.Exception.Response.StatusCode
            try {
                $reader = [System.IO.StreamReader]::new($_.Exception.Response.GetResponseStream())
                $body = $reader.ReadToEnd()
                $reader.Close()
            }
            catch {
                $body = $_.Exception.Message
            }
        }
        else {
            $body = $_.Exception.Message
        }

        Write-Err "Failed to fetch latest release"
        Write-Err "  HTTP $statusCode -- $url"
        if ($body) {
            Write-Err "  Response: $body"
        }
        Write-Err ""
        Write-Err "  Possible causes:"
        Write-Err "    - No published releases in the repository"
        Write-Err "    - Repository is private (needs authentication)"
        Write-Err "    - Repository name has changed"
        Write-Err ""
        Write-Err "  Try: https://github.com/$Repo/releases"
        exit 1
    }
}

# --- Strict-tag failure (spec/07-generic-release/09 §3) ---
# Print the canonical no-fallback message and exit 1. Called from
# Get-Asset whenever -Version was supplied explicitly and the
# requested release asset cannot be downloaded or verified.
function Stop-Strict([string]$detail) {
    Write-Err ""
    Write-Err "Error: requested release $Version not found in $Repo;"
    Write-Err "       refusing to fall back per strict-tag contract."
    Write-Err "       See spec/07-generic-release/09-generic-install-script-behavior.md `$3."
    if ($detail) { Write-Err "       Detail: $detail" }
    exit 1
}

# --- Download asset ---

function Get-Asset([string]$version, [string]$arch) {
    $assetName = "gitmap-${version}-windows-${arch}.zip"
    $baseUrl = "https://github.com/$Repo/releases/download/$version"
    $assetUrl = "$baseUrl/$assetName"
    $checksumUrl = "$baseUrl/checksums.txt"

    # Strict mode: -Version was supplied explicitly. Any failure here
    # MUST exit 1 with the canonical message and MUST NOT fall back.
    $strict = -not [string]::IsNullOrWhiteSpace($Version)
    if ($strict) {
        Write-Step "  [strict] download: $assetUrl"
    }

    $tmpDir = Join-Path $env:TEMP "gitmap-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    $zipPath = Join-Path $tmpDir $assetName
    $checksumPath = Join-Path $tmpDir "checksums.txt"

    Write-Step "Downloading $assetName ($version)..."

    try {
        Invoke-WebRequest -Uri $assetUrl -OutFile $zipPath -UseBasicParsing
        Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumPath -UseBasicParsing
    }
    catch {
        Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        if ($strict) {
            Stop-Strict "download failed: $($_.Exception.Message)"
        }
        Write-Err "Download failed: $_"
        exit 1
    }

    # Verify checksum
    Write-Step "Verifying checksum..."
    $expectedLine = (Get-Content $checksumPath | Where-Object { $_ -match $assetName })
    if (-not $expectedLine) {
        Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        if ($strict) {
            Stop-Strict "asset $assetName not listed in checksums.txt for $version"
        }
        Write-Err "Asset not found in checksums.txt"
        exit 1
    }

    $expectedHash = ($expectedLine -split '\s+')[0]
    $actualHash = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash.ToLower()

    if ($actualHash -ne $expectedHash) {
        Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        if ($strict) {
            Stop-Strict "checksum mismatch for $assetName (expected $expectedHash, got $actualHash)"
        }
        Write-Err "Checksum mismatch!"
        Write-Err "  Expected: $expectedHash"
        Write-Err "  Got:      $actualHash"
        exit 1
    }

    Write-OK "Checksum verified."
    return @{ ZipPath = $zipPath; TmpDir = $tmpDir }
}

# --- Extract and install ---

function Install-Binary([string]$zipPath, [string]$installDir) {
    Write-Step "Installing to $installDir..."

    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }

    $targetPath = Join-Path $installDir $BinaryName
    $extractDir = Join-Path $installDir ".install-extract"

    # Rename-first strategy for running binary
    if (Test-Path $targetPath) {
        $oldPath = "$targetPath.old"
        if (Test-Path $oldPath) { Remove-Item $oldPath -Force }
        Rename-Item $targetPath $oldPath -Force
    }

    if (Test-Path $extractDir) {
        Remove-Item $extractDir -Recurse -Force -ErrorAction SilentlyContinue
    }

    New-Item -ItemType Directory -Path $extractDir -Force | Out-Null
    Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force

    # Match exact names OR versioned patterns like gitmap-v4.54.6-windows-amd64.exe
    $candidateNames = @(
        $BinaryName,
        [System.IO.Path]::GetFileNameWithoutExtension($BinaryName),
        "gitmap-windows-amd64.exe",
        "gitmap-windows-arm64.exe"
    )

    $extracted = Get-ChildItem -Path $extractDir -File -Recurse |
        Where-Object {
            ($candidateNames -icontains $_.Name) -or
            ($_.Name -match "^gitmap-v[\d.]+-windows-(amd64|arm64)\.exe$")
        } |
        Select-Object -First 1

    if (-not $extracted) {
        $availableFiles = Get-ChildItem -Path $extractDir -File -Recurse | Select-Object -ExpandProperty FullName
        Remove-Item $extractDir -Recurse -Force -ErrorAction SilentlyContinue
        Write-Err "Installed archive did not contain $BinaryName"
        if ($availableFiles) {
            Write-Err "Archive files:"
            foreach ($file in $availableFiles) {
                Write-Err "  $file"
            }
        }
        exit 1
    }

    Move-Item $extracted.FullName $targetPath -Force

    Remove-Item $extractDir -Recurse -Force -ErrorAction SilentlyContinue

    if (-not (Test-Path $targetPath)) {
        Write-Err "Install failed: $BinaryName was not written to $installDir"
        exit 1
    }

    # Cleanup .old
    $oldPath = "$targetPath.old"
    if (Test-Path $oldPath) {
        Remove-Item $oldPath -Force -ErrorAction SilentlyContinue
    }

    Write-OK "Installed $BinaryName to $installDir"
}

# --- Download and extract docs-site.zip release asset ---
# Required for `gitmap help-dashboard` (hd). Best-effort: skip silently
# if the release does not bundle docs-site.zip (older versions).
function Install-DocsSite([string]$version, [string]$installDir) {
    $assetName = "docs-site.zip"
    $assetUrl = "https://github.com/$Repo/releases/download/$version/$assetName"
    $tmpZip = Join-Path $env:TEMP "gitmap-docs-site-$(Get-Random).zip"

    Write-Step "Downloading docs-site.zip ($version)..."

    try {
        Invoke-WebRequest -Uri $assetUrl -OutFile $tmpZip -UseBasicParsing -ErrorAction Stop
    }
    catch {
        Write-Step "  docs-site.zip not available for $version - skipping (gitmap hd may not work)"
        Remove-Item $tmpZip -Force -ErrorAction SilentlyContinue
        return
    }

    # Remove any existing docs-site/ before extracting fresh.
    $docsDir = Join-Path $installDir "docs-site"
    if (Test-Path $docsDir) {
        Remove-Item $docsDir -Recurse -Force -ErrorAction SilentlyContinue
    }

    try {
        # The zip's internal layout is docs-site/dist/... so it extracts directly.
        Expand-Archive -Path $tmpZip -DestinationPath $installDir -Force
        Write-OK "Installed docs-site to $docsDir"
    }
    catch {
        Write-Err "Failed to extract docs-site.zip: $_"
    }
    finally {
        Remove-Item $tmpZip -Force -ErrorAction SilentlyContinue
    }
}

# --- Add to PATH ---

function Test-PathEntry([string]$pathValue, [string]$dir) {
    if ([string]::IsNullOrWhiteSpace($pathValue)) {
        return $false
    }

    $parts = $pathValue -split ";"

    foreach ($part in $parts) {
        if ($part.Trim() -ieq $dir) {
            return $true
        }
    }

    return $false
}

function Rebuild-SessionPath([string]$dir) {
    # Rebuild session PATH from registry (Machine + User) to pick up any changes
    $machinePath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
    $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    $parts = @()
    if ($machinePath) { $parts += $machinePath.TrimEnd(";") }
    if ($userPath) { $parts += $userPath.TrimEnd(";") }
    $rebuilt = $parts -join ";"

    # Ensure install dir is present even if not yet persisted
    if (-not (Test-PathEntry $rebuilt $dir)) {
        $rebuilt = $rebuilt.TrimEnd(";") + ";" + $dir
    }

    return $rebuilt
}

function Broadcast-EnvironmentChange {
    Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;

public static class GitMapEnvNative {
    [DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
    public static extern IntPtr SendMessageTimeout(
        IntPtr hWnd,
        uint Msg,
        UIntPtr wParam,
        string lParam,
        uint fuFlags,
        uint uTimeout,
        out UIntPtr lpdwResult
    );
}
"@ -ErrorAction SilentlyContinue | Out-Null

    $HWND_BROADCAST = [IntPtr]0xffff
    $WM_SETTINGCHANGE = 0x001A
    $SMTO_ABORTIFHUNG = 0x0002
    [UIntPtr]$result = [UIntPtr]::Zero

    [void][GitMapEnvNative]::SendMessageTimeout(
        $HWND_BROADCAST,
        $WM_SETTINGCHANGE,
        [UIntPtr]::Zero,
        "Environment",
        $SMTO_ABORTIFHUNG,
        5000,
        [ref]$result
    )
}

function Add-ToPath([string]$dir) {
    $modified = @()

    # --- 1. Windows User PATH (registry) — covers CMD + new PowerShell windows ---
    $currentUserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    $userHasDir = Test-PathEntry $currentUserPath $dir

    if (-not $userHasDir) {
        if ([string]::IsNullOrWhiteSpace($currentUserPath)) {
            $newPath = $dir
        }
        else {
            $newPath = $currentUserPath.TrimEnd(";") + ";" + $dir
        }

        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Broadcast-EnvironmentChange
        Write-OK "Added to User PATH (registry)."
        $modified += "User PATH (registry)"
    }
    else {
        Write-Step "Already in User PATH (registry)."
    }

    # --- 2. PowerShell $PROFILE — ensures PATH in all PowerShell sessions ---
    $psProfilePath = $PROFILE
    if ($psProfilePath) {
        $exportLine = "`$env:PATH = `"$dir;`$env:PATH`""
        $marker = "# gitmap-path"
        $markerLine = "$exportLine $marker"

        $profileExists = Test-Path $psProfilePath
        $alreadyPresent = $false

        if ($profileExists) {
            $content = Get-Content $psProfilePath -Raw -ErrorAction SilentlyContinue
            if ($content -and ($content -match [regex]::Escape($marker))) {
                $alreadyPresent = $true
            }
        }

        if (-not $alreadyPresent) {
            # Ensure parent directory exists
            $profileDir = Split-Path $psProfilePath -Parent
            if ($profileDir -and -not (Test-Path $profileDir)) {
                New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
            }
            Add-Content -Path $psProfilePath -Value "`n$markerLine" -Encoding UTF8
            Write-OK "Added to PowerShell profile: $psProfilePath"
            $modified += "PowerShell `$PROFILE"
        }
        else {
            Write-Step "Already in PowerShell profile."
        }
    }

    # --- 3. Git Bash profiles (~/.bashrc, ~/.bash_profile) ---
    $homeDir = $env:USERPROFILE
    if ($homeDir) {
        $bashExportLine = "export PATH=`"$($dir -replace '\\','/'):`$PATH`""
        $bashMarker = "# gitmap-path"
        $bashProfiles = @(
            (Join-Path $homeDir ".bashrc"),
            (Join-Path $homeDir ".bash_profile")
        )

        foreach ($bashProfile in $bashProfiles) {
            $bashAlreadyPresent = $false
            $bashProfileName = Split-Path $bashProfile -Leaf

            if (Test-Path $bashProfile) {
                $bashContent = Get-Content $bashProfile -Raw -ErrorAction SilentlyContinue
                if ($bashContent -and ($bashContent -match [regex]::Escape($bashMarker))) {
                    $bashAlreadyPresent = $true
                }
            }

            if (-not $bashAlreadyPresent) {
                Add-Content -Path $bashProfile -Value "`n$bashExportLine $bashMarker" -Encoding UTF8
                Write-OK "Added to Git Bash profile: ~/$bashProfileName"
                $modified += "~/$bashProfileName"
            }
            else {
                Write-Step "Already in ~/$bashProfileName."
            }
        }
    }

    if ($modified.Count -gt 0) {
        return @{ Target = ($modified -join ", "); Status = "added" }
    }

    return @{ Target = "All profiles"; Status = "already present" }
}

# --- Remove from PATH (uninstall) ---

function Remove-FromPath([string]$dir) {
    $removed = @()
    $marker = "# gitmap-path"

    # --- 1. Windows User PATH (registry) ---
    $currentUserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentUserPath -and (Test-PathEntry $currentUserPath $dir)) {
        $parts = ($currentUserPath -split ";") | Where-Object { $_.Trim() -ine $dir -and $_.Trim() -ne "" }
        $newPath = $parts -join ";"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Broadcast-EnvironmentChange
        Write-OK "Removed from User PATH (registry)."
        $removed += "User PATH (registry)"
    }

    # --- 2. PowerShell $PROFILE ---
    $psProfilePath = $PROFILE
    if ($psProfilePath -and (Test-Path $psProfilePath)) {
        $lines = Get-Content $psProfilePath
        $filtered = $lines | Where-Object { $_ -notmatch [regex]::Escape($marker) }
        if ($filtered.Count -lt $lines.Count) {
            $filtered | Set-Content $psProfilePath -Encoding UTF8
            Write-OK "Removed marker lines from PowerShell profile: $psProfilePath"
            $removed += "PowerShell `$PROFILE"
        }
    }

    # --- 3. Git Bash profiles ---
    $homeDir = $env:USERPROFILE
    if ($homeDir) {
        $bashProfiles = @(
            (Join-Path $homeDir ".bashrc"),
            (Join-Path $homeDir ".bash_profile")
        )

        foreach ($bashProfile in $bashProfiles) {
            if (Test-Path $bashProfile) {
                $lines = Get-Content $bashProfile
                $filtered = $lines | Where-Object { $_ -notmatch [regex]::Escape($marker) }
                if ($filtered.Count -lt $lines.Count) {
                    $filtered | Set-Content $bashProfile -Encoding UTF8
                    $name = Split-Path $bashProfile -Leaf
                    Write-OK "Removed marker lines from ~/$name"
                    $removed += "~/$name"
                }
            }
        }
    }

    if ($removed.Count -gt 0) {
        Write-Host ""
        Write-OK "PATH entries removed from: $($removed -join ', ')"
    }
    else {
        Write-Step "No gitmap PATH entries found in any profile."
    }

    return $removed
}

function Write-InstallSummary([string]$version, [string]$binPath, [string]$installDir, [hashtable]$pathResult, [bool]$isNoPath) {
    Write-Host ""
    Write-Host "  -----------------------------------------------" -ForegroundColor DarkGray
    Write-Host "  gitmap install summary" -ForegroundColor White
    Write-Host "  -----------------------------------------------" -ForegroundColor DarkGray
    Write-Host "    Version    : $version"
    Write-Host "    Binary     : $binPath"
    Write-Host "    Install Dir: $installDir"

    if ($isNoPath) {
        Write-Host "    PATH       : skipped (-NoPath)"
        return
    }

    Write-Host "    PATH target: $($pathResult.Target) ($($pathResult.Status))"
    Write-Host "    Session    : refreshed for current PowerShell session"

    Write-Host ""
    Write-Host "  Profiles modified:" -ForegroundColor White
    Write-Host "    - User PATH (registry)  : CMD, new PowerShell windows"
    Write-Host "    - PowerShell `$PROFILE    : all PowerShell sessions"
    Write-Host "    - ~/.bashrc             : Git Bash interactive shells"
    Write-Host "    - ~/.bash_profile       : Git Bash login shells"

    Write-Host ""
    Write-Host "  If gitmap is not found in a new terminal, run:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "    PowerShell:  `$env:PATH = `"$installDir;`$env:PATH`"" -ForegroundColor Cyan
    Write-Host "    CMD:         set PATH=$installDir;%PATH%" -ForegroundColor Cyan
    Write-Host "    Git Bash:    export PATH=`"$($installDir -replace '\\','/'):`$PATH`"" -ForegroundColor Cyan
    Write-Host ""
}

# --- Post-install verification ---

# Invoke-InstallVerification runs the three post-install checks the
# user asked for: (1) print installed version by invoking the binary
# directly, (2) confirm `gitmap` resolves via Get-Command in the
# refreshed session PATH, (3) ensure the per-install data folder
# exists (create on miss). All checks emit PASS/WARN; none throw,
# because the binary is already on disk and the user can recover.
function Invoke-InstallVerification([string]$binPath, [string]$installDir, [bool]$isNoPath) {
    $dataDir = Join-Path $installDir "data"

    Write-Host ""
    Write-Step "Verifying installation"

    # 1. Version
    if (Test-Path $binPath) {
        try {
            $verLine = (& $binPath version 2>&1 | Out-String).Trim().Split("`n")[0]
            Write-Host ("    PASS  Version: {0}" -f $verLine) -ForegroundColor Green
        }
        catch {
            Write-Host ("    WARN  Could not run {0} version: {1}" -f $binPath, $_) -ForegroundColor Yellow
        }
    }
    else {
        Write-Host ("    WARN  Binary missing: {0}" -f $binPath) -ForegroundColor Yellow
    }

    # 2. PATH active in this session
    $resolved = Get-Command $BinaryName -ErrorAction SilentlyContinue
    if ($resolved) {
        Write-Host ("    PASS  PATH active: {0} -> {1}" -f $BinaryName, $resolved.Source) -ForegroundColor Green
    }
    elseif ($isNoPath) {
        Write-Host ("    WARN  PATH skipped (-NoPath); invoke with full path: {0}" -f $binPath) -ForegroundColor Yellow
    }
    else {
        Write-Host ("    WARN  {0} not on PATH yet — open a new terminal or reload `$PROFILE." -f $BinaryName) -ForegroundColor Yellow
    }

    # 3. Data folder
    if (Test-Path $dataDir) {
        Write-Host ("    PASS  Data folder exists: {0}" -f $dataDir) -ForegroundColor Green
    }
    else {
        try {
            New-Item -ItemType Directory -Path $dataDir -Force | Out-Null
            Write-Host ("    PASS  Data folder created: {0}" -f $dataDir) -ForegroundColor Green
        }
        catch {
            Write-Host ("    WARN  Could not create data folder: {0}" -f $dataDir) -ForegroundColor Yellow
        }
    }
}

# --- Main ---

function Main {
    Write-Host ""
    Write-Host "  gitmap installer v$InstallerVersion" -ForegroundColor White
    Write-Host "  github.com/$Repo" -ForegroundColor DarkGray
    Write-Host ""

    try {
        $resolvedArch = Resolve-Arch $Arch
        $resolvedVersion = Resolve-Version $Version
        $resolvedDir = Resolve-InstallDir $InstallDir

        $result = Get-Asset $resolvedVersion $resolvedArch

        try {
            Install-Binary $result.ZipPath $resolvedDir
        }
        finally {
            Remove-Item $result.TmpDir -Recurse -Force -ErrorAction SilentlyContinue
        }

        # Bundle the docs site so `gitmap help-dashboard` works after install.
        Install-DocsSite $resolvedVersion $resolvedDir

        $pathResult = @{ Target = "-NoPath"; Status = "skipped" }
        if (-not $NoPath) {
            $pathResult = Add-ToPath $resolvedDir

            # Also try Chocolatey refreshenv if available
            $refreshCmd = Get-Command refreshenv -ErrorAction SilentlyContinue
            if ($refreshCmd) {
                try { refreshenv | Out-Null } catch {}
            }

            # Force-rebuild $env:PATH in this scope so gitmap is usable immediately
            $script:NewPath = Rebuild-SessionPath $resolvedDir
        }
        else {
            $script:NewPath = $env:PATH
        }

        return @{ InstallDir = $resolvedDir; NewPath = $script:NewPath; Version = $resolvedVersion; PathResult = $pathResult }
    }
    catch {
        throw
    }
}

# --- Uninstall safety helpers ---

# Confirm-IsGitmapInstall verifies the binary at $binPath actually
# IS our gitmap CLI before any destructive action. We invoke
# `<binary> version` and look for the literal token "gitmap" in the
# output. A mismatch means the user pointed -InstallDir at the wrong
# folder (or a stale path collides with another tool) — bail loudly
# rather than rip PATH entries belonging to something else. -Force
# overrides the guard for advanced users.
function Confirm-IsGitmapInstall([string]$binPath, [bool]$isForce) {
    if (-not (Test-Path $binPath)) {
        # Nothing to verify against — also nothing to wreck. Allow
        # PATH cleanup so a half-installed/broken state can be removed.
        Write-Step "No binary at $binPath; skipping identity check."
        return $true
    }
    try {
        $out = (& $binPath version 2>&1 | Out-String)
    }
    catch {
        $out = ""
    }
    if ($out -match '(?i)\bgitmap\b') {
        Write-OK "Verified gitmap binary at $binPath."
        return $true
    }
    if ($isForce) {
        Write-Step "Identity check failed but -Force set; continuing."
        return $true
    }
    Write-Err "Refusing to uninstall: $binPath does not look like gitmap."
    Write-Err "  Output: $($out.Trim())"
    Write-Err "  Re-run with -Force to override, or pass the correct -InstallDir."
    return $false
}

# Resolve-DataChoice returns 'keep' or 'purge'. Honors -KeepData /
# -PurgeData unconditionally; -Force defaults to 'keep' (safer);
# otherwise prompts the user. When stdin is not interactive (e.g.
# `iex` in a non-tty pipeline) we also default to 'keep'.
function Resolve-DataChoice([string]$dataDir, [bool]$isKeep, [bool]$isPurge, [bool]$isForce) {
    if (-not (Test-Path $dataDir)) { return 'keep' }
    if ($isPurge) { return 'purge' }
    if ($isKeep) { return 'keep' }
    if ($isForce) { return 'keep' }
    if (-not [Environment]::UserInteractive) { return 'keep' }
    Write-Host ""
    Write-Host ("  Data folder found: {0}" -f $dataDir) -ForegroundColor Yellow
    $reply = Read-Host "  Delete user data too? [y/N]"
    if ($reply -match '^(y|yes)$') { return 'purge' }
    return 'keep'
}

# --- Uninstall mode ---
if ($Uninstall) {
    Write-Host ""
    Write-Host "  gitmap uninstaller" -ForegroundColor White
    Write-Host ""

    $resolvedDir = Resolve-InstallDir $InstallDir
    $binPath = Join-Path $resolvedDir $BinaryName
    $dataDir = Join-Path $resolvedDir "data"

    # Safety guard: confirm this is actually gitmap before we touch
    # PATH entries or delete anything. -Force bypasses the check.
    if (-not (Confirm-IsGitmapInstall $binPath $Force.IsPresent)) {
        Write-Host ""
        return
    }

    $dataChoice = Resolve-DataChoice $dataDir $KeepData.IsPresent $PurgeData.IsPresent $Force.IsPresent

    # Remove PATH entries from all profiles
    $removedProfiles = Remove-FromPath $resolvedDir

    # Remove binary and install directory
    if (Test-Path $binPath) {
        Remove-Item $binPath -Force -ErrorAction SilentlyContinue
        Write-OK "Removed binary: $binPath"
    }

    $oldPath = "$binPath.old"
    if (Test-Path $oldPath) {
        Remove-Item $oldPath -Force -ErrorAction SilentlyContinue
    }

    # Honor the data-folder choice BEFORE the empty-dir sweep so a
    # 'keep' run does not get the install dir removed out from under
    # the surviving data folder.
    if ($dataChoice -eq 'purge' -and (Test-Path $dataDir)) {
        Remove-Item $dataDir -Recurse -Force -ErrorAction SilentlyContinue
        Write-OK "Removed data folder: $dataDir"
    }
    elseif (Test-Path $dataDir) {
        Write-Step "Kept data folder: $dataDir"
    }

    # Remove install dir if empty
    if ((Test-Path $resolvedDir) -and @(Get-ChildItem $resolvedDir).Count -eq 0) {
        Remove-Item $resolvedDir -Force -ErrorAction SilentlyContinue
        Write-OK "Removed empty directory: $resolvedDir"
    }

    # Rebuild session PATH without the dir
    $machinePath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
    $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    $parts = @()
    if ($machinePath) { $parts += ($machinePath -split ";") }
    if ($userPath) { $parts += ($userPath -split ";") }
    $env:PATH = ($parts | Where-Object { $_.Trim() -ine $resolvedDir -and $_.Trim() -ne "" }) -join ";"

    Write-Host ""
    Write-OK "gitmap has been uninstalled."
    Write-Host ""
    return
}


try {
    $installResult = Main

    if (-not $installResult) {
        exit 1
    }

    # Set $env:PATH at the TOP-LEVEL script scope (not inside a function)
    # This ensures the change persists in the caller's session when run via iex
    $env:PATH = $installResult.NewPath

    # Verify the binary works
    $binPath = Join-Path $installResult.InstallDir $BinaryName
    $installedVersion = $installResult.Version
    if (Test-Path $binPath) {
        Write-Host ""
        try {
            $versionOutput = & $binPath version 2>&1
            $installedVersion = ($versionOutput | Out-String).Trim()
            Write-OK "gitmap $installedVersion"
        }
        catch {
            Write-Err "Binary found but failed to run: $_"
        }
    }
    else {
        Write-Err "Binary not found at $binPath"
    }

    Write-InstallSummary $installedVersion $binPath $installResult.InstallDir $installResult.PathResult $NoPath.IsPresent

    Invoke-InstallVerification $binPath $installResult.InstallDir $NoPath.IsPresent

    Write-Host ""
    Write-OK "Done! Run 'gitmap --help' to get started."
    Write-Host ""
}
catch {
    Write-FatalError $_ 1
}
