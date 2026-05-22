# 安装 LibreOffice 并自动写入 server/config.yaml 的 libreoffice-path
# 请以管理员身份运行 PowerShell:
#   Set-ExecutionPolicy Bypass -Scope Process -Force
#   cd D:\golang\gin-vue-admin
#   .\scripts\install-libreoffice.ps1

$ErrorActionPreference = "Stop"
$version = "25.8.7"
$msiName = "LibreOffice_${version}_Win_x86-64.msi"
$url = "https://download.documentfoundation.org/libreoffice/stable/$version/win/x86_64/$msiName"
$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$destDir = Join-Path $repoRoot "server\libreoffice-installer"
$msiPath = Join-Path $destDir $msiName
$configPath = Join-Path $repoRoot "server\config.yaml"

function Find-Soffice {
    $roots = @(
        ${env:ProgramFiles},
        ${env:ProgramFiles(x86)},
        (Join-Path $env:LOCALAPPDATA "Programs")
    ) | Where-Object { $_ }
    foreach ($root in $roots) {
        $hits = Get-ChildItem -Path $root -Filter "soffice.exe" -Recurse -Depth 5 -ErrorAction SilentlyContinue
        if ($hits) { return $hits[0].FullName }
    }
    return $null
}

$soffice = Find-Soffice
if (-not $soffice) {
    New-Item -ItemType Directory -Force -Path $destDir | Out-Null
    if (-not (Test-Path $msiPath)) {
        Write-Host "下载 LibreOffice $version ..."
        Invoke-WebRequest -Uri $url -OutFile $msiPath -UseBasicParsing
    }
    $isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if (-not $isAdmin) {
        Write-Host "需要管理员权限进行静默安装，正在请求提升..."
        Start-Process powershell.exe -Verb RunAs -ArgumentList "-ExecutionPolicy Bypass -File `"$PSCommandPath`"" -Wait
        $soffice = Find-Soffice
        if (-not $soffice) { Write-Error "安装未完成或未找到 soffice.exe"; exit 1 }
    } else {
        Write-Host "静默安装中..."
        $p = Start-Process msiexec.exe -ArgumentList "/i `"$msiPath`" /qn ALLUSERS=1 CREATEDESKTOPSHORTCUT=0 REGISTER_ALL_MSO_TYPES=0" -Wait -PassThru
        if ($p.ExitCode -ne 0) { Write-Warning "msiexec 退出码 $($p.ExitCode)，尝试打开图形安装程序..."; Start-Process $msiPath -Wait }
        Start-Sleep -Seconds 3
        $soffice = Find-Soffice
    }
}

if (-not $soffice) {
    Write-Error "未找到 soffice.exe。请手动安装 LibreOffice 后重新运行本脚本。"
    exit 1
}

Write-Host "已找到: $soffice"

$escaped = $soffice -replace '\\', '\\'
$yaml = Get-Content $configPath -Raw -Encoding UTF8
if ($yaml -match 'libreoffice-path:\s*".*"') {
    $yaml = $yaml -replace 'libreoffice-path:\s*".*"', "libreoffice-path: `"$escaped`""
} else {
    $yaml = $yaml -replace '(enable-libreoffice:\s*true)', "`$1`n    libreoffice-path: `"$escaped`""
}
Set-Content -Path $configPath -Value $yaml -Encoding UTF8 -NoNewline
Write-Host "已写入 $configPath"
Write-Host "请重启 Go 后端服务。"
