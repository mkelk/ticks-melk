# Ticks installer for Windows
# Usage: irm https://raw.githubusercontent.com/mkelk/ticks-melk/main-melk/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "mkelk/ticks-melk"
$Binary = "tk"
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\ticks" }

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Error "32-bit systems are not supported"
    exit 1
}

# Get latest version
Write-Host "Fetching latest version..."
$Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
$Version = $Release.tag_name -replace "^v", ""

Write-Host "Installing tk v$Version for windows/$Arch..."

# Download URL
$Url = "https://github.com/$Repo/releases/download/v$Version/${Binary}_${Version}_windows_${Arch}.tar.gz"

# Create install directory
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

# Download and extract
$TmpFile = Join-Path $env:TEMP "tk.tar.gz"
$TmpDir = Join-Path $env:TEMP "tk-extract"

Invoke-WebRequest -Uri $Url -OutFile $TmpFile

# Extract tar.gz (requires tar, available in Windows 10+)
New-Item -ItemType Directory -Force -Path $TmpDir | Out-Null
tar -xzf $TmpFile -C $TmpDir

# Move binary
Move-Item -Force "$TmpDir\$Binary.exe" "$InstallDir\$Binary.exe"

# Cleanup
Remove-Item -Force $TmpFile
Remove-Item -Recurse -Force $TmpDir

Write-Host "Installed tk to $InstallDir\$Binary.exe"

# Check if in PATH
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($CurrentPath -notlike "*$InstallDir*") {
    Write-Host ""
    Write-Host "Add to your PATH by running:"
    Write-Host "  `$env:Path += `";$InstallDir`""
    Write-Host ""
    Write-Host "Or permanently:"
    Write-Host "  [Environment]::SetEnvironmentVariable('Path', `$env:Path + ';$InstallDir', 'User')"
}

Write-Host ""
Write-Host "Run 'tk version' to verify installation"
