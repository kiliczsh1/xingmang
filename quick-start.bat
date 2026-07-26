@echo off
chcp 65001 >nul
echo ========================================
echo XingNovel - Quick Dev Start
echo ========================================
echo.
echo Frontend: http://localhost:5173
echo Backend : http://localhost:3000
echo.
echo Closing this window will stop both services.
echo.
node dev-launcher.js
if %errorlevel% neq 0 (
    echo.
    echo Launcher exited with an error.
    pause
)
