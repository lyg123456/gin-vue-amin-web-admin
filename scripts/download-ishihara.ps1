# Download Ishihara plates to web/public/portal/ishihara
# Usage: .\scripts\download-ishihara.ps1

$repoRoot = Split-Path -Parent $PSScriptRoot
$destDir = Join-Path $repoRoot 'web\public\portal\ishihara'
New-Item -ItemType Directory -Force -Path $destDir | Out-Null

$plateUrls = @{
    '01.png' = 'https://upload.wikimedia.org/wikipedia/commons/thumb/4/4b/Ishihara_1.PNG/500px-Ishihara_1.PNG'
    '02.png' = 'https://upload.wikimedia.org/wikipedia/commons/6/68/47-rg12.jpg'
    '03.jpg' = 'https://upload.wikimedia.org/wikipedia/commons/7/76/Ishihara_3.jpg'
    '04.png' = 'https://upload.wikimedia.org/wikipedia/commons/thumb/e/e0/Ishihara_9.png/500px-Ishihara_9.png'
    '05.png' = 'https://upload.wikimedia.org/wikipedia/commons/thumb/c/c3/Ishihara_11.PNG/500px-Ishihara_11.PNG'
    '06.png' = 'https://upload.wikimedia.org/wikipedia/commons/thumb/a/ae/Ishihara_19.PNG/500px-Ishihara_19.PNG'
    '07.png' = 'https://upload.wikimedia.org/wikipedia/commons/thumb/f/f0/Ishihara_23.PNG/500px-Ishihara_23.PNG'
    '08.png' = 'https://upload.wikimedia.org/wikipedia/commons/thumb/9/98/Ishihara_2.svg/500px-Ishihara_2.svg.png'
}

$ua = 'gin-vue-admin-portal/1.0 (Educational color-blind test; contact: local)'

function Save-Image($url, $outPath) {
    try {
        Invoke-WebRequest -Uri $url -OutFile $outPath -Headers @{ 'User-Agent' = $ua } -UseBasicParsing -TimeoutSec 120
        return (Test-Path $outPath) -and ((Get-Item $outPath).Length -gt 500)
    } catch {
        Write-Host "   $($_.Exception.Message)"
        return $false
    }
}

$ok = 0
foreach ($kv in $plateUrls.GetEnumerator() | Sort-Object Name) {
    $out = Join-Path $destDir $kv.Key
    Write-Host ">> $($kv.Key)"
    Start-Sleep -Seconds 4
    if (Save-Image $kv.Value $out) {
        Write-Host "   OK $((Get-Item $out).Length) bytes" -ForegroundColor Green
        $ok++
    } else {
        Write-Host "   FAIL" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "Done: $ok / $($plateUrls.Count) -> $destDir"
