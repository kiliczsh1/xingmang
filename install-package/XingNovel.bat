@echo off
chcp 65001 >nul
setlocal
cd /d "%~dp0"
start "XingNovel" "%~dp0XingNovel-node.exe" "%~dp0launcher.js"
exit /b 0
