@echo off
setlocal
title XINGMANG DEV
cd /d "%~dp0"

rem ---------------------------------------------------------------------
rem  Wrapper only. All start logic lives in start-dev.ps1.
rem  Passes every argument straight through, e.g.:
rem      start-dev.bat -NoBrowser
rem      start-dev.bat -FrontendPort 5174
rem
rem  NOTE: keep this file CRLF + pure ASCII.
rem  cmd.exe renders a LF-only .bat unusable, and decodes UTF-8 text as GBK.
rem ---------------------------------------------------------------------

set "PS1=%~dp0start-dev.ps1"
if not exist "%PS1%" (
    echo.
    echo [FATAL] start-dev.ps1 not found next to this script:
    echo         "%PS1%"
    pause
    endlocal
    exit /b 1
)

set "PSEXE=powershell"
where powershell >nul 2>nul
if errorlevel 1 set "PSEXE=%SystemRoot%\System32\WindowsPowerShell\v1.0\powershell.exe"

echo   launching start-dev.ps1 ...
"%PSEXE%" -NoLogo -NoProfile -ExecutionPolicy Bypass -File "%PS1%" %*
set "RC=%ERRORLEVEL%"

if not "%RC%"=="0" (
    echo.
    echo === Failed to start / stopped, exit code %RC% ===
    echo Logs: %~dp0logs
    pause
    endlocal
    exit /b %RC%
)

endlocal
