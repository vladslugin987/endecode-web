# 🚀 ENDECode + WordPress Integration

**Complete photo processing automation system with WordPress/WooCommerce integration**

## 🌟 Features

### 🎯 Core Functionality
- **Invisible watermarking** for photos and videos
- **Visible watermarking** with customizable text and positions
- **Smart file swapping** based on order numbers (desktop app logic)
- **Batch processing** with automatic numbering
- **ZIP archive generation** with custom naming (Collection-OrderXXX.zip)

### 🛒 E-commerce Integration
- **WordPress/WooCommerce** full integration
- **Automatic order processing** upon payment completion
- **Admin approval workflow** with status tracking
- **Collection-based product management**
- **Direct file upload** from WordPress admin panel

### 📧 Notification System
- **Telegram notifications** for administrators
- **Email notifications** (SMTP) for customers and admins
- **Real-time status updates** throughout the workflow
- **Webhook integration** with WordPress

### 💰 Subscription Management
- **4 subscription tiers**: Free, Basic ($29.99), Pro ($89.99), Enterprise ($299.99)
- **Cryptocurrency payments** (BTC, ETH, USDT) via CoinGate
- **Usage tracking** and limits enforcement
- **Admin panel** for subscription management

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   WordPress     │    │   ENDECode      │    │  Notification   │
│   WooCommerce   │◄──►│   API Server    │◄──►│   Services      │
│                 │    │                 │    │  (Email/Telegram)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
    ┌─────────┐            ┌─────────┐            ┌─────────┐
    │  MySQL  │            │  Redis  │            │  SMTP   │
    │Database │            │ Cache   │            │ Gateway │
    └─────────┘            └─────────┘            └─────────┘
```

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Git

### 1. Clone Repository
```bash
git clone https://github.com/vladslugin987/endecode-web.git
cd endecode-web
```

### 2. Start ENDECode Services
```bash
docker-compose up -d
```

### 3. Start WordPress
```bash
cd wordpress-setup
docker-compose up -d
```

### 4. Configure WordPress
1. Open **WordPress Admin**: http://localhost:8082/wp-admin
2. Activate **ENDECode Integration** and **WooCommerce** plugins
3. Configure **ENDECode API**: Settings → ENDECode
   - API URL: `http://endecode-compose-main-photo-processing-web-1:8080`
   - Test connection ✓

### 5. Create Products
1. **Products → Add New**
2. Configure **Photo Processing Settings**:
   - Enable Photo Processing ✓
   - Source Photo Folder: `Collection1`
   - Add Invisible Watermarks ✓
   - Use Order Number ✓ (for "ORDER XXX" encoding)
   - Automatic Swap by Order ✓
   - Create ZIP Archive ✓

## 📋 Complete Workflow

### Customer Experience
1. **Browse products** on your WordPress store
2. **Purchase photo collection** (e.g., wedding, portrait session)
3. **Receive confirmation** with status tracking
4. **Get notified** when photos are ready for download
5. **Download encrypted photos** with watermarked protection

### Admin Experience
1. **Receive notifications** (Telegram + Email) for new orders
2. **Review order** in WordPress admin panel
3. **Upload photos** directly from computer (if needed)
4. **Preview processed photos** in admin interface
5. **Approve order** to generate download link
6. **Customer automatically receives** download link

## 🎛️ Advanced Features

### File Swap Logic (Desktop App Compatible)
- **Order 001**: photo 1 ↔ photo 11
- **Order 002**: photo 2 ↔ photo 12
- **Order 003**: photo 3 ↔ photo 13
- Logic: `baseNumber + 10` (exactly like desktop version)

### Watermarking Options
- **Invisible watermarks**: Always applied (customer name or order number)
- **Visible watermarks**: Optional text overlay on specific photos
- **Position control**: Specify which photo numbers get watermarked
- **Auto-text**: Uses order number if no custom text provided

### Collection Management
```
/uploads/OriginalFiles/
├── Collection1/ (portrait sessions)
├── Collection2/ (wedding packages)
├── Collection3/ (fashion shoots)
└── Collection4/ (events)
```

## 📧 Notification Setup

### Email (SMTP)
Create `.env` file:
```bash
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
ADMIN_EMAIL=admin@yourcompany.com
```

### Telegram
1. Create bot with @BotFather
2. Get bot token and chat ID
3. Add to `.env`:
```bash
TELEGRAM_BOT_TOKEN=1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZ123456789
TELEGRAM_CHAT_ID=-1001234567890
```

**Complete setup guide**: [NOTIFICATION_SETUP.md](NOTIFICATION_SETUP.md)

## 🌐 Available Services

| Service | URL | Description |
|---------|-----|-------------|
| ENDECode API | http://localhost:8090 | Main photo processing service |
| ENDECode Admin | http://localhost:8090 | Admin panel for job management |
| WordPress Site | http://localhost:8082 | Customer-facing store |
| WordPress Admin | http://localhost:8082/wp-admin | Store administration |
| phpMyAdmin | http://localhost:8081 | Database management |

## 📚 Documentation

- **[NEXT_STEPS_GUIDE.md](NEXT_STEPS_GUIDE.md)** - Complete setup and testing guide
- **[NOTIFICATION_SETUP.md](NOTIFICATION_SETUP.md)** - Email and Telegram configuration  
- **[INTEGRATION_COMPLETE.md](INTEGRATION_COMPLETE.md)** - Technical implementation details
- **[wordpress-setup/README.md](wordpress-setup/README.md)** - WordPress-specific documentation

## 🔧 Troubleshooting

### Common Issues

**Upload Failed (400 Bad Request)**
```bash
# Check PHP limits
docker exec endecode-wordpress php -i | grep upload_max_filesize
# Should show: 1024M
```

**API Connection Failed**
```bash
# Try alternative URL
http://endecode-compose-main-photo-processing-web-1:8080
```

**Missing Products**
- Ensure WooCommerce is activated
- Check WooCommerce → Settings → Advanced → Page setup
- Set Homepage to show Shop page

## 💼 Production Deployment

### Security Checklist
- [ ] Change default WordPress admin credentials
- [ ] Configure SSL certificates
- [ ] Set up proper backup system
- [ ] Configure production SMTP settings
- [ ] Set strong API tokens
- [ ] Review file permissions

### Scaling Options
- Use external MySQL database
- Implement Redis clustering
- Set up load balancing
- Configure CDN for static assets

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👤 Author

**vsdev. | Vladislav Slugin**
- Email: vslugin@vsdev.top
- GitHub: [@vladslugin987](https://github.com/vladslugin987)

## 🙏 Acknowledgments

- Built with Go, React, TypeScript, and Docker
- WordPress integration via custom plugin
- Notification system using Telegram Bot API and SMTP
- Photo processing using advanced encoding algorithms

---

**⭐ Star this repository if it helped you!**
