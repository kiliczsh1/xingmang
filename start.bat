@echo off
chcp 65001 >nul
echo ========================================
echo XingNovel - Dev Start
echo ========================================
echo.

echo [1/4] Check Node.js...
node --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Node.js was not found.
    echo Download: https://nodejs.org/
    pause
    exit /b 1
)
echo [OK] Node.js is available

echo.
echo [2/4] Check root dependencies...
if not exist node_modules (
    echo Installing root dependencies...
    call npm install
    if %errorlevel% neq 0 (
        echo [ERROR] Failed to install root dependencies
        pause
        exit /b 1
    )
) else (
    echo Checking root dependencies...
    call npm ls --depth=0 >nul 2>&1
    if %errorlevel% neq 0 (
        echo Dependencies mismatch, reinstalling...
        call npm install
        if %errorlevel% neq 0 (
            echo [ERROR] Failed to reinstall root dependencies
            pause
            exit /b 1
        )
    ) else (
        echo [OK] Root dependencies verified
    )
)

echo.
echo [3/4] Check server dependencies...
cd server
if not exist node_modules (
    echo Installing server dependencies...
    call npm install
    if %errorlevel% neq 0 (
        echo [ERROR] Failed to install server dependencies
        pause
        exit /b 1
    )
) else (
    echo Checking server dependencies...
    call npm ls --depth=0 >nul 2>&1
    if %errorlevel% neq 0 (
        echo Dependencies mismatch, reinstalling...
        call npm install
        if %errorlevel% neq 0 (
            echo [ERROR] Failed to reinstall server dependencies
            pause
            exit /b 1
        )
    ) else (
        echo [OK] Server dependencies verified
    )
)
cd ..

echo.
echo [4/4] Check web dependencies...
cd novel
if not exist node_modules (
    echo Installing web dependencies...
    call npm install
    if %errorlevel% neq 0 (
        echo [ERROR] Failed to install web dependencies
        pause
        exit /b 1
    )
) else (
    echo Checking web dependencies...
    call npm ls --depth=0 >nul 2>&1
    if %errorlevel% neq 0 (
        echo Dependencies mismatch, reinstalling...
        call npm install
        if %errorlevel% neq 0 (
            echo [ERROR] Failed to reinstall web dependencies
            cd ..
            pause
            exit /b 1
        )
    ) else (
        echo [OK] Web dependencies verified
    )
)
cd ..

echo.
echo [4/4] Start single-window dev mode...
echo.
echo Frontend: http://localhost:5173
echo Backend : http://localhost:3000
echo.
echo Logs for both services will appear in this window.
echo Closing this window will stop both services.
echo.
node dev-launcher.js
if %errorlevel% neq 0 (
    echo.
    echo Launcher exited with an error.
    pause
)
