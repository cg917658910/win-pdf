@echo off
setlocal enabledelayedexpansion

REM 构建 64 位
echo Building 64-bit...
wails build -platform windows/amd64
if errorlevel 1 (
  echo 64-bit build failed.
  goto end
)

REM 构建 32 位
echo Building 32-bit...
wails build -platform windows/386 -o win-pdf-32bit.exe
if errorlevel 1 (
  echo 32-bit build failed.
  goto end
)

echo.
echo Build finished.
echo 64-bit: build\bin\win-pdf.exe
echo 32-bit: build\bin\win-pdf-32bit.exe

:end
pause
endlocal
