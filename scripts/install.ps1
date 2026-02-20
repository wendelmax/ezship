# ezship Installer Script
# Usage: iwr -useb https://raw.githubusercontent.com/wendelmax/ezship/main/scripts/install.ps1 | iex

$repo = "wendelmax/ezship"
$installDir = "$HOME\.ezship\bin"

if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

Write-Host "Installing ezship..." -ForegroundColor Cyan

# Detect Architecture
$arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $arch = "arm64"
}
elseif ($env:PROCESSOR_ARCHITECTURE -eq "x86") {
    $arch = "386"
}

$binaryName = "ezship-$arch.exe"
$downloadUrl = "https://github.com/$repo/releases/latest/download/$binaryName"
$targetPath = "$installDir\ezship.exe"

Write-Host "Downloading $binaryName from GitHub..." -ForegroundColor Gray
Invoke-WebRequest -Uri $downloadUrl -OutFile $targetPath

# Update PATH
$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*$installDir*") {
    Write-Host "Adding $installDir to User PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$path;$installDir", "User")
    $env:Path += ";$installDir"
}

Write-Host "ezship installed successfully!" -ForegroundColor Green
Write-Host "Try running: ezship setup docker" -ForegroundColor Gray
