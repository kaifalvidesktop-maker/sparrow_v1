@echo off
setlocal
cd /d "%~dp0"
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

for /f %%i in ('powershell -NoProfile -Command "Get-Date -Format yyyyMMddHHmmssfff"') do set "SPARROW_BUILD_ID=%%i"
if not defined SPARROW_BUILD_ID (
  echo Failed to generate build ID.
  exit /b 1
)

go build -ldflags="-H windowsgui -X main.buildID=%SPARROW_BUILD_ID%" -o sparrow.exe .
if errorlevel 1 exit /b %errorlevel%

echo Sparrow.exe built successfully.
