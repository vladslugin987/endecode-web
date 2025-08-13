# ENDECode Web 🚀

Web-based photo processing application with watermarking and batch operations. Built with Go backend and React frontend.

## ✨ Features

- **🖼️ Photo Processing**: Encrypt/decrypt photos with invisible watermarks
- **📦 Batch Operations**: Create multiple copies with automatic numbering
- **🔍 Admin Panel**: Review processed photos, select archives, approve downloads
- **💧 Visible Watermarks**: Add text overlays to specific photos
- **📁 ZIP Creation**: Automatic archiving with clean naming
- **🎛️ Real-time Console**: Live progress updates via WebSocket
- **🔄 File Swapping**: Advanced file manipulation for security

## 🏗️ Architecture

- **Backend**: Go with OpenCV (gocv) for image processing
- **Frontend**: React + TypeScript with Tailwind CSS
- **Storage**: File-based with Redis for job queuing
- **Deployment**: Docker containerized for easy hosting

## 🚀 Quick Start

### Local Development
```bash
git clone https://github.com/vladslugin987/endecode-web.git
cd endecode-web
docker compose up -d
```

Visit: http://localhost:8090

### Production Deployment

#### Railway (Recommended)
1. Fork this repository
2. Connect to [Railway](https://railway.app)
3. Deploy from GitHub - automatic build with `Dockerfile.production`

#### Manual Docker
```bash
docker build -f Dockerfile.production -t endecode-web .
docker run -p 8080:8080 endecode-web
```

## 🎯 Usage

### Main Application
1. **Upload Photos**: Select folder or drag & drop
2. **Choose Operation**: 
   - Encrypt/Decrypt with invisible watermarks
   - Batch Copy with multiple numbered versions
   - Add visible text watermarks
3. **Monitor Progress**: Real-time console updates

### Admin Panel
1. **Review Jobs**: See all processing jobs
2. **Select Archives**: Choose which ZIP to preview (001, 002, 003...)
3. **Preview Images**: Check watermarks and quality
4. **Approve Downloads**: Generate secure download links

## ⚙️ Configuration

### Environment Variables
```bash
PORT=8080                    # Server port
API_TOKEN=your-secret       # Admin panel protection
UPLOADS_PATH=/app/uploads   # File storage path
```

### WooCommerce Integration (Coming Soon)
- Webhook receiver for automatic order processing
- Email notifications for admins and customers
- Time-limited download links

## 🛠️ Development

### Backend (Go)
```bash
cd photo-processing-server
go mod tidy
go run cmd/web/main.go
```

### Frontend (React)
```bash
cd photo-processing-server/web/frontend
npm install
npm run dev
```

## 📋 API Endpoints

### Core Operations
- `POST /api/encrypt` - Add invisible watermarks
- `POST /api/decrypt` - Extract watermarks
- `POST /api/batch-copy` - Batch processing
- `POST /api/upload` - File upload

### Admin Panel
- `GET /api/admin/jobs` - List processing jobs
- `GET /api/admin/jobs/:id/images` - Get archived images
- `POST /api/admin/jobs/:id/approve` - Approve for download

## 🎨 Screenshots

### Main Interface
- **Home**: File selector, operations panel, real-time console
- **Admin**: Job list, archive selector, image preview grid

### Key Features
- **Archive Selection**: Choose from multiple ZIP files (001, 002, 003...)
- **Watermark Zoom**: 6x pixelated preview of watermarked images
- **Progress Tracking**: Real-time WebSocket updates

## 🔒 Security Features

- **Invisible Watermarks**: Detect image leaks by embedded user data
- **File Swapping**: Advanced obfuscation techniques
- **Token-based Downloads**: Secure, time-limited access
- **Admin Authentication**: API token protection

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📞 Support

For questions or issues, please open a GitHub issue or contact the maintainer.

---

**Built with ❤️ using Go, React, and OpenCV**
