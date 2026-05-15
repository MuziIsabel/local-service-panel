# Build All Script for Local Service Panel
# Run from the project root: .\scripts\build-all.ps1
# Requires: Go, Node.js, npm, Rust (for Tauri)
#
# This script performs a complete build:
#   1. Checks required dependencies
#   2. Installs frontend dependencies
#   3. Builds the Go Agent
#   4. Builds the frontend
#   5. Builds the Tauri MSI installer

param(
    [string]$Version = "0.6.0",
    [switch]$SkipTauri
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$AgentDir = Join-Path $ProjectRoot "agent"
$AppDir = Join-Path $ProjectRoot "app"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Local Service Panel Build v$Version"
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# ---- Dependency checks ----
Write-Host "[1/5] Checking dependencies..." -ForegroundColor Yellow
$hasErrors = $false

# Check Go
try {
    $goVer = go version
    Write-Host "  Go: $goVer" -ForegroundColor Green
} catch {
    Write-Host "  Go: NOT FOUND - install Go 1.22+ from https://go.dev/dl/" -ForegroundColor Red
    $hasErrors = $true
}

# Check Node
try {
    $nodeVer = node --version
    Write-Host "  Node: $nodeVer" -ForegroundColor Green
} catch {
    Write-Host "  Node: NOT FOUND - install Node.js 20+ from https://nodejs.org/" -ForegroundColor Red
    $hasErrors = $true
}

# Check npm
try {
    $npmVer = npm --version
    Write-Host "  npm: v$npmVer" -ForegroundColor Green
} catch {
    Write-Host "  npm: NOT FOUND" -ForegroundColor Red
    $hasErrors = $true
}

# Check Rust (only if building Tauri)
if (-not $SkipTauri) {
    try {
        $rustVer = rustc --version
        Write-Host "  Rust: $rustVer" -ForegroundColor Green
    } catch {
        Write-Host "  Rust: NOT FOUND - install Rust from https://rustup.rs/" -ForegroundColor Red
        $hasErrors = $true
    }
}

if ($hasErrors) {
    Write-Host "`nMissing dependencies. Please install them first." -ForegroundColor Red
    exit 1
}
Write-Host ""

# ---- Frontend dependencies ----
Write-Host "[2/5] Installing frontend dependencies..." -ForegroundColor Yellow
Set-Location $AppDir
npm ci --silent 2>&1 | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Host "  npm install failed" -ForegroundColor Red
    exit 1
}
Write-Host "  Done." -ForegroundColor Green
Write-Host ""

# ---- Build Agent ----
Write-Host "[3/5] Building Go Agent..." -ForegroundColor Yellow
Set-Location $AgentDir
$ldflags = "-X 'github.com/user/local-service-panel/agent/internal/version.Version=$Version'"
go build -ldflags "$ldflags" -o agent.exe ./cmd/agent
if ($LASTEXITCODE -ne 0) {
    Write-Host "  Agent build failed" -ForegroundColor Red
    exit 1
}
Write-Host "  Agent built: $(Join-Path $AgentDir agent.exe)" -ForegroundColor Green
Write-Host ""

# ---- Build Frontend ----
Write-Host "[4/5] Building frontend..." -ForegroundColor Yellow
Set-Location $AppDir
npm run build 2>&1 | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Host "  Frontend build failed" -ForegroundColor Red
    exit 1
}
Write-Host "  Frontend built: $(Join-Path $AppDir dist)" -ForegroundColor Green
Write-Host ""

# ---- Build Tauri MSI ----
if (-not $SkipTauri) {
    Write-Host "[5/5] Building Tauri MSI installer..." -ForegroundColor Yellow
    Set-Location $AppDir
    npx tauri build --bundles msi 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  Tauri build failed" -ForegroundColor Red
        exit 1
    }
    Write-Host "  Tauri MSI built successfully" -ForegroundColor Green
} else {
    Write-Host "[5/5] Skipped (use -SkipTauri to build without MSI)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Build complete!" -ForegroundColor Green
Write-Host "  Agent: $(Join-Path $AgentDir agent.exe)"
Write-Host "  Frontend: $(Join-Path $AppDir dist)"
if (-not $SkipTauri) {
    $msiDir = Join-Path $AppDir "src-tauri\target\release\bundle\msi"
    if (Test-Path $msiDir) {
        Write-Host "  MSI: $(Get-ChildItem $msiDir\*.msi | Select-Object -First 1 -ExpandProperty FullName)"
    }
}
Write-Host "========================================" -ForegroundColor Cyan
