@echo off
setlocal
set "PORT=%~1"
if "%PORT%"=="" set "PORT=58490"

for /f "tokens=5" %%P in ('netstat -ano ^| find ":%PORT% " ^| find "LISTENING"') do (
  echo Killing PID %%P on port %PORT% ...
  taskkill /PID %%P /F >nul 2>&1
)
echo Done.
endlocal