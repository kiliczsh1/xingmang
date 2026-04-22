@echo off
chcp 65001 >nul
echo ========================================
echo XingNovel - Install Dependencies
echo ========================================
echo.
echo This script will install all dependencies
echo using package-lock.json for exact versions.
echo.

echo [1/3] Installing root dependencies...
if exist package-lock.json (
    call npm ci
) else (
    call npm install
)
if %errorlevel% neq 0 (
    echo [ERROR] Failed to install root dependencies
    pause
    exit /b 1
)
echo [OK] Root dependencies installed

echo.
echo [2/3] Installing server dependencies...
cd server
if exist package-lock.json (
    call npm ci
) else (
    call npm install
)
if %errorlevel% neq 0 (
    echo [ERROR] Failed to install server dependencies
    cd ..
    pause
    exit /b 1
)
cd ..
echo [OK] Server dependencies installed

echo.
echo [3/3] Installing web dependencies...
cd novel
if exist package-lock.json (
    call npm ci
) else (
    call npm install
)
if %errorlevel% neq 0 (
    echo [ERROR] Failed to install web dependencies
    cd ..
    pause
    exit /b 1
)
cd ..
echo [OK] Web dependencies installed

echo.
echo ========================================
echo Installation complete!
echo Run start.bat to launch the application.
echo ========================================
pause
