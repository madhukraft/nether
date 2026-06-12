$Repo = "madhukraft/nether"
$BinDir = "$env:USERPROFILE\.nether\bin"
$BinPath = "$BinDir\nether.exe"

switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { $Arch = "amd64" }
    "ARM64" { $Arch = "arm64" }
    default {
        Write-Error "Error: Unsupported architecture '$env:PROCESSOR_ARCHITECTURE'. Expected AMD64 or ARM64."
        exit 1
    }
}

$Url = "https://github.com/$Repo/releases/latest/download/nether-windows-$Arch.exe"

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

# Detect latest version from GitHub redirect
$NewVer = "unknown"
try {
    $request = [System.Net.HttpWebRequest]::Create($Url)
    $request.AllowAutoRedirect = $false
    $response = $request.GetResponse()
    $redirectUrl = $response.GetResponseHeader("Location")
    $response.Close()
    if ($redirectUrl) {
        $match = [regex]::Match($redirectUrl, '/download/([^/]+)/')
        if ($match.Success) {
            $NewVer = $match.Groups[1].Value
        }
    }
} catch {
    Write-Warning "Could not detect latest version: $($_.Exception.Message)"
}

# Check existing installation
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
Write-Output "  URL: $Url"

try {
    $webClient = New-Object System.Net.WebClient
    $webClient.DownloadFile($Url, $TmpFile)
    $webClient.Dispose()
} catch {
    Remove-Item -Force $TmpFile -ErrorAction SilentlyContinue
    Write-Error "Error: Download failed. $($_.Exception.Message)"
    exit 1
}

$fileInfo = Get-Item $TmpFile -ErrorAction SilentlyContinue
if (-not $fileInfo -or $fileInfo.Length -eq 0) {
    Remove-Item -Force $TmpFile -ErrorAction SilentlyContinue
    Write-Error "Error: Downloaded file is empty. The release may not exist for windows/$Arch."
    exit 1
}

New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
Move-Item -Force $TmpFile $BinPath

try {
    $test = & $BinPath -V 2>&1
    if ($LASTEXITCODE -ne 0) { throw "exit code $LASTEXITCODE" }
} catch {
    Write-Warning "Warning: Installed binary failed to run. It may be incompatible with your system."
    Write-Warning "  Error: $($_.Exception.Message)"
}

$pathUpdated = $false
$Path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($Path -notlike "*$BinDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$Path;$BinDir", "User")
    $pathUpdated = $true
}

Write-Output "Installed v$NewVer to $BinPath"
if ($pathUpdated) {
    Write-Output "Added $BinDir to your user PATH."
    Write-Output "Restart your terminal or run: `$env:Path = [Environment]::GetEnvironmentVariable('Path', 'User')"
}
Write-Output "Run 'nether create' to get started."
