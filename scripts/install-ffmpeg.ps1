# Install FFmpeg to server/tools/ffmpeg for office media tools
# Run from repo root: .\scripts\install-ffmpeg.ps1

$ErrorActionPreference = "Stop"
$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$targetDir = Join-Path $repoRoot "server\tools\ffmpeg"
$binPath = Join-Path $targetDir "bin\ffmpeg.exe"
$configPath = Join-Path $repoRoot "server\config.yaml"

if (Test-Path $binPath) {
    Write-Host "FFmpeg exists: $binPath"
} else {
    $version = "8.1.1"
    $zipName = "ffmpeg-$version-essentials_build.zip"
    $urls = @(
        "https://github.com/GyanD/codexffmpeg/releases/download/$version/$zipName",
        "https://ghproxy.com/https://github.com/GyanD/codexffmpeg/releases/download/$version/$zipName"
    )
    $cacheDir = Join-Path $repoRoot "server\tools\ffmpeg-download"
    $zipPath = Join-Path $cacheDir $zipName
    New-Item -ItemType Directory -Force -Path $cacheDir | Out-Null

    Write-Host "Downloading FFmpeg $version ..."
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    $ok = $false
    foreach ($url in $urls) {
        try {
            Write-Host "Try: $url"
            Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing -TimeoutSec 120
            $ok = $true
            break
        } catch {
            Write-Host "Failed: $_"
        }
    }
    if (-not $ok) { throw "All download mirrors failed" }

    Write-Host "Extracting to server/tools ..."
    if (Test-Path $targetDir) { Remove-Item -Recurse -Force $targetDir }
    Expand-Archive -Path $zipPath -DestinationPath (Join-Path $repoRoot "server\tools") -Force
    $extracted = Get-ChildItem (Join-Path $repoRoot "server\tools") -Directory | Where-Object { $_.Name -like "ffmpeg-*" } | Select-Object -First 1
    if ($extracted) {
        Rename-Item $extracted.FullName $targetDir -Force
    }
    if (-not (Test-Path $binPath)) {
        $found = Get-ChildItem $targetDir -Recurse -Filter "ffmpeg.exe" | Select-Object -First 1
        if ($found) {
            $realBin = Join-Path $targetDir "bin"
            New-Item -ItemType Directory -Force -Path $realBin | Out-Null
            Copy-Item $found.FullName (Join-Path $realBin "ffmpeg.exe") -Force
            $ffprobe = Join-Path $found.DirectoryName "ffprobe.exe"
            if (Test-Path $ffprobe) {
                Copy-Item $ffprobe (Join-Path $realBin "ffprobe.exe") -Force
            }
        }
    }
}

if (-not (Test-Path $binPath)) {
    Write-Error "Install failed: $binPath not found"
    exit 1
}

Write-Host "OK: $binPath"

$escaped = $binPath -replace '\\', '\\'
$yaml = Get-Content $configPath -Raw -Encoding UTF8
if ($yaml -match 'ffmpeg-path:\s*".*"') {
    $yaml = $yaml -replace 'ffmpeg-path:\s*".*"', "ffmpeg-path: `"$escaped`""
} else {
    $yaml = $yaml -replace 'ffmpeg-path:\s*""', "ffmpeg-path: `"$escaped`""
}
Set-Content -Path $configPath -Value $yaml -Encoding UTF8 -NoNewline
Write-Host "Updated $configPath - restart Go server."
