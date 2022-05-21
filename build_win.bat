@echo off

echo Building Gargoyle Master...
rem windres gymaster/winappres.rc -o gymaster/winappres.syso
go build -v -ldflags="-s -w" -o ./bin/gymaster.exe ./gymaster
rem del gymaster\*.syso

echo Building Gargoyle Slave...
rem windres gyslave/winappres.rc -o gyslave/winappres.syso
go build -v -ldflags="-s -w" -o ./bin/gyslave.exe ./gyslave
rem del gyslave\*.syso

echo Done! Press any key to exit...
pause>nul