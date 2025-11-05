@echo off
setlocal enabledelayedexpansion

REM ================== 基本设置 ==================
set "START_PORT=58490"
set "ENTRY="
if exist "viewer_tool.html" set "ENTRY=viewer_tool.html"
if not defined ENTRY if exist "viewer_local.html" set "ENTRY=viewer_local.html"
if not defined ENTRY if exist "index.html" set "ENTRY=index.html"
if not defined ENTRY (
  echo [ERROR] 未找到入口文件（viewer_tool.html / viewer_local.html / index.html）
  pause
  exit /b 1
)

REM ================== 进入脚本目录 ==================
cd /d "%~dp0"

REM ================== 自动找可用端口 ==================
set /a PORT=%START_PORT%
:find_port
netstat -ano | find ":%PORT% " | find "LISTENING" >nul 2>&1
if %errorlevel%==0 (
  set /a PORT+=1
  goto :find_port
)

REM ================== 确认 http-server 可用（自动安装） ==================
where http-server >nul 2>&1
if errorlevel 1 (
  echo [INFO] 未检测到 http-server，正在安装（使用国内镜像）...
  call npm config set registry https://registry.npmmirror.com >nul 2>&1
  call npm i -g http-server >nul 2>&1
  if errorlevel 1 (
    echo [ERROR] 安装 http-server 失败；请检查 npm 与网络后重试。
    pause
    exit /b 1
  )
)

REM ================== 放通防火墙端口 ==================
REM 若规则已存在则忽略错误
netsh advfirewall firewall show rule name="DBViewer %PORT%" >nul 2>&1
if errorlevel 1 (
  netsh advfirewall firewall add rule name="DBViewer %PORT%" dir=in action=allow protocol=TCP localport=%PORT%" >nul 2>&1
)

REM ================== 枚举本机 IPv4 地址（排除 127.0.0.1 / 169.254.x.x） ==================
set "IPS="
for /f "usebackq delims=" %%I in (`powershell -NoProfile -Command ^
  "$ips = Get-NetIPAddress -AddressFamily IPv4 | Where-Object { $_.IPAddress -ne '127.0.0.1' -and $_.IPAddress -notmatch '^169\.254\.' } | Select-Object -ExpandProperty IPAddress; $ips -join ' '"`) do (
  set "IPS=%%I"
)

REM ================== 打印访问信息 ==================
echo.
echo ============================================================
echo  DragonBones Viewer is starting...
echo  Entry   : %ENTRY%
echo  Port    : %PORT%
echo  Local   : http://127.0.0.1:%PORT%/%ENTRY%
for %%a in (%IPS%) do echo  Share   : http://%%a:%PORT%/%ENTRY%
echo ------------------------------------------------------------
echo  提示：
echo   1) 保持本窗口开启，別的电脑才能访问
echo   2) 别的电脑请访问上面的 Share 链接（不是 127.0.0.1）
echo   3) 如访问失败：检查是否同一局域网 / 关闭 VPN / 路由器未开启 AP 隔离
echo ============================================================
echo.

REM ================== 打开本机页面 ==================
start "" "http://127.0.0.1:%PORT%/%ENTRY%"

REM ================== 启动服务（绑定所有网卡，禁用缓存） ==================
http-server -p %PORT% -a 0.0.0.0 -c-1

endlocal
