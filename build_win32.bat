@echo off

set "PATH=C:\TDM-GCC-32\bin;C:\Go32\bin;%PATH%"
set "GOROOT=C:\Go32"

echo Building Gargoyle Master...
windres gymaster/winappres.rc -o gymaster/winappres.syso
go build -v -i -ldflags="-s -w" -o ./bin/gymaster.exe ./gymaster
del gymaster\*.syso

echo Building Gargoyle Slave...
windres gyslave/winappres.rc -o gyslave/winappres.syso
go build -v -i -ldflags="-s -w" -o ./bin/gyslave.exe ./gyslave
del gyslave\*.syso

echo Done! Press any key to exit...
pause>nul