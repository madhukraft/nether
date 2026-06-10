$Repo = "madhukraft/nether"
$BinDir = "$env:USERPROFILE\.nether\bin"
$BinPath = "$BinDir\nether.exe"

switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { $Arch = "amd64" }
    "ARM64" { $Arch = "arm64" }
    default { Write-Error "Unsupported architecture"; exit 1 }
}

$Url = "https://github.com/$Repo/releases/latest/download/nether-windows-$Arch.exe"

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$request = [System.Net.HttpWebRequest]::Create($Url)
$request.AllowAutoRedirect = $false
try {
    $response = $request.GetResponse()
    $redirectUrl = $response.GetResponseHeader("Location")
    $response.Close()
} catch {
    $redirectUrl = $_.Exception.Response.GetResponseHeader("Location")
}
$NewVer = [regex]::Match($redirectUrl, '/download/([^/]+)/').Groups[1].Value
if (-not $NewVer) { $NewVer = "unknown" }

if (Test-Path $BinPath) {
    $OldVer = & $BinPath -V 2>$null
    if ($LASTEXITCODE -ne 0 -or -not $OldVer) { $OldVer = $null }
    Write-Output "Existing nether binary found at $BinPath"
    if ($OldVer) {
        Write-Output "  Current version: v$OldVer"
    }
    Write-Output "  New version:     v$NewVer"
    $response = Read-Host "Overwrite? [y/N]"
    if ($response -notmatch '^[yY]') {
        Write-Output "Aborting."
        exit
    }
}

$TmpFile = [System.IO.Path]::GetTempFileName()
Write-Output "Downloading nether v$NewVer for windows/$Arch..."
Invoke-WebRequest -Uri $Url -OutFile $TmpFile -UseBasicParsing

New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
Move-Item -Force $TmpFile $BinPath

$Path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($Path -notlike "*$BinDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$Path;$BinDir", "User")
}

Write-Output "Installed v$NewVer to $BinPath"
Write-Output "Added $BinDir to your user PATH."
Write-Output "Restart your terminal or run: `$env:Path = [Environment]::GetEnvironmentVariable('Path', 'User')"
Write-Output "Then run 'nether create' to get started."
