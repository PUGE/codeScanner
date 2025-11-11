@echo off
echo 开始跨平台编译...

echo 编译 Windows 版本...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o dist/code-scanner.exe

echo 编译 Linux 版本...
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -ldflags="-s -w" -o dist/code-scanner-linux

echo 编译 Mac 版本...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o dist/code-scanner-mac

echo 编译完成！
dir dist
pause