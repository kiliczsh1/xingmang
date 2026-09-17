#Requires -Version 5.1
<#
    start-dev.ps1 - 星芒桌面版开发环境一键启动

    一次拉起三个服务：
      Go 后端        :19825   (xingmang.exe --no-window)
      Vite 前端      :5173    (node node_modules/vite/bin/vite.js)
      Python 番茄桥  :19900   (scripts/fanqie_bridge.py)

    用法：
      .\start-dev.ps1                  # 默认启动并打开浏览器
      .\start-dev.ps1 -NoBrowser       # 不自动打开浏览器
      .\start-dev.ps1 -NoRebuild       # 跳过 Go 增量构建
      .\start-dev.ps1 -SkipBridge      # 不启动 Python 番茄桥
      .\start-dev.ps1 -FrontendPort 5174

    说明：
      - 就绪判定用 TcpClient 直连（约 5ms/次），不再使用 Test-NetConnection（约 4000ms/次）
      - 三个服务的 stdout/stderr 全部重定向到 logs/，控制台只显示脚本自身的进度
      - 退出时用 taskkill /T 杀整棵进程树，避免 node / python 变孤儿继续占端口
#>
param(
    [switch]$NoBrowser,
    [switch]$NoRebuild,
    [switch]$SkipBridge,
    [int]$BackendPort  = 19825,
    [int]$FrontendPort = 5173,
    [int]$BridgePort   = 19900,
    [string]$ListenHost = '0.0.0.0'
)

$ErrorActionPreference = 'Stop'

$root   = $PSScriptRoot
$logDir = Join-Path $root 'logs'
$webDir = Join-Path $root 'web'
$exe    = Join-Path $root 'xingmang.exe'

$script:procs = @()
$script:failed = $false

# ---------------------------------------------------------------- 输出小工具
function Write-Banner {
    Write-Host ''
    Write-Host '  ============================================' -ForegroundColor Cyan
    Write-Host '    XINGMANG DESKTOP  -  DEVELOPMENT MODE'      -ForegroundColor Cyan
    Write-Host '  ============================================' -ForegroundColor Cyan
}

function Write-Step([string]$msg) { Write-Host "  > $msg" -ForegroundColor White }
function Write-Note([string]$msg) { Write-Host "    $msg" -ForegroundColor DarkGray }
function Write-Good([string]$msg) { Write-Host "    [OK]   $msg" -ForegroundColor Green }
function Write-Warn([string]$msg) { Write-Host "    [WARN] $msg" -ForegroundColor Yellow }
function Write-Bad ([string]$msg) { Write-Host "    [FAIL] $msg" -ForegroundColor Red }

function Show-LogTail([string]$path, [int]$lines = 25) {
    if (-not (Test-Path -LiteralPath $path)) { return }
    $content = Get-Content -LiteralPath $path -Tail $lines -ErrorAction SilentlyContinue
    if (-not $content) { return }
    Write-Host ''
    Write-Host "    ---- tail $(Split-Path -Leaf $path) ----" -ForegroundColor DarkGray
    $esc = [char]27
    foreach ($line in $content) {
        $line = $line -replace "$esc\[[0-9;]*[A-Za-z]", ''
        Write-Host "    $line" -ForegroundColor DarkGray
    }
}

# .NET Framework 的 Process.Start 在 UseShellExecute=false（即开启输出重定向）时，
# 会把环境变量塞进一个**大小写敏感**的字典。若环境里同时存在 HTTP_PROXY / http_proxy
# 这类只差大小写的变量，会直接抛 "Item has already been added"。
# 这里在拉起子进程前按大小写去重（同名保留首个，Windows 本身对环境变量大小写不敏感）。
function Repair-EnvCaseConflicts {
    $seen    = @{}
    $removed = @()
    foreach ($key in @([Environment]::GetEnvironmentVariables('Process').Keys)) {
        $name = [string]$key
        $lk   = $name.ToLowerInvariant()
        if ($seen.ContainsKey($lk)) {
            [Environment]::SetEnvironmentVariable($name, $null, 'Process')
            $removed += $name
        } else {
            $seen[$lk] = $name
        }
    }
    return $removed
}

# ---------------------------------------------------------------- 端口 / 进程
# 只取 Listen 状态且 pid > 4 的连接，避免 Stop-Process -Id 0 / 4（System）报错
function Get-PortOwners([int]$Port) {
    Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue |
        Select-Object -ExpandProperty OwningProcess -Unique |
        Where-Object { $_ -gt 4 }
}

# 本脚本 $ErrorActionPreference='Stop'。Windows PowerShell 5.1 有个坑：
# 原生命令（taskkill / python / robocopy …）一旦**重定向了 stderr**（2>&1 或 2>$null），
# 它写的每一行 stderr 都会被包成 NativeCommandError 的 ErrorRecord，并因 Stop 直接终止脚本
# —— 表现为「Failed to start / stopped, exit code 1」，光加 2>$null 也拦不住。
# 所以凡是「允许失败」的原生命令，都必须走这个包装：函数内把 Stop 降为 Continue 再 try/catch。
function Invoke-Native([string]$FilePath, [string[]]$Arguments) {
    $prev = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        & $FilePath @Arguments 2>$null | Out-Null
    } catch {
        return $false
    } finally {
        $ErrorActionPreference = $prev
    }
    return ($LASTEXITCODE -eq 0)
}

# taskkill 返回非 0 是常态（进程已退出、句柄未释放、权限不足），绝不能让它打断启动流程。
function Stop-ProcessTree([int]$ProcessId) {
    return (Invoke-Native 'taskkill.exe' @('/PID', "$ProcessId", '/T', '/F'))
}

function Stop-PortOwners([int]$Port) {
    foreach ($ownerPid in @(Get-PortOwners $Port)) {
        $name = (Get-Process -Id $ownerPid -ErrorAction SilentlyContinue).ProcessName
        Write-Note "releasing port $Port (pid $ownerPid $name)"
        # 失败不报错：端口若仍被占用，会在下面的就绪等待里以 [FAIL] + 日志尾部暴露出来
        $null = Stop-ProcessTree $ownerPid
    }
}

# TcpClient 直连探测：命中 ~5ms，未命中按超时返回，不打印任何进度噪音
function Test-PortReady([int]$Port, [int]$TimeoutMs = 400) {
    $tcp = New-Object System.Net.Sockets.TcpClient
    try {
        $ar = $tcp.BeginConnect('127.0.0.1', $Port, $null, $null)
        if (-not $ar.AsyncWaitHandle.WaitOne($TimeoutMs)) { return $false }
        try { $tcp.EndConnect($ar) | Out-Null } catch { return $false }
        return $tcp.Connected
    } catch {
        return $false
    } finally {
        $tcp.Close()
    }
}

# 带进度条的就绪等待，就绪即返回，不再固定 sleep
function Wait-ForPort([int]$Port, [string]$Label, [int]$TimeoutSec = 40) {
    $sw   = [Diagnostics.Stopwatch]::StartNew()
    $barW = 24
    while ($true) {
        $el = $sw.Elapsed.TotalSeconds
        if (Test-PortReady -Port $Port) {
            $bar = '[' + ('#' * $barW) + ']'
            Write-Host ("`r    {0} {1,5:F1}s  {2} ready" -f $bar, $el, $Label) -NoNewline
            Write-Host ''
            return $true
        }
        if ($el -ge $TimeoutSec) { break }
        $n   = [int][Math]::Floor($barW * ($el / $TimeoutSec))
        $n   = [Math]::Max(0, [Math]::Min($barW, $n))
        $bar = '[' + ('#' * $n) + ('-' * ($barW - $n)) + ']'
        $msg = "`r    {0} {1,5:F1}s  waiting for {2}" -f $bar, $el, $Label
        Write-Host $msg.PadRight(78) -NoNewline -ForegroundColor Yellow
        Start-Sleep -Milliseconds 150
    }
    Write-Host ''
    return $false
}

# ---------------------------------------------------------------- 启动 / 收尾
function Start-Logged {
    param(
        [string]$Name,
        [string]$FilePath,
        [string[]]$Arguments,
        [string]$WorkingDirectory
    )
    $out = Join-Path $logDir "$Name.out.log"
    $err = Join-Path $logDir "$Name.err.log"
    try {
        $p = Start-Process -FilePath $FilePath -ArgumentList $Arguments `
            -WorkingDirectory $WorkingDirectory `
            -RedirectStandardOutput $out -RedirectStandardError $err `
            -WindowStyle Hidden -PassThru
    } catch {
        Write-Warn "$Name 日志重定向不可用，退化为无日志启动：$($_.Exception.Message)"
        $p = Start-Process -FilePath $FilePath -ArgumentList $Arguments `
            -WorkingDirectory $WorkingDirectory -WindowStyle Hidden -PassThru
    }
    $script:procs += [pscustomobject]@{ Name = $Name; Proc = $p; Out = $out; Err = $err }
    return $p
}

function Stop-All {
    foreach ($item in $script:procs) {
        if ($item.Proc -and -not $item.Proc.HasExited) {
            $null = Stop-ProcessTree $item.Proc.Id
        }
    }
    foreach ($port in @($BackendPort, $FrontendPort, $BridgePort)) { Stop-PortOwners $port }
}

function Open-Url([string]$Url) {
    try {
        Start-Process -FilePath $Url | Out-Null
        return
    } catch { }
    & rundll32.exe url.dll,FileProtocolHandler $Url
}

# ================================================================ 主流程
Write-Banner

if (-not (Test-Path -LiteralPath $logDir)) { New-Item -ItemType Directory -Path $logDir -Force | Out-Null }

$dupes = @(Repair-EnvCaseConflicts)
if ($dupes.Count -gt 0) { Write-Note "deduped env: $($dupes -join ', ')" }

# 1. 环境检查 -------------------------------------------------------------
Write-Step 'checking toolchain'
$node = Get-Command node -ErrorAction SilentlyContinue
if (-not $node) { Write-Bad 'node not found in PATH'; exit 1 }
if (-not (Test-Path -LiteralPath (Join-Path $webDir 'node_modules\vite\bin\vite.js'))) {
    Write-Bad 'web\node_modules 缺失或不完整，请先在 web 目录执行 npm install'
    exit 1
}
$python = Get-Command python -ErrorAction SilentlyContinue
if (-not $python -and -not $SkipBridge) { Write-Warn 'python not found，跳过番茄桥'; $SkipBridge = $true }
if (-not $SkipBridge) {
    # 番茄桥依赖 httpx（见 scripts\fanqie_bridge.py 顶部说明）。
    # 本机 PATH 里有多个 python，命中的那个不一定装了 httpx；缺依赖就直接跳过，
    # 否则桥会启动即崩，进而把后端/前端一起带下来（守护逻辑会整体退出）。
    if (-not (Invoke-Native $python.Source @('-c', 'import httpx'))) {
        Write-Warn "python $($python.Source) 缺少 httpx，跳过番茄桥（pip install httpx 后可恢复）"
        $SkipBridge = $true
    }
}
Write-Good "node $($node.Source)"

# 2. 释放旧实例占用的端口 -------------------------------------------------
Write-Step 'releasing ports'
foreach ($p in @($BackendPort, $FrontendPort, $BridgePort)) { Stop-PortOwners $p }
Start-Sleep -Milliseconds 300

# 3. Go 增量构建 ----------------------------------------------------------
# exe 不存在、或任一 .go / 前端产物比 exe 新，才重新编译
$needBuild = $false
$exeInfo   = Get-Item -LiteralPath $exe -ErrorAction SilentlyContinue
if (-not $exeInfo) {
    $needBuild = $true
    Write-Step 'xingmang.exe missing, building'
} else {
    $watch = @()
    foreach ($d in @('cmd', 'internal')) { $watch += (Join-Path $root $d) }
    $newest = Get-ChildItem -Path $watch -Recurse -Filter *.go -ErrorAction SilentlyContinue |
        Sort-Object LastWriteTime -Descending | Select-Object -First 1
    $frontIdx = Join-Path $root 'web\dist\index.html'
    if (Test-Path -LiteralPath $frontIdx) { $newest = @($newest, (Get-Item -LiteralPath $frontIdx)) | Sort-Object LastWriteTime -Descending | Select-Object -First 1 }
    if ($newest -and $newest.LastWriteTime -gt $exeInfo.LastWriteTime) {
        $needBuild = $true
        Write-Step "source newer than exe ($($newest.Name)), rebuilding"
    }
}

if ($needBuild -and -not $NoRebuild) {
    $go = Get-Command go -ErrorAction SilentlyContinue
    if (-not $go) { Write-Bad 'go not found in PATH，无法编译后端'; exit 1 }

    # 前端产物同步到 embed 目录，保证 exe 里的页面是最新的
    $srcDist = Join-Path $root 'web\dist'
    $dstDist = Join-Path $root 'cmd\xingmang\web\dist'
    if (Test-Path -LiteralPath $srcDist) {
        Write-Note 'syncing web\dist -> cmd\xingmang\web\dist'
        # robocopy 返回码 1~7 都算成功，真失败也不该中断启动（见 Invoke-Native 注释）
        $null = Invoke-Native 'robocopy.exe' @($srcDist, $dstDist, '/MIR', '/NFL', '/NDL', '/NJH', '/NJS', '/NP', '/NBT')
    }

    Push-Location $root
    & go build -ldflags="-s -w" -o xingmang.exe ./cmd/xingmang/
    $code = $LASTEXITCODE
    Pop-Location
    if ($code -ne 0) { Write-Bad "go build failed (exit $code)"; exit 1 }
    Write-Good 'backend built'
} elseif ($needBuild) {
    Write-Warn 'sources changed but -NoRebuild specified，沿用旧 exe'
} else {
    Write-Good 'backend up to date'
}

# 4. 拉起服务 -------------------------------------------------------------
Write-Step 'starting services'

$null = Start-Logged -Name 'backend' -FilePath $exe `
    -Arguments @('--no-window', "--port=$BackendPort") -WorkingDirectory $root

if (-not $SkipBridge) {
    $null = Start-Logged -Name 'fanqie' -FilePath $python.Source `
        -Arguments @('scripts\fanqie_bridge.py') -WorkingDirectory $root
}

$viteJs = Join-Path $webDir 'node_modules\vite\bin\vite.js'
$null = Start-Logged -Name 'vite' -FilePath $node.Source `
    -Arguments @("`"$viteJs`"", '--host', $ListenHost, '--port', $FrontendPort, '--strictPort') `
    -WorkingDirectory $webDir

# 5. 就绪等待（并行轮询，就绪即往下走）-------------------------------------
Write-Step 'waiting for services'
$backendOk  = Wait-ForPort -Port $BackendPort  -Label "backend  :$BackendPort"  -TimeoutSec 40
$frontendOk = Wait-ForPort -Port $FrontendPort -Label "frontend :$FrontendPort" -TimeoutSec 40
$bridgeOk   = $true
if (-not $SkipBridge) {
    $bridgeOk = Wait-ForPort -Port $BridgePort -Label "fanqie   :$BridgePort" -TimeoutSec 40
}

# 6. 结果面板 -------------------------------------------------------------
Write-Host ''
Write-Host '  ============================================' -ForegroundColor Cyan
if ($backendOk)  { Write-Good "Backend    http://127.0.0.1:$BackendPort" }
else             { Write-Bad  "Backend    http://127.0.0.1:$BackendPort" }
if ($frontendOk) { Write-Good "Frontend   http://127.0.0.1:$FrontendPort" }
else             { Write-Bad  "Frontend   http://127.0.0.1:$FrontendPort" }
if ($SkipBridge) { Write-Warn "Fanqie     skipped (-SkipBridge)" }
elseif ($bridgeOk) { Write-Good "Fanqie     http://127.0.0.1:$BridgePort" }
else             { Write-Bad  "Fanqie     http://127.0.0.1:$BridgePort" }
Write-Host '  ============================================' -ForegroundColor Cyan
Write-Host "    logs -> $logDir" -ForegroundColor DarkGray

if (-not $backendOk)  { Show-LogTail (Join-Path $logDir 'backend.err.log') }
if (-not $frontendOk) { Show-LogTail (Join-Path $logDir 'vite.err.log') }
if (-not $SkipBridge -and -not $bridgeOk) { Show-LogTail (Join-Path $logDir 'fanqie.err.log') }

if (-not $backendOk -or -not $frontendOk) {
    Write-Host ''
    Write-Bad '服务未就绪，正在退出（日志见上）'
    Stop-All
    exit 1
}

if (-not $NoBrowser) {
    if ($frontendOk) { Open-Url "http://127.0.0.1:$FrontendPort" }
    else             { Open-Url "http://127.0.0.1:$BackendPort" }
}

# 7. 守护：任一服务挂掉就整体退出，不再每秒刷屏 ----------------------------
Write-Host ''
Write-Host '  all services up. Press Ctrl+C to stop everything.' -ForegroundColor Cyan

try {
    while ($true) {
        foreach ($item in $script:procs) {
            if ($item.Proc -and $item.Proc.HasExited) {
                Write-Host ''
                Write-Bad "$($item.Name) exited unexpectedly"
                Show-LogTail $item.Err
                Show-LogTail $item.Out 25
                $script:failed = $true
                break
            }
        }
        if ($script:failed) { break }
        Start-Sleep -Milliseconds 500
    }
} finally {
    Write-Host ''
    Write-Host '  shutting down...' -ForegroundColor Yellow
    Stop-All
    Write-Host '  done.' -ForegroundColor Green
}

if ($script:failed) { exit 1 }
