$Repo = "madhukraft/nether"
$BinDir = "$env:USERPROFILE\.nether\bin"
$BinPath = "$BinDir\nether.exe"

switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { $Arch = "amd64" }
    "ARM64" { $Arch = "arm64" }
    default { Write-Error "Unsupported architecture"; exit 1 }
}

$Url = "https://github.com/$Repo/releases/latest/download/nether-windows-$Arch.exe"
$TmpFile = [System.IO.Path]::GetTempFileName()

Write-Output "Downloading nether for windows/$Arch..."
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
Invoke-WebRequest -Uri $Url -OutFile $TmpFile -UseBasicParsing

$NewVer = & $TmpFile -V 2>$null
if ($LASTEXITCODE -ne 0 -or -not $NewVer) { $NewVer = "unknown" }

if (Test-Path $BinPath) {
    $OldVer = & $BinPath -V 2>$null
    if ($LASTEXITCODE -ne 0 -or -not $OldVer) { $OldVer = $null }
    Write-Output "Existing nether found at $BinPath"
    if ($OldVer) {
        Write-Output "  Current version: $OldVer"
    }
    Write-Output "  New version:     $NewVer"
    $response = Read-Host "Overwrite? [y/N]"
    if ($response -notmatch '^[yY]') {
        Write-Output "Aborting."
        Remove-Item -Force $TmpFile
        exit
    }
}

New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
Move-Item -Force $TmpFile $BinPath

$Path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($Path -notlike "*$BinDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$Path;$BinDir", "User")
}

Write-Output "Installed $NewVer to $BinPath"
Write-Output "Added $BinDir to your user PATH."
Write-Output "Restart your terminal or run: `$env:Path = [Environment]::GetEnvironmentVariable('Path', 'User')"
Write-Output "Then run 'nether create' to get started."
