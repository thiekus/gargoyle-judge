@echo off

set "APP_NAME=Gargoyle Judge Server PP2"
set "WIX_DIR=C:\Program Files (x86)\WiX Toolset v3.11\bin"
set "MSI_NAME=GargoyleJudgeWin_setup"

del /Q obj\*.wixobj
del /Q *.wixpdb

echo * Compiling MSI Installer *
echo.

echo Executing Heat...
"%WIX_DIR%\heat.exe" dir "..\work\assets" -cg AssetsDir -gg -sfrag -template:fragment -dr HeatFiles -var var.SourceDir -out dir_assets.wxs
"%WIX_DIR%\candle.exe" -nologo dir_assets.wxs -dSourceDir=..\work\assets -o obj\

"%WIX_DIR%\heat.exe" dir "..\work\templates" -cg TemplatesDir -gg -sfrag -template:fragment -dr HeatFiles -var var.SourceDir -out dir_templates.wxs
"%WIX_DIR%\candle.exe" -nologo dir_templates.wxs -dSourceDir=..\work\templates -o obj\

"%WIX_DIR%\heat.exe" dir "..\work\lang" -cg LangDir -gg -sfrag -template:fragment -dr HeatFiles -var var.SourceDir -out dir_lang.wxs
"%WIX_DIR%\candle.exe" -nologo dir_lang.wxs -dSourceDir=..\work\lang -o obj\

echo Executing Candle for main script...
"%WIX_DIR%\candle.exe" -nologo "%MSI_NAME%.wxs" "-dAppName=%APP_NAME%" -o obj\ -ext WixUIExtension -ext WiXUtilExtension

echo.
echo Executing Light...
"%WIX_DIR%\light.exe" -nologo -b "D:\projects\gargoyle-judge\work" "obj\*.wixobj" -o "%MSI_NAME%.msi" -ext WixUIExtension -ext WiXUtilExtension

pause