# ezship Build Script
# This script builds the ezship binary and injects the version from the Git tag.

$version = git describe --tags --always --dirty
Write-Host "Building ezship version $version..." -ForegroundColor Gray

if (-not (Test-Path "bin")) { 
    New-Item -ItemType Directory -Path "bin" -Force | Out-Null
}

# Inject the version into wsl.Version using -ldflags
go build -ldflags "-X 'github.com/wendelmax/ezship/internal/wsl.Version=$version'" -o bin/ezship.exe ./cmd/ezship

if ($LASTEXITCODE -eq 0) {
    Write-Host "ezship.exe successfully generated at bin/" -ForegroundColor Green
    Write-Host "Run it with: ./bin/ezship.exe --version" -ForegroundColor Gray
}
else {
    Write-Host "Build failed!" -ForegroundColor Red
}
