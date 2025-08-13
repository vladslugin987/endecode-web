# EnDeCode Web UI

A web-based version of the EnDeCode photo processing application with the same functionality as the desktop Kotlin version.

## 🌐 Features

- **File Processing**: Encrypt/decrypt files with Caesar cipher encoding
- **Batch Operations**: Bulk copy with auto-numbering and watermarks
- **Watermark Management**: Add/remove visible and invisible watermarks
- **Real-time Console**: Live logging and progress updates via WebSocket
- **Drag & Drop**: Easy file/folder selection
- **Progress Tracking**: Visual progress bars for all operations

## 🏗️ Architecture

- **Frontend**: React 18 + TypeScript + Tailwind CSS
- **Backend**: Go HTTP Server + WebSocket
- **Real-time**: WebSocket for live logs and progress
- **State Management**: React Context + useReducer
- **Styling**: Tailwind CSS matching Material 3 design

## 📋 Prerequisites

- **Node.js** 18+ (for React frontend)
- **Go** 1.21+ (for backend server)
- **OpenCV** (for image processing)
- **Git** (for cloning)

## 🚀 Quick Start

### 1. Build the Application

```bash
# Make build script executable
chmod +x build-web.sh

# Run the build script
./build-web.sh
```

### 2. Run the Web Server

```bash
# Start the server
./bin/endecode-web-server
```

### 3. Access the Web UI

- **Web Interface**: http://localhost:8080
- **API Endpoints**: http://localhost:8080/api
- **WebSocket**: ws://localhost:8080/ws
- **Health Check**: http://localhost:8080/health

## 📁 Project Structure

```
photo-processing-server/
├── cmd/web-server/              # Main web server
├── internal/
│   ├── services/                # Core processing services
│   └── web/                     # Web handlers & WebSocket
├── web/frontend/                # React application
│   ├── src/
│   │   ├── components/          # React components
│   │   ├── contexts/           # State management
│   │   ├── hooks/              # Custom React hooks
│   │   ├── services/           # API services
│   │   └── types/              # TypeScript types
│   ├── dist/                   # Built React app
│   └── package.json
├── build-web.sh                # Build script
└── go.mod                      # Go dependencies
```

## 🎛️ Usage

### Main Interface

The web UI replicates the desktop version's 40/60 layout:

- **Left Panel (40%)**: Controls, buttons, settings
- **Right Panel (60%)**: Console with real-time logs

### File Operations

1. **Choose Folder**: Click "Choose folder with files" or drag & drop
2. **Enter Text**: Add name to inject for encoding
3. **Encrypt/Decrypt**: Process files with invisible watermarks
4. **View Progress**: Watch real-time progress and logs

### Batch Operations

1. **Batch Copy**: Create multiple copies with auto-numbering
2. **Add Text**: Add visible text to specific photos
3. **Remove Watermarks**: Delete invisible watermarks

### Settings

- **Auto-clear console**: Automatically clear logs on new operations
- **Progress tracking**: Real-time progress bars
- **Error handling**: Detailed error messages and recovery

## 🔧 Development

### Frontend Development

```bash
cd web/frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

### Backend Development

```bash
# Run Go server in development
go run cmd/web-server/main.go

# Or build and run
go build -o bin/endecode-web-server cmd/web-server/main.go
./bin/endecode-web-server
```

## 📡 API Endpoints

### Processing Operations
- `POST /api/encrypt` - Encrypt files
- `POST /api/decrypt` - Decrypt files  
- `POST /api/batch-copy` - Batch copy with settings
- `POST /api/add-text` - Add text to photo
- `POST /api/remove-watermarks` - Remove watermarks

### File Management
- `POST /api/upload` - Upload files
- `GET /api/processing/:id` - Check processing status
- `GET /api/download/:token` - Download results

### System
- `GET /health` - Health check
- `GET /api/info` - System information

## 🔌 WebSocket Events

### Client → Server
```typescript
{
  type: 'subscribe' | 'unsubscribe',
  jobId?: string
}
```

### Server → Client  
```typescript
{
  type: 'log' | 'progress' | 'complete' | 'error',
  data: {
    message?: string,
    progress?: number,
    jobId?: string,
    error?: string
  }
}
```

## 🎨 Styling

The web UI matches the desktop Material 3 design:

- **Colors**: Primary #1976D2, Secondary #2196F3
- **Typography**: Material Design typography scale
- **Spacing**: Consistent 4dp/8dp/16dp spacing
- **Components**: Cards, buttons, inputs styled to match

## 🔒 Security

- **Input Validation**: All inputs validated client and server-side
- **File Type Checking**: Only supported file types allowed
- **Error Handling**: Secure error messages without sensitive data
- **CORS**: Configured for development and production

## 🚢 Deployment

### Production Build

```bash
# Build both frontend and backend
./build-web.sh

# Run in production
./bin/endecode-web-server
```

### Docker (Optional)

```dockerfile
FROM node:18-alpine AS frontend
WORKDIR /app/web/frontend
COPY web/frontend/package*.json ./
RUN npm install
COPY web/frontend .
RUN npm run build

FROM golang:1.21-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/frontend/dist ./web/frontend/dist
RUN go build -o bin/endecode-web-server cmd/web-server/main.go

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=backend /app/bin/endecode-web-server .
COPY --from=backend /app/web/frontend/dist ./web/frontend/dist
EXPOSE 8080
CMD ["./endecode-web-server"]
```

## 🐛 Troubleshooting

### Common Issues

**React build fails:**
```bash
cd web/frontend
rm -rf node_modules package-lock.json
npm install
```

**Go build fails:**
```bash
go mod tidy
go mod download
```

**WebSocket connection fails:**
- Check if port 8080 is available
- Verify firewall settings
- Check CORS configuration

### Support

- **GitHub**: Report issues on the project repository
- **Contact**: vslugin@vsdev.top

## 📝 Version History

- **v2.1.1-Web**: Initial web UI release
- **v2.1.1**: Desktop version compatibility
- **v2.1.0**: Core processing features

---

**EnDeCode Web UI** - Bringing desktop photo processing to the web with the same powerful features and intuitive interface.