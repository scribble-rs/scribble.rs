@echo off

echo where.exe go.exe | find "go.exe" > nul
IF ERRORLEVEL 1 (
	echo 'go' is required. & echo Please install it.
) else (
	go build -o scribblers.exe .
)
