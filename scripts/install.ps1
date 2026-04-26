Param(
    [string]$Version = "",
    [string]$InstallDir = "$HOME\AppData\Local\Programs\gitflip\bin"
)

$ErrorActionPreference = "Stop"

$RepoOwner = "jayesh9747"
$RepoName = "gitflip"

function Get-Arch {
    switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
        "X64" { return "amd64" }
        "Arm64" { return "arm64" }
        default { throw "Unsupported architecture: $($_.ToString())" }
    }
}

if (-not $Version) {
    $latest = Invoke-RestMethod -Uri "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
    if (-not $latest.tag_name) {
        throw "Could not determine latest release tag."
    }
    $Version = $latest.tag_name
}

$arch = Get-Arch
$versionNumber = $Version.TrimStart("v")
$asset = "gitflip_${versionNumber}_windows_${arch}.zip"
$url = "https://github.com/$RepoOwner/$RepoName/releases/download/$Version/$asset"

$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("gitflip-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tempDir | Out-Null

try {
    $zipPath = Join-Path $tempDir $asset
    Invoke-WebRequest -Uri $url -OutFile $zipPath

    $extractDir = Join-Path $tempDir "extract"
    Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force

    $binary = Join-Path $extractDir "gitflip.exe"
    if (-not (Test-Path $binary)) {
        throw "Expected gitflip.exe inside release archive."
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item $binary (Join-Path $InstallDir "gitflip.exe") -Force

    Write-Host "Installed gitflip $Version -> $InstallDir\gitflip.exe"
    Write-Host "Add $InstallDir to PATH if it is not already there."
} finally {
    if (Test-Path $tempDir) {
        Remove-Item $tempDir -Recurse -Force
    }
}
