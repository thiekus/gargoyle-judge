@echo off

set "PATH=C:\TDM-GCC-32\bin;C:\Go32\bin;%PATH%"
set "GOROOT=C:\Go32"

echo Building Gargoyle Master...
rem windres gymaster/winappres.rc -o gymaster/winappres.syso
go build -v -o ./bin/gymaster.exe ./gymaster
rem del gymaster\*.syso

echo Building Gargoyle Slave...
rem windres gyslave/winappres.rc -o gyslave/winappres.syso
go build -v -o ./bin/gyslave.exe ./gyslave
rem del gyslave\*.syso

echo Done! Press any key to exit...
pause>nul