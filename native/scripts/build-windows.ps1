param(
    [string]$Configuration = "Release"
)

$ErrorActionPreference = "Stop"
$rootDir = (Resolve-Path (Join-Path $PSScriptRoot "../..")).Path
$outputDir = Join-Path $rootDir "build/native/windows"
$solutionPath = Join-Path $rootDir "native/windows/SkillFlow/SkillFlow.sln"

Write-Host "native windows build started: target=$outputDir"
Remove-Item -Recurse -Force $outputDir -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

dotnet publish $solutionPath `
  -c $Configuration `
  -r win-x64 `
  --self-contained true `
  -p:Platform=x64 `
  --nologo `
  -o $outputDir
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -trimpath -ldflags "-s -w" -o (Join-Path $outputDir "skillflowd.exe") (Join-Path $rootDir "cmd/skillflow")
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "native windows build completed: target=$outputDir"
