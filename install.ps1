#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Install gndec-qp - GNDEC Question Paper Downloader
.DESCRIPTION
    Downloads and installs the latest gndec-qp binary for Windows from GitHub Releases.
.LINK
    https://github.com/IshpreetSingh8264/gndec-qp
#>

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$Repo = 'IshpreetSingh8264/gndec-qp'
$Binary = 'qp'
$Green = [ConsoleColor]::Green
$Yellow = [ConsoleColor]::Yellow
$Red = [ConsoleColor]::Red
$Cyan = [ConsoleColor]::Cyan

function Write-Info  { Write-Host ":: " -ForegroundColor $Cyan -NoNewline; Write-Host "$args" }
function Write-Ok    { Write-Host "✓ " -ForegroundColor $Green -NoNewline; Write-Host "$args" }
function Write-Warn  { Write-Host "! " -ForegroundColor $Yellow -NoNewline; Write-Host "$args" }
function Write-Error { Write-Host "✗ " -ForegroundColor $Red -NoNewline; Write-Host "$args" }

function Detect-Arch {
    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    if ($arch -ne "amd64") {
        Write-Error "Only 64-bit Windows is supported."
        exit 1
    }
    Write-Info "Detected: windows / $arch"
    return $arch
}

function Get-LatestVersion {
    try {
        $apiUrl = "https://api.github.com/repos/$Repo/releases/latest"
        $response = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing
        return $response.tag_name
    } catch {
        Write-Warn "Could not fetch latest release. Using 'latest' tag."
        return "latest"
    }
}

function Download-Binary($version, $arch) {
    if ($version -eq "latest") {
        $url = "https://github.com/$Repo/releases/latest/download/${Binary}-windows-${arch}.exe"
    } else {
        $url = "https://github.com/$Repo/releases/download/$version/${Binary}-windows-${arch}.exe"
    }

    $tmpDir = Join-Path $env:TEMP "gndec-qp-install"
    New-Item -ItemType Directory -Force -Path $tmpDir | Out-Null
    $tmpFile = Join-Path $tmpDir "${Binary}.exe"

    Write-Info "Downloading $Binary for windows/$arch..."
    Write-Host "  $url"

    try {
        Invoke-WebRequest -Uri $url -OutFile $tmpFile -UseBasicParsing
    } catch {
        Write-Error "Download failed: $_"
        Write-Error "Try building from source: go install github.com/$Repo@latest"
        Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
        exit 1
    }

    if (-not (Test-Path $tmpFile) -or (Get-Item $tmpFile).Length -eq 0) {
        Write-Error "Downloaded file is empty or missing."
        Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
        exit 1
    }

    Write-Ok "Downloaded successfully"
    return @{ Path = $tmpFile; Dir = $tmpDir }
}

function Install-Binary($fileInfo) {
    $installDir = $null

    $localAppData = [Environment]::GetFolderPath("LocalApplicationData")
    $candidate = Join-Path $localAppData "gndec-qp"
    New-Item -ItemType Directory -Force -Path $candidate | Out-Null

    if (($env:Path -split ';') -contains $candidate) {
        $installDir = $candidate
    } else {
        $userDir = [Environment]::GetFolderPath("UserProfile")
        $candidate2 = Join-Path $userDir ".local\bin"
        New-Item -ItemType Directory -Force -Path $candidate2 | Out-Null
        $installDir = $candidate2
    }

    $destPath = Join-Path $installDir "${Binary}.exe"
    Move-Item -Force $fileInfo.Path $destPath
    Remove-Item -Recurse -Force $fileInfo.Dir -ErrorAction SilentlyContinue

    Write-Ok "Installed to $destPath"

    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -split ';' -notcontains $installDir) {
        $newPath = "$installDir;$currentPath"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        $env:Path = "$installDir;$env:Path"
        Write-Warn "$installDir added to PATH. You may need to restart your terminal."
    }
}

function Verify {
    try {
        $help = & "$Binary" --help 2>&1 | Select-Object -First 1
        Write-Ok "Installation verified"
    } catch {
        Write-Warn "Binary installed but verification failed. Try running 'qp --help' in a new terminal."
    }
}

function Main {
    Write-Host ""
    Write-Host "┌──────────────────────────────────────────┐" -ForegroundColor $Green
    Write-Host "│  GNDEC Question Paper Downloader Install  │" -ForegroundColor $Green
    Write-Host "└──────────────────────────────────────────┘" -ForegroundColor $Green
    Write-Host ""

    $arch = Detect-Arch
    $version = Get-LatestVersion
    Write-Info "Release: $version"

    $fileInfo = Download-Binary $version $arch
    Install-Binary $fileInfo
    Verify

    Write-Host ""
    Write-Ok "Installation complete!"
    Write-Host ""
    Write-Host "  Run 'qp' to launch the interactive TUI"
    Write-Host "  Run 'qp --code PCIT-114' for CLI mode"
    Write-Host "  Run 'qp --code PCIT-114 --auto' to auto-open PDFs"
    Write-Host ""
}

Main
