<#
.SYNOPSIS
    DeepScanBot CLI Install Script for Windows
.DESCRIPTION
    Detects OS/Architecture, fetches the latest binary from GitHub Releases,
    verifies SHA256 checksum, and installs to a standard system path.
.PARAMETER InstallDir
    Installation directory (default: $env:ProgramFiles\DeepScanBot)
.PARAMETER Version
    Version to install (default: latest)
.EXAMPLE
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    iex ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/mindfiredigital/DeepScanBot/main/scripts/install.ps1'))
.EXAMPLE
    .\install.ps1 -InstallDir "C:\tools" -Version "v1.0.0"
#>

param(
    [string]$InstallDir = "",
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
$RepoOwner = "mindfiredigital"
$RepoName = "DeepScanBot"
$ProjectName = "deepscanbot"
$GitHubApi = "https://api.github.com"
$GitHubDl = "https://github.com"

# Default install directory
if ([string]::IsNullOrEmpty($InstallDir)) {
    $InstallDir = [System.IO.Path]::Combine($env:ProgramFiles, "DeepScanBot")
}

# ---------------------------------------------------------------------------
# Helper functions
# ---------------------------------------------------------------------------
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Ok {
    param([string]$Message)
    Write-Host "[OK]   $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Err {
    param([string]$Message)
    Write-Host "[ERR]  $Message" -ForegroundColor Red
    exit 1
}

# ---------------------------------------------------------------------------
# Detect OS and architecture
# ---------------------------------------------------------------------------
$OS = "windows"
$Arch = $env:PROCESSOR_ARCHITECTURE.ToLower()
switch ($Arch) {
    "amd64" { $Arch = "amd64" }
    "x86_64" { $Arch = "amd64" }
    "arm64" { $Arch = "arm64" }
    default { Write-Err "Unsupported architecture: $Arch" }
}

Write-Info "Detected OS: ${OS}, Architecture: ${Arch}"

# ---------------------------------------------------------------------------
# Determine version
# ---------------------------------------------------------------------------
if ([string]::IsNullOrEmpty($Version)) {
    Write-Info "Fetching latest release version..."
    try {
        $releaseUrl = "${GitHubApi}/repos/${RepoOwner}/${RepoName}/releases/latest"
        $release = Invoke-RestMethod -Uri $releaseUrl -Headers @{"Accept"="application/vnd.github.v3+json"}
        $Version = $release.tag_name
    }
    catch {
        Write-Err "Failed to fetch latest version. Check network or rate limits: $_"
    }
    Write-Info "Latest release: ${Version}"
}
else {
    # Ensure version has 'v' prefix for tag matching
    if (-not $Version.StartsWith("v")) {
        $Version = "v${Version}"
    }
    Write-Info "Installing version: ${Version}"
}

# ---------------------------------------------------------------------------
# Build binary name and download URLs
# ---------------------------------------------------------------------------
$BinaryName = "${ProjectName}_${OS}_${Arch}.exe"
$ChecksumFile = "checksums.txt"
$BinaryUrl = "${GitHubDl}/${RepoOwner}/${RepoName}/releases/download/${Version}/${BinaryName}"
$ChecksumUrl = "${GitHubDl}/${RepoOwner}/${RepoName}/releases/download/${Version}/${ChecksumFile}"

$TmpDir = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), [System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $TmpDir -Force | Out-Null

try {
    # -----------------------------------------------------------------------
    # Download binary
    # -----------------------------------------------------------------------
    Write-Info "Downloading ${BinaryName}..."
    $BinaryPath = [System.IO.Path]::Combine($TmpDir, $BinaryName)
    try {
        Invoke-WebRequest -Uri $BinaryUrl -OutFile $BinaryPath -UseBasicParsing
    }
    catch {
        Write-Err "Failed to download binary from ${BinaryUrl}: $_"
    }
    Write-Ok "Binary downloaded successfully"

    # -----------------------------------------------------------------------
    # Download checksums
    # -----------------------------------------------------------------------
    Write-Info "Downloading checksums..."
    $ChecksumPath = [System.IO.Path]::Combine($TmpDir, $ChecksumFile)
    try {
        Invoke-WebRequest -Uri $ChecksumUrl -OutFile $ChecksumPath -UseBasicParsing
    }
    catch {
        Write-Err "Failed to download checksums from ${ChecksumUrl}: $_"
    }
    Write-Ok "Checksums downloaded successfully"

    # -----------------------------------------------------------------------
    # Verify SHA256 checksum
    # -----------------------------------------------------------------------
    Write-Info "Verifying SHA256 checksum..."

    $checksums = Get-Content $ChecksumPath
    $expectedHash = $null
    foreach ($line in $checksums) {
        if ($line -match "^([a-fA-F0-9]+)\s+\*?${BinaryName}$") {
            $expectedHash = $matches[1].ToLower()
            break
        }
    }

    if ([string]::IsNullOrEmpty($expectedHash)) {
        Write-Err "Binary name '${BinaryName}' not found in checksums.txt. Ensure version matches."
    }

    $computedHash = (Get-FileHash -Path $BinaryPath -Algorithm SHA256).Hash.ToLower()

    if ($expectedHash -ne $computedHash) {
        Write-Err "Checksum mismatch! Expected: ${expectedHash}, Computed: ${computedHash}"
    }
    Write-Ok "SHA256 checksum verified successfully"

    # -----------------------------------------------------------------------
    # Install binary
    # -----------------------------------------------------------------------
    if (-not (Test-Path $InstallDir)) {
        Write-Info "Creating installation directory: ${InstallDir}"
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $TargetPath = [System.IO.Path]::Combine($InstallDir, "${ProjectName}.exe")
    if (Test-Path $TargetPath) {
        Write-Warn "Overwriting existing binary at ${TargetPath}"
    }

    Move-Item -Path $BinaryPath -Destination $TargetPath -Force
    Write-Ok "DeepScanBot ${Version} installed successfully to ${TargetPath}"

    # -----------------------------------------------------------------------
    # Verify installation
    # -----------------------------------------------------------------------
    try {
        $output = & $TargetPath version
        Write-Ok "Installation verified: $output"
    }
    catch {
        Write-Warn "Binary installed but verification failed. Check PATH and permissions."
    }

    # -----------------------------------------------------------------------
    # PATH update reminder
    # -----------------------------------------------------------------------
    $currentPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User)
    if ($currentPath -notlike "*${InstallDir}*") {
        Write-Warn "${InstallDir} is not in your PATH."
        Write-Warn "Add it by running:"
        Write-Warn "  [Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', [EnvironmentVariableTarget]::User) + ';${InstallDir}', [EnvironmentVariableTarget]::User)"
        Write-Warn "Then restart your terminal."
    }
    else {
        Write-Ok "${InstallDir} is already in your PATH"
    }

    Write-Info "Installation complete! Run 'deepscanbot --help' to get started."
}
finally {
    # Cleanup
    if (Test-Path $TmpDir) {
        Remove-Item -Path $TmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}