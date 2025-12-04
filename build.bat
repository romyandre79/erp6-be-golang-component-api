@echo off
setlocal enabledelayedexpansion

echo Reading version from plugin.json...

for /f "tokens=2 delims=:," %%a in ('findstr /i "version" plugin.json') do (
    set raw=%%a
)

set raw=%raw:"=%
set version=%raw: =%

echo Current version: %version%

for /f "tokens=1,2,3 delims=." %%a in ("%version%") do (
    set major=%%a
    set minor=%%b
    set patch=%%c
)

set /a patch+=1
set new_version=%major%.%minor%.%patch%

echo New version: %new_version%

echo Updating plugin.json...

powershell -Command "(Get-Content plugin.json) -replace '\"version\": \"[0-9.]+\"', '\"version\": \"%new_version%\"' | Set-Content plugin.json"

echo Cleaning old build files...
if exist build (
    rmdir /s /q build
)
mkdir build

echo Building api Plugin for all OS...

set TARGETS=windows/amd64 windows/arm64 linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

for %%T in (%TARGETS%) do (
    for /f "tokens=1,2 delims=/" %%a in ("%%T") do (
        set GOOS=%%a
        set GOARCH=%%b

        echo Building for %%a/%%b...

        if "%%a"=="windows" (
            go build -o build/api_%%a_%%b.exe main.go
        ) else (
            go build -o build/api_%%a_%%b main.go
        )

        if !ERRORLEVEL! NEQ 0 (
            echo Build FAILED for %%a/%%b
            exit /b 1
        )
    )
)

echo Creating plugin ZIP...

set ZIPFILE=api_plugin_v%new_version%.zip

powershell Compress-Archive -Path plugin.json,build\* -DestinationPath %ZIPFILE% -Force

echo Deleting build folder...
rmdir /s /q build

echo.
echo Build complete: %ZIPFILE%
echo Upload using:
echo curl -X POST http://localhost:8888/api/plugins/upload -F "plugin=@%ZIPFILE%"
echo.
