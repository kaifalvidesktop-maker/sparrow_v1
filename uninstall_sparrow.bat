@echo off
setlocal
cd /d "%~dp0"

if exist "sparrow.exe" (
  sparrow.exe --unregister-file-associations
) else (
  echo Sparrow.exe was not found.
)
pause
