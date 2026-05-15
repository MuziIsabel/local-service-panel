# Install Local Service Panel Agent as a Windows Service.
# Run this script as Administrator.
#
# Usage:
#   .\install.ps1 [-AgentPath <path>] [-DataDir <path>]
#
# Defaults:
#   AgentPath: .\agent.exe (sibling to this script)
#   DataDir:   C:\ProgramData\LocalServicePanel

param(
    [string]$AgentPath = "",
    [string]$DataDir = ""
)

#Requires -RunAsAdministrator

$ErrorActionPreference = "Stop"

# Default paths
if (-not $AgentPath) {
    $AgentPath = Join-Path -Path $PSScriptRoot -ChildPath "agent.exe"
}
if (-not $DataDir) {
    $DataDir = "C:\ProgramData\LocalServicePanel"
}

$AgentDest = Join-Path -Path $DataDir -ChildPath "agent.exe"

Write-Host "=== Local Service Panel Agent Installer ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Agent source: $AgentPath"
Write-Host "Data directory: $DataDir"
Write-Host ""

# 1. Create data directory structure
Write-Host "[1/4] Creating data directory structure..." -ForegroundColor Yellow
$null = New-Item -ItemType Directory -Path (Join-Path $DataDir "config") -Force
$null = New-Item -ItemType Directory -Path (Join-Path $DataDir "data") -Force
$null = New-Item -ItemType Directory -Path (Join-Path $DataDir "logs") -Force
$null = New-Item -ItemType Directory -Path (Join-Path $DataDir "logs\apps") -Force
Write-Host "  Done."

# 2. Copy agent binary
Write-Host "[2/4] Copying agent binary..." -ForegroundColor Yellow
if (-not (Test-Path $AgentPath)) {
    Write-Error "Agent binary not found at: $AgentPath"
    exit 1
}
Copy-Item -Path $AgentPath -Destination $AgentDest -Force
Write-Host "  Copied to: $AgentDest"

# 3. Register Windows Service
Write-Host "[3/4] Registering Windows Service..." -ForegroundColor Yellow
$existing = Get-Service -Name "LocalServicePanelAgent" -ErrorAction SilentlyContinue
if ($existing) {
    Write-Host "  Service already exists, stopping and removing..."
    Stop-Service -Name "LocalServicePanelAgent" -Force -ErrorAction SilentlyContinue
    & $AgentDest -service uninstall
    Start-Sleep -Seconds 2
}

& $AgentDest -data $DataDir -service install
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to install service."
    exit 1
}
Write-Host "  Service registered."

# 4. Start the service
Write-Host "[4/4] Starting service..." -ForegroundColor Yellow
Start-Service -Name "LocalServicePanelAgent"
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to start service."
    exit 1
}

Write-Host ""
Write-Host "=== Installation complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Agent is running as service 'LocalServicePanelAgent'."
Write-Host "Data directory: $DataDir"
Write-Host ""

# Verify
Start-Sleep -Seconds 2
try {
    $health = Invoke-RestMethod -Uri "http://127.0.0.1:17645/api/healthz" -ErrorAction Stop
    Write-Host "Health check: $($health.data.status) (v$($health.data.version))" -ForegroundColor Green
} catch {
    Write-Warning "Health check failed. The service may still be starting."
    Write-Warning "Run 'Get-Service LocalServicePanelAgent' to check status."
}
