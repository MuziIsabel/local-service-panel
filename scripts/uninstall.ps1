# Uninstall Local Service Panel Agent Windows Service.
# Run this script as Administrator.
#
# Usage:
#   .\uninstall.ps1 [-DataDir <path>] [-CleanData]
#
# Defaults:
#   DataDir: C:\ProgramData\LocalServicePanel

param(
    [string]$DataDir = "",
    [switch]$CleanData
)

#Requires -RunAsAdministrator

$ErrorActionPreference = "Stop"

if (-not $DataDir) {
    $DataDir = "C:\ProgramData\LocalServicePanel"
}

$AgentPath = Join-Path -Path $DataDir -ChildPath "agent.exe"

Write-Host "=== Local Service Panel Agent Uninstaller ===" -ForegroundColor Cyan
Write-Host ""

# 1. Stop the service
Write-Host "[1/3] Stopping service..." -ForegroundColor Yellow
$svc = Get-Service -Name "LocalServicePanelAgent" -ErrorAction SilentlyContinue
if ($svc -and $svc.Status -eq "Running") {
    Stop-Service -Name "LocalServicePanelAgent" -Force
    Write-Host "  Service stopped."
} else {
    Write-Host "  Service is not running."
}

# 2. Unregister the service
Write-Host "[2/3] Unregistering service..." -ForegroundColor Yellow
if (Test-Path $AgentPath) {
    & $AgentPath -data $DataDir -service uninstall
    if ($LASTEXITCODE -ne 0) {
        Write-Warning "Service uninstall command returned non-zero exit code."
    }
    Write-Host "  Service unregistered."
} else {
    Write-Warning "Agent binary not found at $AgentPath. Attempting direct SCM removal..."
    sc.exe delete "LocalServicePanelAgent"
    Write-Host "  Done."
}

# 3. Optional: clean data directory
Write-Host "[3/3] Cleaning up..." -ForegroundColor Yellow
if ($CleanData) {
    Write-Host "  Removing data directory: $DataDir"
    Remove-Item -Path $DataDir -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host "  Data directory removed."
} else {
    Write-Host "  Keeping data directory: $DataDir"
    Write-Host "  Use -CleanData to remove it."
}

Write-Host ""
Write-Host "=== Uninstallation complete ===" -ForegroundColor Green
