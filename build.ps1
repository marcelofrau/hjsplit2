param(
    [ValidateSet('win32', 'win64', 'linux', 'all', 'native')]
    [string]$Platform = 'all'
)

# ── Config ──────────────────────────────────────────────────────────────────
$config = @{}
if (Test-Path 'build.conf') {
    Get-Content 'build.conf' | ForEach-Object {
        if ($_ -match '^([A-Za-z_][A-Za-z0-9_]*)=(.*)') {
            $config[$matches[1]] = $matches[2].Trim('"').Trim("'")
        }
    }
}

# ── Version ─────────────────────────────────────────────────────────────────
if (-not (Test-Path 'version.txt')) {
    Write-Host "Error: version.txt not found" -ForegroundColor Red
    exit 1
}
$Version = (Get-Content 'version.txt' -Raw).Trim()

$OutputDir = 'dist'
$Package = '.'
$LdFlags = "-s -w -H windowsgui -X hjsplit2/internal/version.Version=$Version"

# Clean dist
if (Test-Path $OutputDir) { Remove-Item -Recurse -Force $OutputDir }
New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

$global:anyOk = $false

# ── Build helper ────────────────────────────────────────────────────────────
function Build-Target {
    param([string]$GOOS, [string]$GOARCH, [string]$Suffix)

    $name = "hjsplit2-v$Version-$GOOS-$GOARCH$Suffix"
    $output = Join-Path $OutputDir $name

    Write-Host "==> Building: $name" -ForegroundColor Cyan

    # Generate architecture-specific .syso for executable icon
    $rcFile = 'appicon.rc'
    $sysoFile = 'appicon.syso'
    $windres = (Get-Command 'windres' -ErrorAction SilentlyContinue).Source
    if ($windres -and $GOOS -eq 'windows') {
        $targetFlag = if ($GOARCH -eq '386') { 'pe-i386' } else { 'pe-x86-64' }
        & $windres -F $targetFlag -o $sysoFile $rcFile 2>&1 | Out-Null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "   Icon resource: $sysoFile ($targetFlag)" -ForegroundColor Magenta
        }
    }

    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    $env:CGO_ENABLED = '1'

    # Set CC per architecture (prevents env pollution from 32-bit compiler)
    if ($GOOS -eq 'windows') {
        if ($GOARCH -eq '386') {
            if (-not $cc32) {
                Write-Host "    32-bit compiler not configured" -ForegroundColor Yellow
                return
            }
            $env:CC = $cc32
        } else {
            $env:CC = (Get-Command 'gcc' -ErrorAction SilentlyContinue).Source
        }
    }

    $result = go build -ldflags="$LdFlags" -o $output $Package 2>&1

    if ($LASTEXITCODE -eq 0) {
        $sizeKB = [math]::Round((Get-Item $output).Length / 1KB)
        Write-Host "    OK: $output ($sizeKB KB)" -ForegroundColor Green
        $global:anyOk = $true

        # UPX compression if configured
        if ($config['USE_UPX'] -eq 'true') {
            $upx = (Get-Command 'upx' -ErrorAction SilentlyContinue).Source
            if ($upx) {
                Write-Host "    Compressing with UPX..." -ForegroundColor Cyan
                & $upx --best --no-color --no-progress $output 2>&1 | Out-Null
                $compKB = [math]::Round((Get-Item $output).Length / 1KB)
                Write-Host "    Compressed: $compKB KB" -ForegroundColor Green
            } else {
                Write-Host "    UPX not found. Install: scoop install upx" -ForegroundColor Yellow
            }
        }
    } else {
        Write-Host "    FAILED" -ForegroundColor Red
        Write-Host "    $($result -join "`n    ")"
        if (Test-Path $output) { Remove-Item $output -Force }
    }

    # Clean up .syso (generated per-architecture)
    if (Test-Path $sysoFile) { Remove-Item $sysoFile -Force }
}

# ── Detect available compilers ──────────────────────────────────────────────
$cc64 = (gcc -dumpmachine 2>$null) -replace '\s',''
Write-Host "Host compiler (64-bit): $cc64" -ForegroundColor Magenta

$cc32 = ''
if ($config['CC_386']) {
    $cc32 = $config['CC_386']
    Write-Host "32-bit compiler (from config): $cc32" -ForegroundColor Magenta
} elseif ($config['MINGW32_PATH']) {
    $candidate = Join-Path $config['MINGW32_PATH'] 'bin\gcc.exe'
    if (Test-Path $candidate) {
        $cc32 = $candidate
        Write-Host "32-bit compiler (from config path): $cc32" -ForegroundColor Magenta
    }
}
if (-not $cc32) {
    $candidate = (Get-Command 'i686-w64-mingw32-gcc' -ErrorAction SilentlyContinue).Source
    if ($candidate) { $cc32 = $candidate; Write-Host "32-bit compiler (from PATH): $cc32" -ForegroundColor Magenta }
}

# WSL
$wslDistro = ''
if ($config['WSL_DISTRO']) { $wslDistro = $config['WSL_DISTRO'] }
$wslExe = (Get-Command 'wsl.exe' -ErrorAction SilentlyContinue)
$wslAvailable = $wslExe -ne $null
if ($wslAvailable) { Write-Host "WSL available" -ForegroundColor Magenta }

# Linux cross-compile via WSL
function Build-LinuxViaWSL {
    $wslArgs = @()
    if ($wslDistro) { $wslArgs += '-d', $wslDistro }

    Write-Host "Checking Go in WSL..." -ForegroundColor Cyan
    $checkArgs = $wslArgs + @('go', 'version')
    $goCheck = (& 'wsl.exe' $checkArgs 2>&1 | Out-String).Trim()
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Go not found in WSL." -ForegroundColor Yellow
        Write-Host "  Install: wsl sudo apt update && wsl sudo apt install golang-go" -ForegroundColor Yellow
        return
    }

    $wslRoot = "/mnt/c/Users/$env:USERNAME/workspace/hjsplit2"
    $buildSh = "${wslRoot}/build.sh"
    $buildArgs = $wslArgs + @($buildSh, 'linux')

    Write-Host "Building Linux via WSL..." -ForegroundColor Cyan
    $result = & 'wsl.exe' $buildArgs 2>&1
    $exitCode = $LASTEXITCODE

    # Always show WSL output so user can see errors
    if ($result) {
        Write-Host "$result" -ForegroundColor DarkYellow
    }

    # Find the linux binary that should have been produced
    $linuxName = "hjsplit2-v$Version-linux-amd64"
    $linuxBinary = Join-Path $OutputDir $linuxName
    $binaryOk = (Test-Path $linuxBinary) -and ((Get-Item $linuxBinary).Length -gt 0)

    if ($exitCode -eq 0 -and $binaryOk) {
        $sizeKB = [math]::Round((Get-Item $linuxBinary).Length / 1KB)
        Write-Host "    OK: $linuxBinary ($sizeKB KB)" -ForegroundColor Green
        $global:anyOk = $true
    } else {
        Write-Host "    FAILED via WSL" -ForegroundColor Red
        if (-not $binaryOk -and (Test-Path $linuxBinary)) {
            Remove-Item $linuxBinary -Force
        }
    }
}

Write-Host ""

# ── Build ───────────────────────────────────────────────────────────────────
switch ($Platform) {
    'native' {
        Build-Target -GOOS 'windows' -GOARCH $env:GOARCH -Suffix '.exe'
    }
    'win32' {
        if ($cc32) {
            Build-Target -GOOS 'windows' -GOARCH '386' -Suffix '.exe'
        } else {
            Write-Host "win32: no 32-bit MinGW configured." -ForegroundColor Yellow
            Write-Host "  Set MINGW32_PATH or CC_386 in build.conf" -ForegroundColor Yellow
            Write-Host "  Example: MINGW32_PATH=C:/Apps/mingw32" -ForegroundColor Yellow
        }
    }
    'win64' {
        if ($cc64) { Build-Target -GOOS 'windows' -GOARCH 'amd64' -Suffix '.exe' }
        else { Write-Host "win64: no MinGW found on PATH" -ForegroundColor Yellow }
    }
    'linux' {
        if ($wslAvailable) { Build-LinuxViaWSL }
        else { Write-Host "Linux build requires WSL. Install: wsl --install" -ForegroundColor Yellow }
    }
    'all' {
        Write-Host "--- Building all platforms ---" -ForegroundColor Cyan

        # Linux first (WSL build.sh cleans dist)
        if ($wslAvailable) { Build-LinuxViaWSL }
        else { Write-Host "Skipping linux (WSL not available)" -ForegroundColor DarkYellow }

        # Then Windows
        if ($cc64) { Build-Target -GOOS 'windows' -GOARCH 'amd64' -Suffix '.exe' }
        else { Write-Host "Skipping win64 (no MinGW)" -ForegroundColor DarkYellow }

        if ($cc32) {
            Build-Target -GOOS 'windows' -GOARCH '386' -Suffix '.exe'
        } else {
            Write-Host "Skipping win32 (configure MINGW32_PATH in build.conf)" -ForegroundColor DarkYellow
        }
    }
}

if (-not $global:anyOk) {
    Write-Host "`nNo builds succeeded." -ForegroundColor Red
    exit 1
}
