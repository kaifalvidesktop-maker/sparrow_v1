@echo off
setlocal
cd /d "%~dp0"

if not exist "sparrow.exe" (
  echo Sparrow.exe was not found in this folder.
  pause
  exit /b 1
)

echo Registering Sparrow for HTML, HTM, PDF, HTTP and HTTPS...
sparrow.exe --register-file-associations

echo.
echo Sparrow is now registered with Windows.
echo To make it the default browser, open:
echo   Settings ^> Apps ^> Default apps ^> Sparrow Browser
echo.
echo HTML and PDF files can then be double-clicked and will open in Sparrow.
pause
