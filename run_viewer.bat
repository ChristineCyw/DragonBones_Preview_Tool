@echo off
cd /d "%~dp0"

REM 如果端口被占用自动递增
set PORT=58490
:check
netstat -ano | find ":%PORT% " >nul
if not errorlevel 1 (
  set /a PORT+=1
  goto check
)

echo Starting local viewer at port %PORT%...
start "" "http://127.0.0.1:%PORT%/viewer_tool.html"
DBViewerServer.exe --dir="%cd%" --entry=viewer_tool.html --port=%PORT%
pause
