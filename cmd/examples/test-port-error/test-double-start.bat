@echo off
echo Starting first instance in background...
start /B "" cmd /c "cd /d c:\Users\prima\SynologyDrive\golang\lokstra-dev2\cmd\examples\25-single-binary-deployment && go run main.go > nul 2>&1"

echo Waiting 5 seconds for server to start...
timeout /t 5 /nobreak > nul

echo.
echo Now starting second instance (should fail with nice error)...
echo ========================================================
cd /d c:\Users\prima\SynologyDrive\golang\lokstra-dev2\cmd\examples\25-single-binary-deployment
go run main.go
