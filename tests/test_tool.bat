@echo off
:: ---------------------------------------------------------------------------
:: PURPOSE

:: This script will test the Eris command line tool locally.
:: If started with the optional argument "local", it doesn't pull Docker images
:: necessary to test the tool.

:: ---------------------------------------------------------------------------
:: REQUIREMENTS

:: Eris tool, Docker, and Docker Machine installed locally.
:: Docker Machine is expected to be running and its environment
:: ('docker-machine env') set.

:: ---------------------------------------------------------------------------
:: USAGE

:: test_tool.bat [local]

echo Hello! The marmots will begin testing now.
echo.
echo.
echo Docker API information
echo.
docker version

echo.
echo Checking the Eris to Docker connection
echo.
eris init -dp --yes
echo.
eris version
echo.

:: Pull images if run without the 'local' parameter.
if x%1 == xlocal goto nopull
for /f "tokens=*" %%i in ('eris version --quiet') do set ERIS_VERSION=%%i
echo.
echo ERIS_VERSION=%ERIS_VERSION%
echo.

call :pull quay.io/eris/base
call :pull quay.io/eris/data
call :pull quay.io/eris/ipfs
call :pull quay.io/eris/keys
call :pull quay.io/eris/erisdb:%ERIS_VERSION%
call :pull quay.io/eris/epm:%ERIS_VERSION%
:nopull

:: Actual tests.
set ERIS_PULL_APPROVE=true
set ERIS_MIGRATE_APPROVE=true

go test ./perform/...
call :passed Perform %errorlevel%

go test ./util/...
call :passed Util %errorlevel%

go test ./data/...
call :passed Data %errorlevel%

go test ./files/...
call :passed Config %errorlevel%

go test ./keys/...
call :passed Keys %errorlevel%

go test ./services/...
call :passed Services %errorlevel%

go test ./chains/...
call :passed Chains %errorlevel%

go test ./actions/...
call :passed Actions %errorlevel%

go test ./contracts/...
call :passed Contracts %errorlevel%

echo.
echo Congratulations! All Package Level Tests Passed.
echo.
exit /b


:pull
echo Pulling image =^> %1
docker pull %1 >nul:
goto :eof

:passed
if %2 equ 0 (
        echo *** Congratulations! ***  %1 Package Level Tests Have Passed
) else (
        echo Boo :^( A Package Level Test has failed.
        exit /b 1
)
goto :eof
