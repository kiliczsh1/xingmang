@echo off
setlocal
title XINGMANG BUILD
cd /d "%~dp0"

rem =====================================================================
rem  build.bat - release packaging entry (double-click friendly).
rem
rem  All packaging logic lives in the Go tool tools/release
rem  (go run ./tools/release). This script only supplies the default
rem  flags plus the final pause, so its usage is identical to build.ps1.
rem
rem  The pre-build checks (running exe / missing web\node_modules /
rem  runtime data sitting in dist\) also live in tools/release, not here.
rem  That way build.bat and build.ps1 cannot drift apart - and every
rem  check runs before the ~1 minute frontend build, not after it.
rem
rem  NOTE: keep this file CRLF + pure ASCII.
rem  cmd.exe renders a LF-only .bat unusable, and on a non-UTF-8 system it
rem  decodes UTF-8 text as GBK, so non-ASCII messages would print as
rem  garbage. (See start-dev.bat for the same convention.)
rem
rem  Usage:
rem    build.bat                                  default: windows/amd64 + frontend + dist\windows-amd64.zip
rem    build.bat -Version 0.772                   versioned archive dist\V-0.772.zip
rem    build.bat -SkipFrontend                    skip frontend build (rebuilt anyway if web\dist is stale)
rem    build.bat -NoZip                           emit only dist\<os>-<arch>\, no archive
rem    build.bat -IncludeMigrate                  also compile migrate.exe
rem    build.bat -OS linux -Arch arm64            cross compile linux/arm64
rem    build.bat -OS windows -Arch amd64,arm64    build several targets in one run
rem
rem  Output:
rem    dist\<os>-<arch>\        xingmang.exe + README.txt
rem    dist\<os>-<arch>.zip     "latest" archive, when no -Version is given
rem    dist\V-<ver>.zip         versioned archive, when -Version is given
rem                             (multi-target: V-<ver>-<os>-<arch>.zip)
rem
rem  WARNING - always pass -Version when the build is going to be handed
rem  to someone. Without it the version comes from the git short commit,
rem  so the UI shows a hash like "d7896fe-dirty" that cannot be matched
rem  against the changelog. This script deliberately has no built-in
rem  version number: bumping it is the release step, not the script's job.
rem
rem  WARNING - dist\<os>-<arch>\ is wiped and rebuilt on every run, and
rem  the program stores its data next to the exe. So if anyone ever
rem  double-clicked the exe inside dist\, their works are in there and
rem  will be deleted. tools/release refuses to run in that case and lists
rem  what it found; copy it out first. (The archive itself always strips
rem  books\ data\ workflow_outputs\ logs\, so nothing personal ships.)
rem =====================================================================

rem Pause when launched by double-click (but not when called from an
rem existing terminal), so error output does not flash by. Matching the
rem full path %~f0 keeps "cmd /c build.bat" out of that branch.
rem Two notes:
rem   - call find.exe by absolute path: a Git/MSYS usr\bin on PATH shadows
rem     Windows' find.exe with the Unix one, which rejects /i and silently
rem     kills this check.
rem   - %cmdcmdline% holds the command line as written, so the path inside
rem     it may use forward slashes while %~f0 never does. Normalize first.
set "PAUSE_END="
set "CMDLINE=%cmdcmdline%"
set "CMDLINE=%CMDLINE:/=\%"
echo %CMDLINE% | "%SystemRoot%\System32\find.exe" /i "%~f0" >nul && set "PAUSE_END=1"

rem Fail early with something readable, instead of the shell's cryptic
rem "'go' is not recognized as an internal or external command".
where go >nul 2>nul
if errorlevel 1 (
    echo.
    echo [FATAL] go not found on PATH - install Go first: https://go.dev/dl/
    if defined PAUSE_END pause
    endlocal
    exit /b 1
)

if not exist "%~dp0tools\release\main.go" (
    echo.
    echo [FATAL] tools\release\main.go not found next to this script:
    echo         "%~dp0tools\release\main.go"
    if defined PAUSE_END pause
    endlocal
    exit /b 1
)

rem Default is an archive; pass -NoZip explicitly to get the bare directory.
go run ./tools/release -Zip %*
set "RC=%ERRORLEVEL%"

if not "%RC%"=="0" (
    echo.
    echo === BUILD FAILED, exit code %RC% ===
    if defined PAUSE_END pause
    endlocal
    exit /b %RC%
)

if defined PAUSE_END pause
endlocal
