@echo off
setlocal

echo Checking for go.mod file...
if not exist go.mod (
    echo Error: go.mod file not found in current directory. Please run this script in a Go project root directory.
    pause
    exit /b 1
)

echo Upgrading all dependencies to latest compatible versions (minor/patch)...
"D:\Program Files\go1.25.0\bin\go.exe" get -u ./...

echo Downloading all dependencies to local cache...
"D:\Program Files\go1.25.0\bin\go.exe" mod download

echo Cleaning go.mod and go.sum...
"D:\Program Files\go1.25.0\bin\go.exe" mod tidy

echo Generating/updating vendor directory...
"D:\Program Files\go1.25.0\bin\go.exe" mod vendor

echo.
echo Operation completed!
echo - go.mod and go.sum updated with latest compatible dependencies
echo - vendor directory generated (contains all required dependencies)
echo.
echo Note: To force upgrade to latest major versions (may introduce breaking changes), manually run: go get some/module@latest
pause