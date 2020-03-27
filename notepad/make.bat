@ECHO OFF

REM git.exe checkout dev/notepad
REM git.exe pull --all

del notepad.exe
go.exe build -ldflags="-s -w -H=windowsgui" -o notepad.exe .
