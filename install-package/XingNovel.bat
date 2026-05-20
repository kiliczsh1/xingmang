@echo off
chcp 65001 >nul
setlocal
cd /d "%~dp0"

if not exist "%~dp0XingNovel-node.exe" (
    echo [ERROR] Node runtime not found: XingNovel-node.exe
    pause
    exit /b 1
)
if not exist "%~dp0launcher.js" (
    echo [ERROR] Launcher script not found: launcher.js
    pause
    exit /b 1
)
if not exist "%~dp0server\index.js" (
    echo [ERROR] Server entry not found: server\index.js
    pause
    exit /b 1
)

taskkill /F /IM XingNovel-node.exe >nul 2>nul
timeout /t 1 /nobreak >nul

start "XingNovel" "%~dp0XingNovel-node.exe" "%~dp0launcher.js"

timeout /t 2 /nobreak >nul
tasklist /FI "IMAGENAME eq XingNovel-node.exe" 2>nul | find /I "XingNovel-node.exe" >nul
if errorlevel 1 (
    echo [ERROR] Failed to start XingNovel
    if exist "%~dp0launcher.log" (
        echo/
        echo === Launcher Log ===
        type "%~dp0launcher.log"
    )
    pause
    exit /b 1
)
exit /b 0
