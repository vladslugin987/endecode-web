@echo off
REM Photo Processing Web Server Installation Script for Windows
REM This script sets up everything needed to run the photo processing web application

title Photo Processing Web Server Installer

echo ==========================================
echo   Photo Processing Web Server Installer
echo ==========================================
echo.

REM Check if Docker is installed
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker is not installed!
    echo Please install Docker Desktop for Windows first:
    echo https://docs.docker.com/docker-for-windows/install/
    echo.
    pause
    exit /b 1
)
echo [SUCCESS] Docker is installed

REM Check if Docker Compose is available
docker compose version >nul 2>&1
if %errorlevel% neq 0 (
    docker-compose --version >nul 2>&1
    if %errorlevel% neq 0 (
        echo [ERROR] Docker Compose is not available!
        echo Please ensure Docker Desktop includes Docker Compose
        echo.
        pause
        exit /b 1
    ) else (
        set DOCKER_COMPOSE=docker-compose
        echo [SUCCESS] Docker Compose is available
    )
) else (
    set DOCKER_COMPOSE=docker compose
    echo [SUCCESS] Docker Compose is available
)

REM Create data directories
echo [INFO] Creating data directories...
if not exist "data" mkdir data
if not exist "data\photos" mkdir data\photos
if not exist "data\processed" mkdir data\processed
if not exist "data\temp" mkdir data\temp
echo [SUCCESS] Data directories created

REM Start the application
echo [INFO] Starting Photo Processing Web Server...
echo [INFO] This may take a few minutes for the first run (building images)...
echo.

%DOCKER_COMPOSE% up -d --build

if %errorlevel% neq 0 (
    echo [ERROR] Failed to start the application
    echo Please check the logs: %DOCKER_COMPOSE% logs
    pause
    exit /b 1
)

echo [SUCCESS] Photo Processing Web Server is starting!
echo [INFO] Waiting for the application to be ready...

REM Wait a bit for the application to start
timeout /t 10 /nobreak >nul

echo.
echo ==================================
echo [SUCCESS] Installation Complete!
echo ==================================
echo.
echo [INFO] Your Photo Processing Web Server is now running!
echo.
echo Web Interface: http://localhost:8080
echo Drop photos in: .\data\photos\
echo Processed files: .\data\processed\
echo.
echo Useful commands:
echo   View logs:     %DOCKER_COMPOSE% logs -f
echo   Stop server:   %DOCKER_COMPOSE% down
echo   Restart:       %DOCKER_COMPOSE% restart
echo   Update:        %DOCKER_COMPOSE% up -d --build
echo.
echo Press any key to open the web interface...
pause >nul

REM Open web browser to the application
start http://localhost:8080