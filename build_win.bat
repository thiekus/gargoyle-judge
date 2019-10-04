@echo off

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