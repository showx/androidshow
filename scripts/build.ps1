$ErrorActionPreference = "Stop"
$version = if ($env:VERSION) { $env:VERSION } else { "0.1.0" }
$ldflags = "-s -w -X main.version=$version"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

New-Item -ItemType Directory -Force -Path "bin", "dist" | Out-Null

Write-Host "构建当前平台 -> bin/ashow.exe"
$env:CGO_ENABLED = "0"
go build -ldflags $ldflags -o "bin/ashow.exe" ./cmd/ashow

$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Out = "dist/ashow-windows-amd64.exe" },
    @{ GOOS = "windows"; GOARCH = "arm64"; Out = "dist/ashow-windows-arm64.exe" },
    @{ GOOS = "linux";   GOARCH = "amd64"; Out = "dist/ashow-linux-amd64" },
    @{ GOOS = "linux";   GOARCH = "arm64"; Out = "dist/ashow-linux-arm64" },
    @{ GOOS = "darwin";  GOARCH = "amd64"; Out = "dist/ashow-darwin-amd64" },
    @{ GOOS = "darwin";  GOARCH = "arm64"; Out = "dist/ashow-darwin-arm64" }
)

foreach ($t in $targets) {
    Write-Host ("交叉编译 {0}/{1} -> {2}" -f $t.GOOS, $t.GOARCH, $t.Out)
    $env:GOOS = $t.GOOS
    $env:GOARCH = $t.GOARCH
    $env:CGO_ENABLED = "0"
    go build -ldflags $ldflags -o $t.Out ./cmd/ashow
}

Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue

Write-Host "完成。输出目录: bin/ 与 dist/"
Get-ChildItem bin, dist | Format-Table Name, Length
