# LinkAI CLI installer for Windows (PowerShell).
# Downloads the pre-built linkai.exe and installs the agent skill. No Go / Node.js required.
#
# Quick install:
#   irm https://cdn.link-ai.tech/cli/install.ps1 | iex
#   # or from GitHub:
#   irm https://raw.githubusercontent.com/MinimalFuture/linkai-cli/main/install.ps1 | iex
#
# Environment overrides (all optional):
#   $env:LINKAI_VERSION      version to install            (default: latest)
#   $env:LINKAI_INSTALL_DIR  where to put linkai.exe        (default: %LOCALAPPDATA%\linkai\bin)
#   $env:LINKAI_SOURCE       download source: cdn | github  (default: cdn, GitHub fallback)
#   $env:LINKAI_NO_SKILL     set to 1 to skip the agent skill

$ErrorActionPreference = 'Stop'

$Repo       = 'MinimalFuture/linkai-cli'
$Binary     = 'linkai'
$CdnBase    = 'https://cdn.link-ai.tech/cli'
$GithubBase = "https://github.com/$Repo/releases/download"

$Version   = if ($env:LINKAI_VERSION) { $env:LINKAI_VERSION } else { 'latest' }
$Source    = if ($env:LINKAI_SOURCE)  { $env:LINKAI_SOURCE }  else { 'cdn' }
$NoSkill   = $env:LINKAI_NO_SKILL -eq '1'
$InstallDir = if ($env:LINKAI_INSTALL_DIR) { $env:LINKAI_INSTALL_DIR } else { "$env:LOCALAPPDATA\linkai\bin" }

function Write-Info { param($m) Write-Host "  $m" }
function Write-Ok   { param($m) Write-Host "  [ok] $m" -ForegroundColor Green }
function Write-Err  { param($m) Write-Host "  [x] $m" -ForegroundColor Red; exit 1 }

function Test-Url {
  param($Url)
  try {
    Invoke-WebRequest -Uri $Url -Method Head -TimeoutSec 12 -UseBasicParsing | Out-Null
    return $true
  } catch { return $false }
}

# CDN primary, GitHub fallback (unless explicitly pinned).
function Resolve-Source {
  switch ($Source) {
    'github' { return }
    'cdn' {
      if (Test-Url "$CdnBase/install.ps1") { return }
      Write-Info 'CDN unreachable, falling back to GitHub Releases'
      $script:Source = 'github'
    }
    default { Write-Err "invalid LINKAI_SOURCE='$Source' (use 'cdn' or 'github')" }
  }
}

function Resolve-Version {
  if ($script:Version -ne 'latest') { return }
  if ($script:Source -eq 'cdn') {
    try {
      $v = (Invoke-WebRequest -Uri "$CdnBase/latest.txt" -TimeoutSec 12 -UseBasicParsing).Content.Trim()
      if ($v) { $script:Version = $v; return }
    } catch {}
  }
  # GitHub redirect trick.
  try {
    $resp = Invoke-WebRequest -Uri "https://github.com/$Repo/releases/latest" -MaximumRedirection 0 -ErrorAction SilentlyContinue -UseBasicParsing
  } catch {
    $loc = $_.Exception.Response.Headers.Location
    if ($loc) { $script:Version = ($loc.ToString() -split '/tag/')[-1] }
  }
  if ($script:Version -eq 'latest' -or -not $script:Version) {
    Write-Err 'could not resolve the latest version — set $env:LINKAI_VERSION'
  }
  # Normalize to a bare version (no leading 'v'); tags, CDN/GitHub paths and
  # archive names are all versioned without the 'v'.
  $script:Version = $script:Version -replace '^v', ''
}

function Get-AssetUrl {
  param($Name)
  if ($script:Source -eq 'cdn') { return "$CdnBase/$($script:Version)/$Name" }
  return "$GithubBase/$($script:Version)/$Name"
}

function Install-Binary {
  Resolve-Version

  $arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
  } else { 'amd64' }

  $archive = "$($Binary)_$($script:Version)_windows_$arch.zip"
  $url     = Get-AssetUrl $archive

  $tmp = Join-Path $env:TEMP ("linkai-" + [System.Guid]::NewGuid().ToString('N'))
  New-Item -ItemType Directory -Path $tmp -Force | Out-Null
  try {
    Write-Info "==> Downloading $Binary $($script:Version) (windows/$arch) from $($script:Source)"
    $zipPath = Join-Path $tmp $archive
    Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing

    # Verify checksum when available.
    try {
      $sumPath = Join-Path $tmp 'checksums.txt'
      Invoke-WebRequest -Uri (Get-AssetUrl 'checksums.txt') -OutFile $sumPath -UseBasicParsing
      $line = Select-String -Path $sumPath -Pattern ([regex]::Escape($archive)) | Select-Object -First 1
      if ($line) {
        $expected = ($line.Line -split '\s+')[0]
        $actual   = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash.ToLower()
        if ($expected.ToLower() -ne $actual) { Write-Err "checksum mismatch (expected $expected, got $actual)" }
        Write-Ok 'SHA256 checksum verified'
      }
    } catch {}

    Write-Info '==> Extracting'
    Expand-Archive -Path $zipPath -DestinationPath $tmp -Force

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    $exeSrc = Get-ChildItem -Path $tmp -Recurse -Filter "$Binary.exe" | Select-Object -First 1
    if (-not $exeSrc) { Write-Err "linkai.exe not found in the downloaded archive" }
    Copy-Item -Path $exeSrc.FullName -Destination (Join-Path $InstallDir "$Binary.exe") -Force

    Write-Ok "Binary installed: $InstallDir\$Binary.exe ($($script:Version))"
    Add-ToUserPath $InstallDir
  } finally {
    Remove-Item -Path $tmp -Recurse -Force -ErrorAction SilentlyContinue
  }
}

# Persist the install dir on the user PATH (idempotent) and update the current session.
function Add-ToUserPath {
  param($Dir)
  $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
  if ($userPath -split ';' -contains $Dir) {
    return  # already present
  }
  $newPath = if ([string]::IsNullOrEmpty($userPath)) { $Dir } else { "$userPath;$Dir" }
  [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
  $env:Path = "$env:Path;$Dir"  # current session
  Write-Ok "Added $Dir to your user PATH"
  Write-Info '  Open a new terminal for the PATH change to take effect everywhere.'
}

# The skill (SKILL.md + references/) ships embedded in the binary, so we drive
# `linkai skill install` rather than downloading a separate archive. This keeps
# the installed skill in lockstep with the binary version — no CDN round-trip.
function Install-Skill {
  # Prefer the binary we just installed; otherwise fall back to a linkai.exe
  # already on PATH.
  $bin = Join-Path $InstallDir "$Binary.exe"
  if (-not (Test-Path $bin)) {
    $cmd = Get-Command $Binary -ErrorAction SilentlyContinue
    if ($cmd) { $bin = $cmd.Source }
    else {
      Write-Info "'$Binary' not found; install the binary first, then run '$Binary skill install'."
      return
    }
  }

  Write-Info '==> Installing agent skill'
  # `skill install` prints its own per-destination lines. A non-zero exit (e.g.
  # a build without the embedded skill) is non-fatal.
  & $bin skill install
  if ($LASTEXITCODE -ne 0) {
    Write-Info 'Skill install skipped (this binary may not embed the skill).'
  }
}

Write-Host ''
Write-Info 'LinkAI CLI installer'
Write-Host ''

Resolve-Source
Install-Binary
if (-not $NoSkill) { Install-Skill }

Write-Host ''
Write-Ok 'Done!'
Write-Info ''
Write-Info 'Next steps:'
Write-Info '  linkai auth login     # authenticate with LinkAI'
Write-Info '  linkai --help         # explore commands'
Write-Host ''
