@echo off
REM updatectl Windows Installer

echo Installing updatectl...

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Go is not installed. Please install Go first.
    pause
    exit /b 1
)

REM Build the binary
echo Building updatectl...
go build -o updatectl.exe main.go

REM Create install directory
if not exist "C:\Program Files\updatectl" mkdir "C:\Program Files\updatectl"

REM Copy binary
copy updatectl.exe "C:\Program Files\updatectl\"

echo updatectl installed successfully to C:\Program Files\updatectl\
echo Make sure C:\Program Files\updatectl is in your PATH.
echo Run 'updatectl init' to set up the daemon.

pause