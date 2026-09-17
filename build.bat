@echo off
setlocal
rem =====================================================================
rem build.bat - 星芒桌面版发布打包入口（双击可用）
rem
rem 打包逻辑在 Go 工具 tools/release（go run ./tools/release），
rem 本脚本只负责默认参数与结束时暂停，用法与 build.ps1 一致。
rem
rem 用法：
rem   build.bat                                  # 默认：windows/amd64 + 前端 + 出包 dist\windows-amd64.zip
rem   build.bat -Version 0.736                   # 指定版本号，归档名 dist\V-0.736.zip
rem   build.bat -SkipFrontend                    # 跳过前端构建；web\dist 过期会自动重建
rem   build.bat -NoZip                           # 只出 dist\<os>-<arch>\ 目录，不打归档
rem   build.bat -IncludeMigrate                  # 额外编译 migrate.exe
rem   build.bat -OS linux -Arch arm64            # 交叉编译 linux/arm64
rem   build.bat -OS windows -Arch amd64,arm64    # 同时编译多架构
rem
rem 产物：
rem   dist\<os>-<arch>\        xingmang.exe + README.txt
rem   dist\<os>-<arch>.zip     未指定版本号时的覆盖式「最新包」
rem   dist\V-<版本>.zip        指定 -Version 时的版本化归档（多架构时为 V-<版本>-<os>-<arch>.zip）
rem =====================================================================

cd /d "%~dp0"

rem 双击运行时（不是在已有终端里调用）结束后暂停，避免报错信息一闪而过。
rem 用完整路径 %~f0 判断，避免把 cmd /c build.bat 这类终端调用也当成双击。
set "PAUSE_END="
echo %cmdcmdline% | find /i "%~f0" >nul && set "PAUSE_END=1"

rem 默认产出 zip 归档；只要目录时显式传 -NoZip
go run ./tools/release -Zip %*
if errorlevel 1 (
    echo.
    echo === 打包失败 ===
    if defined PAUSE_END pause
    exit /b 1
)

if defined PAUSE_END pause
