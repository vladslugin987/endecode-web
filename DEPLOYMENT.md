# 🚀 Photo Processing Web Server - Deployment Guide

This guide explains how to run the Photo Processing Web Server without installing Go, Node.js, or any development tools locally. Everything runs in Docker containers.

## 📋 Prerequisites

The only requirement is **Docker** installed on your system:

- **Windows**: [Docker Desktop for Windows](https://docs.docker.com/docker-for-windows/install/)
- **macOS**: [Docker Desktop for Mac](https://docs.docker.com/docker-for-mac/install/)
- **Linux**: [Docker Engine](https://docs.docker.com/engine/install/)

## 🎯 Quick Start

### Option 1: Automated Installation (Recommended)

**Linux/macOS:**
```bash
chmod +x install.sh
./install.sh
```

**Windows:**
Double-click `install.bat` or run in Command Prompt:
```cmd
install.bat
```

The automated installer will:
1. Check Docker installation
2. Create necessary directories
3. Build and start the application
4. Show you how to access it

### Option 2: Manual Installation

1. **Create data directories:**
   ```bash
   mkdir -p data/{photos,processed,temp}
   ```

2. **Start the application:**
   ```bash
   docker compose up -d --build
   ```

3. **Wait for build to complete** (first run takes 3-5 minutes)

4. **Access the application:** Open http://localhost:8090 in your browser

## 🌐 Using the Web Interface

Once running, you can access the Photo Processing Web Server at:
**http://localhost:8090**

### Features Available:

1. **File Upload**: Drag & drop or click to select photos
2. **Text Encryption**: Add encrypted text to photos using Caesar cipher
3. **Watermark Management**: Add/remove binary watermarks
4. **Batch Processing**: Process multiple files at once
5. **Real-time Console**: See processing logs in real-time
6. **File Download**: Download processed files directly

### Directory Structure:
```
data/
├── photos/     ← Drop your input photos here (or use web interface)
├── processed/  ← Processed files appear here
└── temp/       ← Temporary files (auto-cleaned)
```

## 🛠️ Management Commands

### View Application Status
```bash
docker compose ps
```

### View Real-time Logs
```bash
docker compose logs -f
```

### Stop the Application
```bash
docker compose down
```

### Restart the Application
```bash
docker compose restart
```

### Update Application (rebuild with latest changes)
```bash
docker compose up -d --build
```

### Remove Everything (including data)
```bash
docker compose down -v
rm -rf data/
```

## 🔧 Configuration

### Environment Variables

You can customize the application by editing `docker-compose.yml`:

```yaml
environment:
  - LOG_LEVEL=info     # debug, info, warn, error
  - PORT=8080          # Change internal web port inside the container
```

### Port Configuration

To change the web port mapping (e.g., to 3000 on host):

1. Edit `docker-compose.yml`:
   ```yaml
   ports:
     - "3000:8080"  # Host:Container
   ```

2. Restart:
   ```bash
   docker compose up -d
   ```

3. Access at: http://localhost:3000

## 🐛 Troubleshooting

### Application Won't Start
1. Check Docker is running:
   ```bash
   docker --version
   ```

2. Check logs for errors:
   ```bash
   docker compose logs
   ```

3. Ensure host port is not in use (default 8090 in this compose):
   ```bash
   # Linux/macOS
   lsof -i :8090
   
   # Windows
   netstat -an | findstr :8090
   ```

### Build Failures
1. Clear Docker cache:
   ```bash
   docker system prune -a
   ```

2. Rebuild from scratch:
   ```bash
   docker compose down
   docker compose up -d --build --force-recreate
   ```

### Permission Issues (Linux/macOS)
If you get permission errors with the data directory:
```bash
sudo chown -R $USER:$USER data/
chmod -R 755 data/
```

### Web Interface Not Loading
1. Wait 1-2 minutes after starting (building takes time)
2. Check health status:
   ```bash
   docker compose ps
   ```
3. Look for "healthy" status in the output

## 📊 System Requirements

- **RAM**: 2GB minimum, 4GB recommended
- **Storage**: 1GB for Docker images + space for your photos
- **CPU**: Any modern processor (builds will be slower on older systems)
- **Network**: Internet connection for initial Docker image download

## 🔒 Security Notes

- The application runs on localhost only by default
- All processing happens locally - no data leaves your machine
- Files are stored in the local `data/` directory
- Use strong passwords if deploying to remote servers

## 🌍 Remote Access (Advanced)

To access from other devices on your network, edit `docker-compose.yml`:

```yaml
ports:
  - "0.0.0.0:8090:8080"  # Allow access from any IP
```

⚠️ **Warning**: Only do this on trusted networks!

## 📝 Architecture

The deployment includes:

1. **Go Backend**: Handles file processing, encryption, watermarks
2. **React Frontend**: Modern web interface built with TypeScript
3. **WebSocket**: Real-time communication for logs and progress
4. **Volume Mounts**: Persistent data storage on your machine

## 🔄 Updates

To update to a newer version:

1. Pull latest code (if using git):
   ```bash
   git pull
   ```

2. Rebuild and restart:
   ```bash
   docker compose up -d --build
   ```

## 💡 Tips

- **Batch Processing**: Upload multiple files at once for faster processing
- **File Formats**: Supports common image formats (JPG, PNG, GIF, BMP, TIFF)
- **Performance**: Larger images take longer to process
- **Storage**: Processed files accumulate in `data/processed/` - clean up periodically

## 🆘 Getting Help

If you encounter issues:

1. Check this documentation
2. Review the console logs in the web interface
3. Check Docker logs: `docker compose logs`
4. Ensure your system meets the requirements
5. Try the troubleshooting steps above

---

**Ready to start processing photos? Run the installer and open http://localhost:8090!** 🎉