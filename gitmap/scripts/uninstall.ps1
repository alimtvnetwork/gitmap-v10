<#
.SYNOPSIS
    Uninstaller for gitmap CLI.

.DESCRIPTION
    Removes the gitmap binary, cleans up the install directory,
    and removes the directory from the user PATH.

.PARAMETER InstallDir
    Directory where gitmap is installed. Default: $env:LOCALAPPDATA\gitmap

.EXAMPLE
    irm https://raw.githubusercontent.com/alimtvnetwork/gitmap-v6/main/gitmap/scripts/uninstall.ps1 | iex

.NOTES
    Repository: https://github.com/alimtvnetwork/gitmap-v6
#>

param(
    [string]$InstallDir = ""
)

$ErrorActionPreference = "Stop"

$BinaryName = "gitmap.exe"

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

# --- Resolve install directory ---

function Resolve-InstallDir([string]$dir) {
    if ($dir -ne "") { return $dir }
    return Join-Path $env:LOCALAPPDATA "gitmap"
}

# --- Remove from PATH ---

function Test-PathEntry([string]$pathValue, [string]$dir) {
    if ([string]::IsNullOrWhiteSpace($pathValue)) { return $false }
    $parts = $pathValue -split ";"
    foreach ($part in $parts) {
        if ($part.Trim() -ieq $dir) { return $true }
    }
    return $false
}

function Remove-FromPath([string]$dir) {
    $currentUserPath = [Environment]::GetEnvironmentVariable("PATH", "User")

    if (-not (Test-PathEntry $currentUserPath $dir)) {
        Write-Step "Install directory not found in PATH (nothing to remove)."
        return
    }

    $parts = ($currentUserPath -split ";") | Where-Object { $_.Trim() -ine $dir -and $_.Trim() -ne "" }
    $newPath = $parts -join ";"

    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")

    # Broadcast change
    Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;

public static class GitMapUninstallNative {
    [DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
    public static extern IntPtr SendMessageTimeout(
        IntPtr hWnd, uint Msg, UIntPtr wParam, string lParam,
        uint fuFlags, uint uTimeout, out UIntPtr lpdwResult
    );
}
"@ -ErrorAction SilentlyContinue | Out-Null

    $HWND_BROADCAST = [IntPtr]0xffff
    $WM_SETTINGCHANGE = 0x001A
    $SMTO_ABORTIFHUNG = 0x0002
    [UIntPtr]$result = [UIntPtr]::Zero

    [void][GitMapUninstallNative]::SendMessageTimeout(
        $HWND_BROADCAST, $WM_SETTINGCHANGE, [UIntPtr]::Zero,
        "Environment", $SMTO_ABORTIFHUNG, 5000, [ref]$result
    )

    # Update current session
    $machinePath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
    $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    $freshPath = @($machinePath, $userPath) | Where-Object { $_ } | ForEach-Object { $_.TrimEnd(";") }
    $env:PATH = ($freshPath -join ";")

    Write-OK "Removed from PATH."
}

# --- Remove files ---

function Remove-InstallDir([string]$dir) {
    if (-not (Test-Path $dir)) {
        Write-Err "Install directory not found: $dir"
        return $false
    }

    $binPath = Join-Path $dir $BinaryName
    if (Test-Path $binPath) {
        Remove-Item $binPath -Force
        Write-OK "Removed $BinaryName"
    }

    # Remove remaining files (data, old binaries, etc.)
    $remaining = Get-ChildItem -Path $dir -Force -ErrorAction SilentlyContinue
    if ($remaining.Count -eq 0) {
        Remove-Item $dir -Force
        Write-OK "Removed install directory: $dir"
    }
    else {
        Write-Step "Directory not empty, keeping: $dir"
        Write-Step "  Remaining: $($remaining.Name -join ', ')"
    }

    return $true
}

# --- Main ---

function Main {
    Write-Host ""
    Write-Host "  gitmap uninstaller" -ForegroundColor White
    Write-Host ""

    $resolvedDir = Resolve-InstallDir $InstallDir

    Write-Step "Uninstalling from $resolvedDir..."

    $removed = Remove-InstallDir $resolvedDir
    Remove-FromPath $resolvedDir

    Write-Host ""
    if ($removed) {
        Write-OK "gitmap has been uninstalled."
    }
    else {
        Write-Err "gitmap was not found at $resolvedDir."
    }
    Write-Host ""
}

Main
