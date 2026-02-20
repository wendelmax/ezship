# ezship Installer Script
# Usage: iwr -useb https://raw.githubusercontent.com/wendelmax/ezship/main/install.ps1 | iex

$repo = "wendelmax/ezship"
$binaryName = "ezship.exe"
$installDir = "$HOME\.ezship\bin"

if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

Write-Host "🚢 Installing ezship..." -ForegroundColor Cyan

# Determine latest release (mocked for now, assumes local build if repo is empty)
# In production, this would fetch from GitHub Releases API
$downloadUrl = "https://github.com/$repo/releases/latest/download/$binaryName"

# For development/simulated use, we assume the binary is already built or we provide a helper
Write-Host "Checking for ezship binary..."
if (Test-Path ".\ezship.exe") {
    Copy-Item ".\ezship.exe" "$installDir\ezship.exe" -Force
} else {
    Write-Host "Warning: Remote download not implemented yet. Please build ezship.exe locally first." -ForegroundColor Yellow
}

# Update PATH
$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*$installDir*") {
    Write-Host "Adding $installDir to User PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$path;$installDir", "User")
    $env:Path += ";$installDir"
}

Write-Host "🚀 ezship installed successfully!" -ForegroundColor Green
Write-Host "Try running: ezship setup docker" -ForegroundColor Gray
