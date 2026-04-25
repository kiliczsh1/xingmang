@echo off
chcp 65001 >nul
setlocal

set "ROOT=%~dp0"
set "APP_NAME=XingNovel"
set "NODE_RUNTIME_EXE=%APP_NAME%-node.exe"
set "APP_LAUNCHER=%APP_NAME%.bat"
set "STOP_LAUNCHER=Stop-%APP_NAME%.bat"
set "OUTPUT_DIR_NAME=install-package"

pushd "%ROOT%" >nul

echo ========================================
echo   %APP_NAME% - Portable Build
echo ========================================
echo.

where node >nul 2>nul || goto :no_node
where npm >nul 2>nul || goto :no_npm

echo [1/6] Load project build config...
call :load_project_config
if errorlevel 1 goto :fail

set "OUTPUT_DIR=%ROOT%%OUTPUT_DIR_NAME%"
set "LAUNCHER_JS=%OUTPUT_DIR%\launcher.js"
set "OUTPUT_NODE_EXE=%OUTPUT_DIR%\%NODE_RUNTIME_EXE%"
set "OUTPUT_SERVER_DIR=%OUTPUT_DIR%\server"
set "OUTPUT_SERVER_DB_DIR=%OUTPUT_SERVER_DIR%\database"
set "OUTPUT_SERVER_NODE_MODULES=%OUTPUT_SERVER_DIR%\node_modules"
set "OUTPUT_SERVER_UPLOADS=%OUTPUT_SERVER_DIR%\uploads"
set "OUTPUT_FRONTEND_DIST=%OUTPUT_DIR%\novel\dist"

echo   output: %OUTPUT_DIR_NAME%
echo   launcher: %APP_LAUNCHER%
echo.

echo [2/6] Check dependencies...
set "NEED_NOVEL_INSTALL=0"
set "NEED_SERVER_INSTALL=0"

if not exist "%ROOT%novel\node_modules" set "NEED_NOVEL_INSTALL=1"
if not exist "%ROOT%server\node_modules" set "NEED_SERVER_INSTALL=1"

if "%NEED_NOVEL_INSTALL%"=="0" if not exist "%ROOT%novel\node_modules\.bin\vite.cmd" set "NEED_NOVEL_INSTALL=1"
if "%NEED_SERVER_INSTALL%"=="0" if not exist "%ROOT%server\node_modules\better-sqlite3\build\Release\better_sqlite3.node" set "NEED_SERVER_INSTALL=1"

if "%NEED_NOVEL_INSTALL%"=="1" (
    echo Installing frontend dependencies...
    pushd "%ROOT%novel" >nul
    call npm.cmd install
    if errorlevel 1 (
        popd >nul
        goto :fail
    )
    popd >nul
) else (
    echo Frontend dependencies OK
)

if "%NEED_SERVER_INSTALL%"=="1" (
    echo Installing backend dependencies...
    pushd "%ROOT%server" >nul
    call npm.cmd install
    if errorlevel 1 (
        popd >nul
        goto :fail
    )
    popd >nul
) else (
    echo Backend dependencies OK
)

echo.
echo [3/6] Build frontend (novel/dist)...
pushd "%ROOT%novel" >nul
call npm.cmd run build
if errorlevel 1 (
    popd >nul
    goto :fail
)
popd >nul

echo.
echo [4/6] Prepare portable package...
call :stop_packaged_app
if errorlevel 1 goto :fail
call :remove_dir "%OUTPUT_DIR%"
if errorlevel 1 goto :fail

mkdir "%OUTPUT_DIR%"
if errorlevel 1 goto :fail
mkdir "%OUTPUT_SERVER_DIR%"
if errorlevel 1 goto :fail
mkdir "%OUTPUT_SERVER_DB_DIR%"
if errorlevel 1 goto :fail
mkdir "%OUTPUT_SERVER_UPLOADS%"
if errorlevel 1 goto :fail
mkdir "%OUTPUT_SERVER_UPLOADS%\experience-shares"
if errorlevel 1 goto :fail

call :resolve_node_exe
if errorlevel 1 goto :fail
copy /Y "%SOURCE_NODE_EXE%" "%OUTPUT_NODE_EXE%" >nul
if errorlevel 1 goto :fail

call :copy_file "%ROOT%server\index.js" "%OUTPUT_SERVER_DIR%\index.js"
if errorlevel 1 goto :fail
call :copy_file "%ROOT%server\package.json" "%OUTPUT_SERVER_DIR%\package.json"
if errorlevel 1 goto :fail
if exist "%ROOT%server\package-lock.json" (
    call :copy_file "%ROOT%server\package-lock.json" "%OUTPUT_SERVER_DIR%\package-lock.json"
    if errorlevel 1 goto :fail
)
call :copy_file "%ROOT%server\database\db.js" "%OUTPUT_SERVER_DB_DIR%\db.js"
if errorlevel 1 goto :fail
call :copy_dir "%ROOT%server\routes" "%OUTPUT_SERVER_DIR%\routes"
if errorlevel 1 goto :fail
if exist "%ROOT%server\utils" (
    call :copy_dir "%ROOT%server\utils" "%OUTPUT_SERVER_DIR%\utils"
    if errorlevel 1 goto :fail
)
call :copy_dir "%ROOT%server\node_modules" "%OUTPUT_SERVER_NODE_MODULES%"
if errorlevel 1 goto :fail
call :copy_dir "%ROOT%novel\dist" "%OUTPUT_FRONTEND_DIST%"
if errorlevel 1 goto :fail

call :write_launcher_js
if errorlevel 1 goto :fail
call :write_launcher_bat
if errorlevel 1 goto :fail
call :write_stop_bat
if errorlevel 1 goto :fail

echo.
echo [5/6] Verify packaged files...
if not exist "%OUTPUT_NODE_EXE%" (
    echo [ERROR] Missing Node runtime: %NODE_RUNTIME_EXE%
    goto :fail
)
if not exist "%OUTPUT_SERVER_DIR%\index.js" (
    echo [ERROR] Missing packaged server entry: server\index.js
    goto :fail
)
if not exist "%OUTPUT_SERVER_NODE_MODULES%\better-sqlite3\build\Release\better_sqlite3.node" (
    echo [ERROR] Missing packaged dependency: server\node_modules\better-sqlite3
    goto :fail
)
if not exist "%OUTPUT_SERVER_NODE_MODULES%\pdfjs-dist\legacy\build\pdf.mjs" (
    echo [ERROR] Missing packaged dependency: server\node_modules\pdfjs-dist
    goto :fail
)
if not exist "%OUTPUT_FRONTEND_DIST%\index.html" (
    echo [ERROR] Missing packaged frontend build: novel\dist\index.html
    goto :fail
)
if not exist "%LAUNCHER_JS%" (
    echo [ERROR] Missing launcher.js
    goto :fail
)
if not exist "%OUTPUT_DIR%\%APP_LAUNCHER%" (
    echo [ERROR] Missing launcher: %APP_LAUNCHER%
    goto :fail
)

echo.
echo [6/6] Smoke test package...
call :smoke_test_package
if errorlevel 1 goto :fail

echo.
echo ========================================
echo Build OK
echo Output: %OUTPUT_DIR_NAME%\%APP_LAUNCHER%
echo ========================================
echo.
echo Tips:
echo   - Double click %APP_LAUNCHER% to start the packaged app
echo   - Double click %STOP_LAUNCHER% to stop all packaged app processes
echo   - Database file: server\database\novel.db
echo   - Uploads folder: server\uploads
echo.
popd >nul
if /i "%NO_PAUSE%"=="1" goto :success_exit
pause
exit /b 0

:success_exit
exit /b 0

:load_project_config
if not exist "%ROOT%server\package.json" (
    echo [ERROR] Missing server\package.json
    exit /b 1
)
if not exist "%ROOT%server\index.js" (
    echo [ERROR] Missing server\index.js
    exit /b 1
)
if not exist "%ROOT%server\database\db.js" (
    echo [ERROR] Missing server\database\db.js
    exit /b 1
)
if not exist "%ROOT%novel\package.json" (
    echo [ERROR] Missing novel\package.json
    exit /b 1
)
if not exist "%ROOT%novel\vite.config.ts" (
    echo [ERROR] Missing novel\vite.config.ts
    exit /b 1
)

if exist "%ROOT%package-for-pkg.json" (
    for /f "usebackq delims=" %%I in (`node -p "const cfg=require('./package-for-pkg.json'); (cfg.pkg&&cfg.pkg.outputPath)||'install-package'" 2^>nul`) do set "OUTPUT_DIR_NAME=%%I"
)
exit /b 0

:resolve_node_exe
set "SOURCE_NODE_EXE="
for /f "delims=" %%I in ('where node') do (
    if not defined SOURCE_NODE_EXE set "SOURCE_NODE_EXE=%%~fI"
)
if not defined SOURCE_NODE_EXE (
    echo [ERROR] Cannot locate node.exe
    exit /b 1
)
exit /b 0

:copy_file
copy /Y "%~1" "%~2" >nul
if errorlevel 1 (
    echo [ERROR] Failed to copy file: %~1
    exit /b 1
)
exit /b 0

:copy_dir
if not exist "%~1" exit /b 0
xcopy "%~1" "%~2\" /E /I /Y /Q >nul
if errorlevel 1 (
    echo [ERROR] Failed to copy directory: %~1
    exit /b 1
)
exit /b 0

:write_launcher_js
(
    echo const net = require('net'^);
    echo const path = require('path'^);
    echo const fs = require('fs'^);
    echo const { exec } = require('child_process'^);
    echo/
    echo const rootDir = __dirname;
    echo const serverEntry = path.join(rootDir, 'server', 'index.js'^);
    echo const logFile = path.join(rootDir, 'launcher.log'^);
    echo/
    echo const log = (msg^) =^> {
    echo   const timestamp = new Date(^).toISOString(^);
    echo   const line = `[${timestamp}] ${msg}\n`;
    echo   try { fs.appendFileSync(logFile, line^); } catch (e^) {}
    echo   console.log(line.trim(^)^);
    echo };
    echo/
    echo const canUsePort = (port^) =^> new Promise((resolve^) =^> {
    echo   const tester = net.createServer(^);
    echo   tester.once('error', (err^) =^> {
    echo     if (err.code === 'EADDRINUSE'^) {
    echo       resolve(false^);
    echo     } else {
    echo       resolve(false^);
    echo     }
    echo   }^);
    echo   tester.once('listening', (^) =^> {
    echo     tester.close((^) =^> resolve(true^)^);
    echo   }^);
    echo   tester.listen(port, '127.0.0.1'^);
    echo }^);
    echo/
    echo const findPort = async (start, end^) =^> {
    echo   for (let port = start; port ^<= end; port += 1^) {
    echo     const available = await canUsePort(port^);
    echo     if (available^) {
    echo       return port;
    echo     }
    echo     log(`Port ${port} is in use, trying next...`^);
    echo   }
    echo   throw new Error(`No available port found between ${start} and ${end}`^);
    echo };
    echo/
    echo const openBrowser = (port^) =^> {
    echo   if (process.env.NO_OPEN_BROWSER === '1'^) {
    echo     return;
    echo   }
    echo   setTimeout((^) =^> {
    echo     exec(`start "" "http://localhost:${port}"`^);
    echo   }, 2000^);
    echo };
    echo/
    echo const main = async (^) =^> {
    echo   log('Starting XingNovel launcher...'^);
    echo   log(`Root directory: ${rootDir}`^);
    echo   log(`Server entry: ${serverEntry}`^);
    echo/
    echo   if (!fs.existsSync(serverEntry^)^) {
    echo     throw new Error(`Server entry not found: ${serverEntry}`^);
    echo   }
    echo/
    echo   process.chdir(rootDir^);
    echo   log(`Working directory changed to: ${process.cwd(^)}`^);
    echo/
    echo   const port = await findPort(3014, 3024^);
    echo   log(`Found available port: ${port}`^);
    echo/
    echo   process.env.PORT = String(port^);
    echo   log(`Loading server entry...`^);
    echo   require(serverEntry^);
    echo   log('Server entry loaded successfully'^);
    echo   openBrowser(port^);
    echo };
    echo/
    echo main(^).catch((error^) =^> {
    echo   log(`[ERROR] Failed to start app: ${error.message}`^);
    echo   log(`Stack: ${error.stack}`^);
    echo   console.error('[launcher] Failed to start app:', error^);
    echo   process.exit(1^);
    echo }^);
) > "%LAUNCHER_JS%"
if errorlevel 1 (
    echo [ERROR] Failed to write launcher.js
    exit /b 1
)
exit /b 0

:write_launcher_bat
(
    echo @echo off
    echo chcp 65001 ^>nul
    echo setlocal
    echo cd /d "%%~dp0"
    echo/
    echo if not exist "%%~dp0%NODE_RUNTIME_EXE%" (
    echo     echo [ERROR] Node runtime not found: %NODE_RUNTIME_EXE%
    echo     pause
    echo     exit /b 1
    echo ^)
    echo if not exist "%%~dp0launcher.js" (
    echo     echo [ERROR] Launcher script not found: launcher.js
    echo     pause
    echo     exit /b 1
    echo ^)
    echo if not exist "%%~dp0server\index.js" (
    echo     echo [ERROR] Server entry not found: server\index.js
    echo     pause
    echo     exit /b 1
    echo ^)
    echo/
    echo taskkill /F /IM %NODE_RUNTIME_EXE% ^>nul 2^>nul
    echo timeout /t 1 /nobreak ^>nul
    echo/
    echo start "%APP_NAME%" "%%~dp0%NODE_RUNTIME_EXE%" "%%~dp0launcher.js"
    echo/
    echo timeout /t 2 /nobreak ^>nul
    echo tasklist /FI "IMAGENAME eq %NODE_RUNTIME_EXE%" 2^>nul ^| find /I "%NODE_RUNTIME_EXE%" ^>nul
    echo if errorlevel 1 (
    echo     echo [ERROR] Failed to start %APP_NAME%
    echo     if exist "%%~dp0launcher.log" (
    echo         echo/
    echo         echo === Launcher Log ===
    echo         type "%%~dp0launcher.log"
    echo     ^)
    echo     pause
    echo     exit /b 1
    echo ^)
    echo exit /b 0
) > "%OUTPUT_DIR%\%APP_LAUNCHER%"
if errorlevel 1 (
    echo [ERROR] Failed to write %APP_LAUNCHER%
    exit /b 1
)
exit /b 0

@echo掉
(
    echo @echo off
    echo chcp 65001 ^>nul
    echo taskkill /F /IM %NODE_RUNTIME_EXE% ^>nul 2^>nul
    echo exit /b 0
) > "%OUTPUT_DIR%\%STOP_LAUNCHER%"
if errorlevel 1 (
    echo [ERROR] Failed to write %STOP_LAUNCHER%
    exit /b 1
)
exit /b 0

:smoke_test_package
set "SMOKE_TEST_RESULT="
powershell -NoProfile -Command "$nodeExe=Join-Path $env:OUTPUT_DIR '%NODE_RUNTIME_EXE%'; $launcher=Join-Path $env:OUTPUT_DIR 'launcher.js'; $proc=Start-Process -FilePath 'cmd.exe' -ArgumentList '/c', ('set NO_OPEN_BROWSER=1&&\"\"{0}\"\" \"\"{1}\"\"' -f $nodeExe, $launcher) -PassThru; Start-Sleep -Seconds 4; $ok=$false; foreach ($port in 3014..3024) { try { $resp=Invoke-RestMethod -Uri ('http://127.0.0.1:{0}/api/health' -f $port) -TimeoutSec 3; if ($resp.status -eq 'ok') { $ok=$true; break } } catch {} }; Get-Process '%APP_NAME%-node' -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue; Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue; if (-not $ok) { exit 1 }"
if errorlevel 1 (
    echo [ERROR] Smoke test failed. Portable package did not start correctly.
    exit /b 1
)
exit /b 0

:stop_packaged_app
taskkill /F /IM %NODE_RUNTIME_EXE% >nul 2>nul
timeout /t 1 /nobreak >nul
exit /b 0

:remove_dir
set "TARGET_DIR=%~1"
if not exist "%TARGET_DIR%" exit /b 0
for /L %%I in (1,1,3) do (
    rmdir /S /Q "%TARGET_DIR%" >nul 2>nul
    if not exist "%TARGET_DIR%" exit /b 0
    echo [INFO] Waiting for locked files to release: %TARGET_DIR% (attempt %%I/3)
    call :stop_packaged_app
)
echo [ERROR] Failed to clean directory: %TARGET_DIR%
exit /b 1

:no_node
echo [ERROR] Node.js not found. Install from https://nodejs.org/
goto :fail

:no_npm
echo [ERROR] npm not found. Check your Node.js installation.
goto :fail

:fail
set "BUILD_EXIT_CODE=%errorlevel%"
if "%BUILD_EXIT_CODE%"=="0" set "BUILD_EXIT_CODE=1"
的回声。
回显：========================================
echo [ERROR] Build failed (errorlevel=%BUILD_EXIT_CODE%)
回显：========================================
的回声。
popd >nul
if /i "%NO_PAUSE%"=="1" goto :fail_exit
pause

:fail_exit
exit /b %BUILD_EXIT_CODE%
