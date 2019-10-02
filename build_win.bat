@echo off

echo Building Gargoyle Master...
windres gymaster/winappres.rc -o gymaster/winappres.syso
go build -v -i -ldflags="-s -w" -o ./work/gymaster.exe ./gymaster
del gymaster\*.syso

echo Building Gargoyle Slave...
windres gyslave/winappres.rc -o gyslave/winappres.syso
go build -v -i -ldflags="-s -w" -o ./work/gyslave.exe ./gyslave
del gyslave\*.syso

echo Done! Press any key to exit...
pause>nul