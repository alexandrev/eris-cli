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
echo *** Docker API information
echo.
docker version

echo.
echo *** Checking Eris to Docker connection
echo.
eris init -dp --yes

:: Pull images if run without the 'local' parameter.
if x%1 == xlocal goto nopull
for /f "tokens=*" %%i in ('eris version --quiet') do set ERIS_VERSION=%%i
echo.
echo *** ERIS_VERSION=%ERIS_VERSION%
echo.

call :pull quay.io/eris/base
call :pull quay.io/eris/data
call :pull quay.io/eris/ipfs
call :pull quay.io/eris/keys
call :pull quay.io/eris/erisdb:%ERIS_VERSION%
call :pull quay.io/eris/epm:%ERIS_VERSION%
:nopull

echo.
echo.
echo *** Package tests
echo.
echo.

set ERIS_PULL_APPROVE=true
set ERIS_MIGRATE_APPROVE=true

eris services start ipfs

for /f "tokens=*" %%i in ('docker inspect --format="{{.NetworkSettings.IPAddress}}" eris_service_ipfs_1') do set ERIS_IPFS_HOST=http^://%%i
echo.
echo *** ERIS_IPFS_HOST=%ERIS_IPFS_HOST%
echo.

go test -v ./perform/...
call :passed Perform %errorlevel%

go test -v ./util/...
call :passed Util %errorlevel%

go test -v ./data/...
call :passed Data %errorlevel%

go test -v ./files/...
call :passed Config %errorlevel%

go test -v ./keys/...
call :passed Keys %errorlevel%

go test -v ./services/...
call :passed Services %errorlevel%

go test -v ./chains/...
call :passed Chains %errorlevel%

go test -v ./actions/...
call :passed Actions %errorlevel%

go test -v ./contracts/...
call :passed Contracts %errorlevel%

echo.
echo.
echo *** Congratulations! All Package Level Tests Passed.
echo.
echo.
exit /b

:pull
echo Pulling image %1
docker pull %1 >nul:
goto :eof

:passed
if %2 equ 0 (
        echo.
        echo *** Congratulations! *** %1 Package Level Tests Have Passed
        echo.
) else (
        echo.
        echo *** Boo :^( A Package Level Test has failed.
        echo.
)
goto :eof
