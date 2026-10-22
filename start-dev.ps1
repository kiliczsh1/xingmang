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
      - **关掉本窗口 = 关掉整个开发环境**：脚本把自己放进 Job Object，
        窗口被 X / Alt+F4 / 任务管理器结束时内核会连带结束全部子进程
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

# ---------------------------------------------------------------- 关窗即全灭
# 用户口径：**关掉这个黑窗口就等于关掉整个程序**，不该再让他去任务管理器收尸。
#
# 脚本自己的 finally 只在「脚本正常走完」时执行。点窗口的 X / Alt+F4 / 任务管理器
# 结束 powershell.exe，进程是被系统直接终止的，finally 一行都跑不到 —— backend、
# vite、番茄桥三个子进程就全成了孤儿，继续占着 19825 / 5173 / 19900。
#
# 这里用 Windows Job Object 补一条内核级保障：建一个带
# JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE 的作业对象，把**本脚本进程**放进去。
# 此后它派生的所有进程（含 node 派生的 esbuild、python 派生的子进程）自动继承该作业；
# 脚本进程无论以何种方式结束（正常退出、关窗、被强杀、崩溃），内核关闭 job 句柄时
# 都会强制结束作业内全部进程。
#
# ⚠ 必须在拉起任何子进程之前完成 assign：job 只对「加入之后」派生的进程自动生效。
#
# 与 finally 不冲突：正常退出路径上 Stop-All 已经先清干净，job 关闭时是空转。
# 失败不致命 —— 老系统或禁止嵌套 job 的环境退化为「只靠 finally 清理」，并打一行警告。
$script:jobOk = $false
try {
    Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;

public static class JobKeeper
{
    [StructLayout(LayoutKind.Sequential)]
    private struct JOBOBJECT_BASIC_LIMIT_INFORMATION
    {
        public long PerProcessUserTimeLimit;
        public long PerJobUserTimeLimit;
        public uint LimitFlags;
        public UIntPtr MinimumWorkingSetSize;
        public UIntPtr MaximumWorkingSetSize;
        public uint ActiveProcessLimit;
        public UIntPtr Affinity;
        public uint PriorityClass;
        public uint SchedulingClass;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct IO_COUNTERS
    {
        public ulong ReadOperationCount;
        public ulong WriteOperationCount;
        public ulong OtherOperationCount;
        public ulong ReadTransferCount;
        public ulong WriteTransferCount;
        public ulong OtherTransferCount;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct JOBOBJECT_EXTENDED_LIMIT_INFORMATION
    {
        public JOBOBJECT_BASIC_LIMIT_INFORMATION BasicLimitInformation;
        public IO_COUNTERS IoInfo;
        public UIntPtr ProcessMemoryLimit;
        public UIntPtr JobMemoryLimit;
        public UIntPtr PeakProcessMemoryUsed;
        public UIntPtr PeakJobMemoryUsed;
    }

    [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern IntPtr CreateJobObject(IntPtr lpJobAttributes, string lpName);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool SetInformationJobObject(IntPtr hJob, int infoClass,
        ref JOBOBJECT_EXTENDED_LIMIT_INFORMATION info, uint cbInfoLength);

    [DllImport("kernel32.dll", SetLastError = true)]
    private static extern bool AssignProcessToJobObject(IntPtr hJob, IntPtr hProcess);

    [DllImport("kernel32.dll")]
    private static extern IntPtr GetCurrentProcess();

    // 故意常驻不释放：进程结束时内核关闭该句柄 → 触发 KILL_ON_JOB_CLOSE。
    private static IntPtr _job = IntPtr.Zero;

    public static string Enforce()
    {
        if (_job != IntPtr.Zero) return "ok";

        IntPtr job = CreateJobObject(IntPtr.Zero, null);
        if (job == IntPtr.Zero)
            return "CreateJobObject failed, win32=" + Marshal.GetLastWin32Error();

        var info = new JOBOBJECT_EXTENDED_LIMIT_INFORMATION();
        info.BasicLimitInformation.LimitFlags = 0x2000; // JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
        const int ExtendedLimitInformation = 9;
        if (!SetInformationJobObject(job, ExtendedLimitInformation, ref info,
                (uint)Marshal.SizeOf(typeof(JOBOBJECT_EXTENDED_LIMIT_INFORMATION))))
            return "SetInformationJobObject failed, win32=" + Marshal.GetLastWin32Error();

        if (!AssignProcessToJobObject(job, GetCurrentProcess()))
            return "AssignProcessToJobObject failed, win32=" + Marshal.GetLastWin32Error();

        _job = job;
        return "ok";
    }
}
'@
    $jr = [JobKeeper]::Enforce()
    if ($jr -eq 'ok') {
        $script:jobOk = $true
    } else {
        Write-Host "  [WARN] 关窗自动收尾未启用（$jr）——关掉本窗口后可能残留子进程" -ForegroundColor Yellow
    }
} catch {
    Write-Host "  [WARN] 关窗自动收尾未启用（$($_.Exception.Message)）" -ForegroundColor Yellow
}

$root   = $PSScriptRoot
$logDir = Join-Path $root 'logs'
$webDir = Join-Path $root 'web'
$exe    = Join-Path $root 'xingmang.exe'

$script:procs = @()
$script:failed = $false
# 后端主动退出（界面「退出程序」/ Ctrl+C）时置真：此时其余服务是被它同步带走的，
# 属于正常收尾而非故障。详见守护循环里的判定。
$script:backendStopped = $false
# 本次启动被握手拦下（19825 上已有另一个后端）。此时**不能** Stop-All：
# 那个后端不是这次起的，按端口清理会把别人还在用的开发环境一起端掉。
$script:launchRejected = $false

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
    # 首选 Get-NetTCPConnection：结构化、快。
    #
    # ⚠ 它**偶发返回空**：NetTCPIP 模块未加载 / CIM 冷启动失败时，-ErrorAction SilentlyContinue
    # 会把错误整个吞掉，表现就是「netstat 明明显示端口被占，这里却查不到占用者」。
    # 后果是一连串静默错判：Stop-PortOwners 什么都不杀 → 第 2.5 步探测也可能没拦住 →
    # 新后端 bind 失败立即退出 → 守护循环把它当成「用户主动退出程序」→ 静默 exit 0。
    # 用户看到的是「启动成功」，实际上 vite/番茄桥刚被清掉、后端根本没起来。真实踩过一次。
    # 所以查不到时用 netstat 兜底。
    $pids = @(Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue |
        Select-Object -ExpandProperty OwningProcess -Unique |
        Where-Object { $_ -gt 4 })
    if ($pids.Count -gt 0) { return $pids }

    # 兜底：解析 netstat。只匹配 ASCII 片段，不受控制台代码页影响。
    $prev = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        $lines = & netstat.exe -ano -p TCP 2>$null
    } catch {
        $lines = @()
    } finally {
        $ErrorActionPreference = $prev
    }
    $found = @()
    foreach ($line in $lines) {
        if ("$line" -match "^\s*TCP\s+\S+:$Port\s+\S+\s+LISTENING\s+(\d+)\s*$") {
            $found += [int]$Matches[1]
        }
    }
    return @($found | Where-Object { $_ -gt 4 } | Sort-Object -Unique)
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

# 返回原生命令的退出码（而非「是否等于 0」）。
# robocopy 这类命令成功时返回 1~7，用上面那个「-eq 0」的判断会把成功误判为失败。
# 失败时返回 -1（命令未能启动）。
function Invoke-NativeExit([string]$FilePath, [string[]]$Arguments) {
    $prev = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        & $FilePath @Arguments 2>$null | Out-Null
        return $LASTEXITCODE
    } catch {
        return -1
    } finally {
        $ErrorActionPreference = $prev
    }
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

# 读 19825 上的 /api/health。返回对象表示「已有一个后端在正常应答」，$null 表示端口干净。
#
# 只测端口是不够的：端口有人不代表是本程序（可能是别的服务），也分辨不出是「旧进程
# 还没退干净」还是「另一个目录的实例」。这里取回 version / commit / buildTime，
# 报错时能直接告诉用户对面是谁。
#
# 用 HttpWebRequest 而非 Invoke-RestMethod：后者走 IE 代理设置，本机装了全局代理
# （Watt Toolkit 之类）时会连 127.0.0.1 也绕一圈，超时表现不稳定。
function Get-BackendHealth([int]$Port, [int]$TimeoutMs = 1200) {
    $prev = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        $req = [System.Net.HttpWebRequest]::Create("http://127.0.0.1:$Port/api/health")
        $req.Proxy   = $null
        $req.Timeout = $TimeoutMs
        $req.ReadWriteTimeout = $TimeoutMs
        $resp = $req.GetResponse()
        $sr = New-Object System.IO.StreamReader($resp.GetResponseStream())
        $body = $sr.ReadToEnd()
        $sr.Close()
        $resp.Close()
        $obj = $body | ConvertFrom-Json
        if ($obj -and $obj.status -eq 'ok') { return $obj }
        return $null
    } catch {
        return $null
    } finally {
        $ErrorActionPreference = $prev
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

# 只结束本脚本自己拉起来的进程，**不按端口**动别人。
# 用于「启动被握手拦下」的场景：19825 上那个后端不是这次起的，绝不能连坐清掉。
function Stop-OwnProcs {
    foreach ($item in $script:procs) {
        if ($item.Proc -and -not $item.Proc.HasExited) {
            $null = Stop-ProcessTree $item.Proc.Id
        }
    }
}

function Stop-All {
    Stop-OwnProcs
    foreach ($port in @($BackendPort, $FrontendPort, $BridgePort)) { Stop-PortOwners $port }
}

# ---------------------------------------------------------------- 伴生进程清单
# 后端退出时（界面「退出程序」/ Ctrl+C / 被结束）会读这份清单，把 Vite 与番茄桥
# 一起结束 —— 用户口径是「关闭后端 == 关闭整个程序」。
#
# 为什么不只靠本脚本的 finally：finally 只在**脚本自己**正常退出时执行。
# 用户直接点关闭按钮或结束 powershell，子进程就成孤儿占着端口，
# 下次启动撞端口、用户以为「程序没关干净」。清单把这份保障交给后端，与脚本存活解耦。
$script:manifestPath = Join-Path $logDir '.dev-companions.json'

function Write-CompanionManifest {
    $items = @()
    foreach ($item in $script:procs) {
        # 只登记长驻的伴生服务；backend 自己的退出由本函数触发，不必登记。
        if ($item.Name -ne 'vite' -and $item.Name -ne 'fanqie') { continue }
        $p = $item.Proc
        if (-not $p) { continue }
        # 已退出的进程拿 .Id 会抛异常，先跳过（它也没什么可清理的）。
        if ($p.HasExited) { continue }

        # ⚠ PS 5.1 不支持在 hashtable 字面量里直接写 if 表达式，必须先算好再放。
        $exeName = 'node.exe'
        $exePort = $FrontendPort
        if ($item.Name -eq 'fanqie') { $exeName = 'python.exe'; $exePort = $BridgePort }

        $items += [pscustomobject]@{
            name = $item.Name
            pid  = $p.Id
            exe  = $exeName
            port = $exePort
        }
    }
    $payload = [pscustomobject]@{
        created    = (Get-Date).ToUniversalTime().ToString('o')
        companions = @($items)
    }
    try {
        # ⚠ 必须写成**无 BOM** 的 UTF-8。
        # PS 5.1 的 `Set-Content -Encoding UTF8` 会加 BOM（它没有 UTF8NoBOM 选项，
        # 那是 PS 6+ 才引入的），而后端用 encoding/json 解析，BOM 会让它直接报错 ——
        # 结果是清单看着完全正常，后端却永远不清理（真实踩过，排查成本极高）。
        $json = $payload | ConvertTo-Json -Depth 4
        [System.IO.File]::WriteAllText(
            $script:manifestPath, $json, (New-Object System.Text.UTF8Encoding($false)))
        Write-Note "companion manifest: $($items.Count) 个伴生进程 -> $(Split-Path -Leaf $script:manifestPath)"
    } catch {
        Write-Warn "写入伴生进程清单失败（后端将无法自动清理）: $($_.Exception.Message)"
    }
}

function Remove-CompanionManifest {
    # 清单是一次性的。正常退出时本脚本已 Stop-All，清单已无用途；
    # 若后端先退出并已清理，它会自己删掉这份文件，此处静默跳过。
    if (Test-Path -LiteralPath $script:manifestPath) {
        Remove-Item -LiteralPath $script:manifestPath -Force -ErrorAction SilentlyContinue
    }
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

# 2.5 确认 19825 真的空出来了 ----------------------------------------------
# 这一步**不是**版本握手（开发版已经不参与握手了），纯粹是 tcp 层面的复查：
# 上一步 Stop-PortOwners 如果没杀掉（进程权限、卡在 IO），后面 Wait-ForPort 会因为
# 旧进程还在应答而**误判成功** → 打开浏览器 → 守护循环才发现新起的 backend 已退出。
# 与其拉一套假的再自我了断，不如提前报错：一个进程都不起。
$lingering = $null
for ($i = 0; $i -lt 6; $i++) {
    $lingering = Get-BackendHealth $BackendPort
    if (-not $lingering) { break }
    Start-Sleep -Milliseconds 500
}
if ($lingering) {
    Write-Bad "端口 $BackendPort 上仍有一个后端在应答，本次启动取消（没有关闭任何已有服务）"
    $verText  = if ($lingering.version)   { $lingering.version }   else { '未知（未上报版本号）' }
    $btText   = if ($lingering.buildTime) { $lingering.buildTime } else { 'unknown' }
    $cmText   = if ($lingering.commit)    { $lingering.commit }    else { 'unknown' }
    Write-Note "占用者：version=$verText  commit=$cmText  buildTime=$btText"
    foreach ($ownerPid in @(Get-PortOwners $BackendPort)) {
        $ownerName = (Get-Process -Id $ownerPid -ErrorAction SilentlyContinue).ProcessName
        Write-Note "pid $ownerPid ($ownerName)"
    }
    Write-Note '处理：先在那个环境里点「退出程序」，或在任务管理器结束该 xingmang.exe，然后重跑本脚本'
    exit 1
}

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

    # 前端产物同步到 embed 目录，保证 exe 里的页面是最新的。
    #
    # ⚠ 必须校验结果，两个坑都踩过：
    #   1. robocopy 成功时返回 1~7（1=复制了文件、2=删除了多余文件…），只有 >=8 才是失败，
    #      所以不能用 Invoke-Native 的「exit code -eq 0」来判断；
    #   2. 这里曾经多写了一个并不存在的参数 /NBT，robocopy 直接以 16 退出、一个文件都没复制，
    #      而返回值被 `$null =` 丢弃 → exe 长期内嵌旧前端，
    #      表现就是「明明更新了，界面/内容还是上一版」。
    # 因此改为：读退出码 + 同步后比对首页指纹，任何异常都打 [WARN] 而不是静默通过。
    $srcDist = Join-Path $root 'web\dist'
    $dstDist = Join-Path $root 'cmd\xingmang\web\dist'
    if (Test-Path -LiteralPath $srcDist) {
        Write-Note 'syncing web\dist -> cmd\xingmang\web\dist'
        $rc = Invoke-NativeExit 'robocopy.exe' @($srcDist, $dstDist, '/MIR', '/NFL', '/NDL', '/NJH', '/NJS', '/NP')
        if ($rc -lt 0) {
            Write-Warn 'robocopy 未能启动，跳过同步（exe 将内嵌现有的 embed 产物）'
        } elseif ($rc -ge 8) {
            Write-Warn "同步失败 (robocopy rc=$rc)，exe 内嵌的界面可能不是最新的"
        } else {
            # 退出码只代表命令自身没报错，再比对首页指纹确认内容真的换过来了
            $srcIdx = Join-Path $srcDist 'index.html'
            $dstIdx = Join-Path $dstDist 'index.html'
            $same = $false
            if ((Test-Path -LiteralPath $srcIdx) -and (Test-Path -LiteralPath $dstIdx)) {
                $h1 = (Get-FileHash -LiteralPath $srcIdx -Algorithm MD5).Hash
                $h2 = (Get-FileHash -LiteralPath $dstIdx -Algorithm MD5).Hash
                $same = ($h1 -eq $h2)
            }
            if ($same) { Write-Note "embed 产物已同步 (rc=$rc)" }
            else       { Write-Warn '同步后 index.html 与源不一致，exe 内嵌的界面可能不是最新的' }
        }
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

# 6. 落盘伴生进程清单（供后端退出时同步清理）-----------------------------
# 必须在三个服务都拉起来之后写：此时 PID 才是最终有效的。
Write-CompanionManifest

# 7. 结果面板 -------------------------------------------------------------
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

# 8. 守护：任一服务挂掉就整体退出，不再每秒刷屏 ----------------------------
Write-Host ''
Write-Host '  all services up. Press Ctrl+C to stop everything.' -ForegroundColor Cyan

try {
    while ($true) {
        foreach ($item in $script:procs) {
            if ($item.Proc -and $item.Proc.HasExited) {
                Write-Host ''
                if ($item.Name -eq 'backend') {
                    # 后端退出 = 「关闭整个程序」的正常路径（界面「退出程序」或 Ctrl+C）。
                    # 它会在收尾时读伴生清单，把 Vite / 番茄桥一起结束。
                    #
                    # ⚠ 这里不能按「谁先退出」判断：后端是**先杀伴生进程、后结束自己**，
                    # 本循环 500ms 轮询往往先看到 fanqie/vite 消失（它们被后端带走）。
                    # 所以只认「退出的是不是因为…」没用，直接以 backend 的退出作为唯一信号，
                    # 否则用户点一次「退出程序」就会看到一屏 [FAIL] 和一个非零退出码。
                    $exitCode = -1
                    try { $exitCode = $item.Proc.ExitCode } catch { }
                    if ($exitCode -eq 3) {
                        # 3 = 后端握手判定「19825 上已有另一个实例」，本次没启动服务就退了。
                        # 这既不是故障，也不是用户点了「退出程序」：那个后端不是这次起的，
                        # 绝不能 Stop-All 按端口连坐 —— 那会把别人正常跑着的环境一起掐断。
                        #
                        # ⚠ 正常情况下走不到这里：本脚本编出来的是**开发版**（version=dev），
                        # 而开发版在 cmd/xingmang/main.go 里整段跳过版本握手（isDevBuild）。
                        # 只有 -NoRebuild 沿用了一个发布版 exe 时才可能命中，故保留。
                        Write-Warn '启动被握手拦下：已有另一个后端在运行，本次未启动服务'
                        Show-LogTail $item.Err 8
                        $script:launchRejected = $true
                    } elseif ($exitCode -eq 4) {
                        # 4 = 后端自己试占端口失败（main.go 的 exitPortBusy）。
                        # 端口被占死了，服务根本没起来。这**不是**「用户点了退出程序」，
                        # 必须整体收掉并让退出码非 0；否则会被下面的 else 当成正常收尾，
                        # 静默 exit 0，留下一个只有半套服务的假环境。
                        Write-Bad "端口被占用，后端未能启动"
                        Show-LogTail $item.Err
                        $script:failed = $true
                    } else {
                        Write-Note "backend 已退出 (exit $exitCode)，正在同步结束其余服务"
                        if ($exitCode -ne 0) {
                            # 非零码说明后端是被异常终止的，给点线索便于排查
                            Show-LogTail $item.Err
                        }
                        $script:backendStopped = $true
                    }
                } else {
                    # 非 backend 退出，有两种可能，必须区分开：
                    #   a) 后端正在收尾并把它同步带走 —— 这是正常路径；
                    #   b) 它自己崩了 —— 这才是故障。
                    #
                    # ⚠ 关键时序：后端是**先杀伴生进程、最后结束自己**，
                    # 本循环 500ms 轮询往往先看到 fanqie/vite 消失，而那时 backend 尚未退出。
                    # 所以不能立刻判 FAIL，否则用户每点一次「退出程序」都会看到一屏红字。
                    # 给一个宽限期再复查 backend —— 它的清理是毫秒级的，2 秒足够。
                    Start-Sleep -Seconds 2
                    $backendExited = $false
                    foreach ($peer in $script:procs) {
                        if ($peer.Name -eq 'backend' -and $peer.Proc -and $peer.Proc.HasExited) {
                            $backendExited = $true
                        }
                    }
                    if ($backendExited) {
                        Write-Note "$($item.Name) 已由后端退出同步结束"
                        $script:backendStopped = $true
                    } else {
                        Write-Bad "$($item.Name) exited unexpectedly"
                        Show-LogTail $item.Err
                        Show-LogTail $item.Out 25
                        $script:failed = $true
                    }
                }
                break
            }
        }
        if ($script:failed -or $script:backendStopped -or $script:launchRejected) { break }
        Start-Sleep -Milliseconds 500
    }
} finally {
    Write-Host ''
    if ($script:launchRejected) {
        # 握手拒绝了本次启动。19825 上那个后端不是这次的，只清掉本脚本多起的进程，
        # **不按端口**去动别人的服务，避免一次误启动把整台开发环境端掉。
        Write-Host '  cleaning up processes started by this run only...' -ForegroundColor Yellow
        Stop-OwnProcs
        Remove-CompanionManifest
    } else {
        Write-Host '  shutting down...' -ForegroundColor Yellow
        Stop-All
        Remove-CompanionManifest
    }
    Write-Host '  done.' -ForegroundColor Green
}

if ($script:failed -or $script:launchRejected) { exit 1 }
exit 0
