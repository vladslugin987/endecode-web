# План автоматизации системы кодирования фотографий для фотомодельной компании

## Текущая ситуация vs Целевая система

### Текущий процесс (ручной):
1. Клиент покупает → 2. Админ видит → 3. Админ кодирует → 4. Админ отправляет

### Целевой процесс (автоматизированный):
1. Клиент покупает → 2. Автокодирование → 3. Уведомление админу → 4. Проверка → 5. Автоотправка

---

## Архитектура системы

### Общая схема компонентов:
```
WooCommerce Store
        ↓ (webhook)
Processing Service ← → File Storage
        ↓
Admin Panel ← → Notification Service
        ↓
Download Service ← → Temporary Links DB
```

### Детальная архитектура:

#### 1. **WooCommerce Integration Layer**
- **Webhook Receiver** - принимает уведомления о покупках
- **Order Processor** - извлекает информацию о заказе и файлах
- **Queue Manager** - управляет очередью обработки

#### 2. **File Processing Service** (портированная логика с Kotlin)
- **File Manager** - копирование и организация файлов
- **Encoding Engine** - водяные знаки и шифрование
- **Image Processor** - обработка изображений (OpenCV)
- **Video Processor** - обработка видео файлов
- **Archive Creator** - создание ZIP архивов

#### 3. **Admin Dashboard**
- **Processing Monitor** - мониторинг обработки файлов
- **Verification Interface** - визуальная проверка результатов
- **File Preview** - предпросмотр закодированных файлов
- **Action Controller** - утверждение/отмена отправки

#### 4. **Notification Service**
- **Email Sender** - SMTP уведомления
- **Telegram Bot** - мгновенные уведомления
- **Template Engine** - шаблоны сообщений

#### 5. **Download Service**
- **Link Generator** - создание временных ссылок
- **Access Controller** - проверка доступа и срока действия
- **File Streamer** - безопасная передача файлов

#### 6. **Storage System**
- **Source Files Storage** - исходные файлы
- **Processed Files Storage** - закодированные файлы
- **Backup Storage** - резервные копии

---

## Выбранный технологический стек - VPS Only

### VPS-based решение (рекомендуемое)

**Преимущества:** 
- Полный контроль над системой
- Фиксированная стоимость
- Простая установка одной командой
- Независимость от внешних сервисов
- Легкое масштабирование при необходимости

**Недостатки:** 
- Ответственность за backup и мониторинг
- Необходимость администрирования сервера

**Компоненты:**
```
Nginx (reverse proxy + SSL)
    ↓
Go Web Server + Worker Pool
    ↓
Redis (queue + cache) + PostgreSQL (metadata)
    ↓
Local File System (хранение файлов)
    ↓
SMTP Service + Telegram Bot API
```

**Оценка стоимости (месяц):**
- VPS (4 core, 8GB RAM, 200GB SSD): ~$25-40
- Domain + SSL (Let's Encrypt): ~$10-15
- Email service (опционально): ~$0-10
- **Итого: ~$35-65/месяц**

**Рекомендуемые характеристики VPS:**
- CPU: 4+ cores
- RAM: 8+ GB
- Storage: 200+ GB SSD
- Bandwidth: 1TB+
- OS: Ubuntu 22.04 LTS

---

## Детальный технологический стек

### **Go + PostgreSQL + Redis + Local Storage**

**Обоснование выбора Go:**
1. **Производительность** - компилируемый язык, быстрая обработка
2. **Простота деплоя** - один исполняемый файл
3. **Отличная поддержка конкурентности** - горутины для параллельной обработки
4. **Горутины** - эффективная обработка конкурентных задач
5. **Небольшое потребление памяти**
6. **Отличные библиотеки** для обработки изображений (OpenCV)
7. **Быстрая компиляция** и простота развертывания
8. **Кроссплатформенность** - легко собрать под любую ОС

**Альтернативы:**
- **Python** (если важнее скорость разработки)
- **Rust** (если важна максимальная производительность)

### Полная структура проекта:
```
photo-processing-server/
├── cmd/
│   └── server/
│       └── main.go        # Основное приложение
├── internal/
│   ├── handlers/       # HTTP обработчики
│   │   ├── webhook.go   # WooCommerce webhooks  
│   │   ├── admin.go     # Admin panel API
│   │   └── download.go  # Download service
│   ├── services/       # Бизнес логика
│   │   ├── processor.go # Обработка файлов
│   │   ├── encoding.go  # Водяные знаки
│   │   ├── imaging.go   # Обработка изображений (OpenCV)
│   │   ├── watermark.go # Бинарные водяные знаки
│   │   ├── notify.go    # Уведомления
│   │   └── links.go     # Временные ссылки
│   ├── models/         # Модели базы данных
│   ├── config/         # Конфигурация
│   └── utils/          # Общие утилиты
├── web/                # Веб-интерфейс
│   ├── static/         # Статические файлы
│   └── templates/      # HTML шаблоны
├── deployments/        # Автоматическая установка
│   ├── docker-compose.yml
│   ├── install.sh      # Скрипт установки
│   ├── nginx.conf      # Конфиг Nginx
│   └── .env.example    # Пример конфигурации
├── scripts/            # Скрипты сборки и развертывания
│   ├── build.sh
│   └── release.sh
├── migrations/         # SQL миграции
├── go.mod
├── go.sum
└── README.md
```

---

## План портирования логики с Kotlin

### 1. File Processing (BatchUtils.kt → processor.go)
```go
type ProcessingJob struct {
    OrderID     string
    SourcePath  string 
    NumCopies   int
    BaseText    string
    AddSwap     bool
    AddWatermark bool
    CreateZip   bool
}

func ProcessBatch(job ProcessingJob) error {
    // 1. Copy source files to processing directory
    // 2. Apply text encoding (port from EncodingUtils.kt)
    // 3. Add binary watermarks (port from WatermarkUtils.kt)
    // 4. Add visible watermarks (port from ImageUtils.kt)
    // 5. Perform swap operation if enabled
    // 6. Create ZIP archives
    // 7. Move to final storage location
    // 8. Update database with processing results
}
```

### 2. Encoding (EncodingUtils.kt → encoding.go)
```go
const SHIFT = 7

func EncodeText(text string) string {
    // Port Caesar cipher logic
}

func DecodeText(text string) string {
    // Port decoding logic  
}

func AddWatermark(text string) string {
    return fmt.Sprintf("<<==%s==>>", EncodeText(text))
}
```

### 3. Image Processing (ImageUtils.kt → imaging.go)
```go
import "gocv.io/x/gocv" // OpenCV для Go

func AddVisibleWatermark(imagePath, text string) error {
    // Port OpenCV logic for visible watermarks
}
```

### 4. Binary Watermarks (WatermarkUtils.kt → watermark.go)
```go
func AddBinaryWatermark(filepath, encodedText string) error {
    // Port binary watermark logic
}

func RemoveBinaryWatermark(filepath string) error {
    // Port removal logic
}
```

---

## Интеграция с WooCommerce

### 1. Настройка Webhook в WooCommerce
```php
// В functions.php темы WordPress
add_action('woocommerce_order_status_completed', 'send_to_processing_service');

function send_to_processing_service($order_id) {
    $order = wc_get_order($order_id);
    
    $data = array(
        'order_id' => $order_id,
        'customer_email' => $order->get_billing_email(),
        'items' => array(),
        'custom_fields' => array()
    );
    
    foreach ($order->get_items() as $item) {
        // Извлекаем информацию о фотосессии
        $data['items'][] = array(
            'product_id' => $item->get_product_id(),
            'photoshoot_id' => get_post_meta($item->get_product_id(), 'photoshoot_id', true),
            'quantity' => $item->get_quantity()
        );
    }
    
    // Отправляем на наш сервис
    wp_remote_post('https://your-service.com/webhook/order', array(
        'headers' => array('Content-Type' => 'application/json'),
        'body' => json_encode($data)
    ));
}
```

### 2. Обработка Webhook в Go сервисе
```go
type WooCommerceOrder struct {
    OrderID       string `json:"order_id"`
    CustomerEmail string `json:"customer_email"`
    Items         []OrderItem `json:"items"`
}

type OrderItem struct {
    ProductID    string `json:"product_id"`
    PhotoshootID string `json:"photoshoot_id"`
    Quantity     int    `json:"quantity"`
}

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
    var order WooCommerceOrder
    json.NewDecoder(r.Body).Decode(&order)
    
    // Добавляем в очередь обработки
    job := ProcessingJob{
        OrderID:      order.OrderID,
        SourcePath:   fmt.Sprintf("/photos/%s", order.Items[0].PhotoshootID),
        // ... остальные параметры
    }
    
    redis.Client.LPush("processing_queue", job.ToJSON())
    
    w.WriteHeader(200)
}
```

---

## Admin Panel - Визуальная проверка

### Дизайн интерфейса (React + Go API):

```
┌─────────────────────────────────────────┐
│ Order #12345 - Ready for Review         │
│ Customer: john@example.com              │
│ Status: [Processing Complete] ●         │
└─────────────────────────────────────────┘

┌─── Processing Logs ───┐  ┌─── Swap Info ────┐
│ ✓ 15 files encoded    │  │ Photo-003.jpg ↔  │  
│ ✓ 3 images watermark  │  │ Photo-013.jpg    │
│ ✓ 2 videos processed  │  │                  │
│ ✓ ZIP created         │  │ Photo-007.jpg ↔  │
│ ✓ Uploaded to storage │  │ Photo-017.jpg    │
└───────────────────────┘  └──────────────────┘

┌─── Watermark Preview ──┐  ┌─── File Stats ───┐
│ [Image with visible     │  │ Total: 20 files  │
│  watermark highlighted]│  │ Images: 18       │  
│ Text: "CONFIDENTIAL"   │  │ Videos: 2        │
│ Position: Bottom-Right │  │ Size: 245.7 MB   │
└────────────────────────┘  └──────────────────┘

     [Cancel] [Send to Customer]
```

### Go API для админ панели:
```go
type ProcessingStatus struct {
    OrderID        string                `json:"order_id"`
    Status         string                `json:"status"`
    ProcessingLogs []string             `json:"processing_logs"`
    SwapInfo       []SwapOperation      `json:"swap_info"`  
    WatermarkPreview string             `json:"watermark_preview"`
    FileStats      ProcessingStats      `json:"file_stats"`
}

func GetProcessingStatus(w http.ResponseWriter, r *http.Request) {
    orderID := r.URL.Query().Get("order_id")
    status := GetOrderProcessingStatus(orderID)
    json.NewEncoder(w).Encode(status)
}

func ApproveOrder(w http.ResponseWriter, r *http.Request) {
    orderID := r.URL.Query().Get("order_id")
    
    // Создаем временную ссылку
    downloadLink := CreateTemporaryLink(orderID, 7*24*time.Hour) // 1 неделя
    
    // Отправляем email клиенту
    SendDownloadEmail(orderID, downloadLink)
    
    w.WriteHeader(200)
}
```

---

## Система временных ссылок

### 1. Структура базы данных (PostgreSQL):
```sql
CREATE TABLE download_links (
    id SERIAL PRIMARY KEY,
    token VARCHAR(64) UNIQUE NOT NULL,
    order_id VARCHAR(50) NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    downloaded_at TIMESTAMP,
    download_count INTEGER DEFAULT 0,
    max_downloads INTEGER DEFAULT 5
);

CREATE INDEX idx_download_token ON download_links(token);
CREATE INDEX idx_expires_at ON download_links(expires_at);
```

### 2. Генерация ссылок:
```go
import "crypto/rand"

func CreateTemporaryLink(orderID string, duration time.Duration) string {
    token := generateSecureToken(32)
    
    link := DownloadLink{
        Token:         token,
        OrderID:       orderID,
        CustomerEmail: getCustomerEmail(orderID),
        FilePath:      fmt.Sprintf("/processed/%s.zip", orderID),
        ExpiresAt:     time.Now().Add(duration),
        MaxDownloads:  5,
    }
    
    db.Create(&link)
    
    return fmt.Sprintf("https://your-domain.com/download/%s", token)
}

func generateSecureToken(length int) string {
    bytes := make([]byte, length)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}
```

### 3. Обработка скачивания:
```go
func HandleDownload(w http.ResponseWriter, r *http.Request) {
    token := mux.Vars(r)["token"]
    
    var link DownloadLink
    result := db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&link)
    
    if result.Error != nil {
        http.Error(w, "Link expired or invalid", 404)
        return
    }
    
    if link.DownloadCount >= link.MaxDownloads {
        http.Error(w, "Download limit exceeded", 403)  
        return
    }
    
    // Увеличиваем счетчик
    db.Model(&link).Updates(DownloadLink{
        DownloadCount: link.DownloadCount + 1,
        DownloadedAt:  time.Now(),
    })
    
    // Стримим файл
    http.ServeFile(w, r, link.FilePath)
}
```

---

## Система уведомлений

### 1. Email уведомления (SMTP):
```go
import "net/smtp"

type EmailService struct {
    SMTPHost     string
    SMTPPort     string  
    SMTPUser     string
    SMTPPassword string
}

func (e *EmailService) SendAdminNotification(orderID string) error {
    to := []string{"admin@photocompany.com"}
    subject := fmt.Sprintf("New order #%s ready for review", orderID)
    
    body := fmt.Sprintf(`
        Order #%s has been processed and is ready for review.
        
        Admin Panel: https://admin.photocompany.com/orders/%s
        
        Please check the processing results and approve for delivery.
    `, orderID, orderID)
    
    return e.sendEmail(to, subject, body)
}

func (e *EmailService) SendCustomerDownloadLink(orderID, email, downloadLink string) error {
    to := []string{email}
    subject := "Your photos are ready for download"
    
    body := fmt.Sprintf(`
        Dear Customer,
        
        Your order #%s is ready for download.
        
        Download Link: %s
        
        This link will expire in 7 days and allows 5 downloads.
        
        Best regards,
        Photo Company Team
    `, orderID, downloadLink)
    
    return e.sendEmail(to, subject, body)
}
```

### 2. Telegram уведомления:
```go
import "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type TelegramService struct {
    Bot    *tgbotapi.BotAPI
    ChatID int64
}

func (t *TelegramService) SendAdminAlert(orderID string) error {
    message := fmt.Sprintf(`
🚨 *New Order Processing Complete*

Order ID: %s
Status: Ready for Review
Admin Panel: [Check Order](https://admin.photocompany.com/orders/%s)

Please review and approve for delivery.
    `, orderID, orderID)
    
    msg := tgbotapi.NewMessage(t.ChatID, message)
    msg.ParseMode = "Markdown"
    
    _, err := t.Bot.Send(msg)
    return err
}
```

---

## Автоматическая установка с GitHub

### Однокомандная установка

```bash
curl -fsSL https://raw.githubusercontent.com/your-username/photo-processing-server/main/deployments/install.sh | bash
```

### Скрипт install.sh будет выполнять:

```bash
#!/bin/bash
set -e

echo "🚀 Installing Photo Processing Server..."

# 1. Проверяем системные требования
if ! command -v docker &> /dev/null; then
    echo "📦 Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    sudo usermod -aG docker $USER
fi

if ! command -v docker-compose &> /dev/null; then
    echo "📦 Installing Docker Compose..."
    sudo curl -L "https://github.com/docker/compose/releases/download/v2.0.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
fi

# 2. Скачиваем последний release
GITHUB_REPO="your-username/photo-processing-server"
LATEST_RELEASE=$(curl -s https://api.github.com/repos/$GITHUB_REPO/releases/latest | grep '"tag_name"' | cut -d '"' -f 4)

echo "📥 Downloading version $LATEST_RELEASE..."
wget https://github.com/$GITHUB_REPO/archive/$LATEST_RELEASE.tar.gz
tar -xzf $LATEST_RELEASE.tar.gz
cd photo-processing-server-*

# 3. Копируем конфигурацию
cp deployments/.env.example .env

# 4. Генерируем случайные пароли
DB_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 64)

# 5. Обновляем .env файл
sed -i "s/DB_PASSWORD=.*/DB_PASSWORD=$DB_PASSWORD/" .env
sed -i "s/JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" .env

# 6. Запрашиваем основные настройки
echo "⚙️ Configuring server..."
read -p "Enter your domain (e.g. photos.yourdomain.com): " DOMAIN
read -p "Enter admin email: " ADMIN_EMAIL
read -p "Enter SMTP server (e.g. smtp.gmail.com): " SMTP_HOST
read -s -p "Enter SMTP password: " SMTP_PASSWORD
echo
read -p "Enter Telegram Bot Token (optional): " TELEGRAM_TOKEN

sed -i "s/DOMAIN=.*/DOMAIN=$DOMAIN/" .env
sed -i "s/ADMIN_EMAIL=.*/ADMIN_EMAIL=$ADMIN_EMAIL/" .env
sed -i "s/SMTP_HOST=.*/SMTP_HOST=$SMTP_HOST/" .env
sed -i "s/SMTP_PASSWORD=.*/SMTP_PASSWORD=$SMTP_PASSWORD/" .env
sed -i "s/TELEGRAM_TOKEN=.*/TELEGRAM_TOKEN=$TELEGRAM_TOKEN/" .env

# 7. Создаём директории
sudo mkdir -p /opt/photo-processing
sudo mkdir -p /var/lib/photo-processing/{photos,processed,temp}
sudo chown -R $USER:$USER /var/lib/photo-processing

# 8. Копируем файлы
sudo cp -r . /opt/photo-processing/
sudo chown -R $USER:$USER /opt/photo-processing

# 9. Устанавливаем SSL через Let's Encrypt
if [ ! -z "$DOMAIN" ]; then
    echo "🔒 Setting up SSL certificate..."
    sudo apt-get update && sudo apt-get install -y certbot
    sudo certbot certonly --standalone -d $DOMAIN --email $ADMIN_EMAIL --agree-tos --non-interactive
fi

# 10. Запускаем сервисы
echo "🚀 Starting services..."
cd /opt/photo-processing
docker-compose -f deployments/docker-compose.yml up -d

# 11. Ожидаем запуска
echo "⏳ Waiting for services to start..."
sleep 30

# 12. Запускаем миграции
echo "📊 Running database migrations..."
docker-compose -f deployments/docker-compose.yml exec -T app ./photo-processor migrate

echo "✅ Installation completed!"
echo ""
echo "Admin Panel: https://$DOMAIN/admin"
echo "Webhook URL: https://$DOMAIN/webhook/order"
echo "Default admin credentials will be sent to $ADMIN_EMAIL"
echo ""
echo "Next steps:"
echo "1. Configure your WooCommerce webhook to point to: https://$DOMAIN/webhook/order"
echo "2. Add your photo directories to /var/lib/photo-processing/photos/"
echo "3. Check logs: docker-compose -f /opt/photo-processing/deployments/docker-compose.yml logs"
```

### Docker Compose конфигурация:

```yaml
# deployments/docker-compose.yml
version: '3.8'

services:
  app:
    build: ../
    container_name: photo-processor
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    env_file:
      - ../.env
    volumes:
      - /var/lib/photo-processing:/app/data
      - ../web:/app/web
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15-alpine
    container_name: photo-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: photoprocessing
      POSTGRES_USER: photoprocessing
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    container_name: photo-redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  nginx:
    image: nginx:alpine
    container_name: photo-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - /etc/letsencrypt:/etc/letsencrypt:ro
    depends_on:
      - app

volumes:
  postgres_data:
  redis_data:
```

### Пример .env конфигурации:

```bash
# deployments/.env.example
# Сервер
DOMAIN=photos.yourdomain.com
PORT=8080
ENVIRONMENT=production

# База данных
DB_HOST=postgres
DB_PORT=5432
DB_NAME=photoprocessing
DB_USER=photoprocessing
DB_PASSWORD=GENERATED_PASSWORD

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# Безопасность
JWT_SECRET=GENERATED_JWT_SECRET
API_SECRET=GENERATED_API_SECRET

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
ADMIN_EMAIL=admin@yourdomain.com

# Telegram (опционально)
TELEGRAM_TOKEN=your-bot-token
TELEGRAM_CHAT_ID=your-chat-id

# Обработка файлов
PHOTOS_PATH=/app/data/photos
PROCESSED_PATH=/app/data/processed
TEMP_PATH=/app/data/temp
MAX_FILE_SIZE=100MB
WORKER_COUNT=4

# Водяные знаки
WATERMARK_ENABLED=true
SWAP_ENABLED=true
VISIBLE_WATERMARK_ENABLED=true
```

### GitHub Actions для автоматических release:

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v3
        with:
          go-version: 1.21
      
      - name: Build binaries
        run: |
          GOOS=linux GOARCH=amd64 go build -o photo-processor-linux-amd64 ./cmd/server
          GOOS=linux GOARCH=arm64 go build -o photo-processor-linux-arm64 ./cmd/server
      
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            photo-processor-linux-amd64
            photo-processor-linux-arm64
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Простое обновление:

```bash
# Обновление до новой версии
cd /opt/photo-processing
./scripts/update.sh
```

---

## Веб-версия текущего Kotlin UI

### Анализ существующего интерфейса

На основе того, что мы видим в ConsoleState.kt, текущий Kotlin UI имеет:
- **Логи обработки** - список сообщений в реальном времени
- **Прогресс бар** - отображение прогресса обработки
- **Параметры обработки** - настройки для BatchUtils

### План портирования на веб (будущее развитие)

#### 1. Основной интерфейс (копия Kotlin UI)
```
┌──────────────────────────────────────────────────┐
│ Photo Processing Tool - Web Version                  │
├──────────────────────────────────────────────────┤
│ Source Folder: [Browse...] /photos/session-123      │
│ Number of Copies: [5    ]                           │
│ Base Text: [Project Alpha 001        ]              │
│                                                      │
│ ☑ Add Swap Operation                                │
│ ☑ Add Watermark                                    │
│ ☑ Create ZIP Archives                              │
│ ☑ Add Visible Watermark                            │
│                                                      │
│ Watermark Text: [CONFIDENTIAL    ]                  │
│ Photo Number: [3  ]                                 │
│                                                      │
│ [Start Processing]  [Clear Logs]                    │
├──────────────────────────────────────────────────┤
│ Progress: ████████░░░░░░░░░░░░ 65%              │
├──────────────────────────────────────────────────┤
│ Console Logs:                                        │
│ ✓ Directory copied: 001                            │
│ ✓ test.txt: Success ✔                             │
│ ✓ photo-001.jpg: Success ✔                        │
│ ✓ Added watermark to video: video.mp4               │
│ ✓ Added text to photo-003.jpg                       │
│ ✓ Swapping files: photo-001.jpg <--> photo-011.jpg  │
│ ✓ Created ZIP archive: /path/to/001/Project.zip      │
│ ▷ Processing folder: 002...                         │
│                                                      │
│ [Auto-scroll] [Export Logs] [Settings]              │
└──────────────────────────────────────────────────┘
```

#### 2. Технологии для веб-версии:
- **Frontend**: React + TypeScript + Tailwind CSS
- **Реальное время**: WebSocket для логов и прогресса
- **File Upload**: Drag & Drop интерфейс
- **Мобильная версия**: Адаптивный дизайн

#### 3. Дополнительные возможности веб-версии:
- **Предпросмотр файлов** - миниатюры изображений
- **Batch операции** - обработка нескольких папок
- **История обработки** - список предыдущих операций
- **Настройки по умолчанию** - сохранение пресетов
- **Многопользовательскость** - разные профили обработки

#### 4. План разработки (в будущем):
1. **Фаза 1** - Портирование базового UI (2 недели)
2. **Фаза 2** - WebSocket интеграция (1 неделя)
3. **Фаза 3** - File upload и drag&drop (1 неделя)
4. **Фаза 4** - Мобильная адаптация (1 неделя)
5. **Фаза 5** - Дополнительные фичи (2-3 недели)

---

## План поэтапной реализации

### **Фаза 0: Подготовка автоустановки (1 неделя)**
1. **Создание скриптов развертывания**
   - install.sh скрипт
   - Docker Compose конфигурация
   - Nginx конфигурация с SSL
   - .env.example шаблон

2. **GitHub Actions настройка**
   - Автоматическая сборка бинарников
   - Создание GitHub releases
   - Обновление install.sh скрипта

3. **Тестирование автоустановки**
   - Проверка на чистом Ubuntu VPS
   - Валидация всех компонентов

### **Фаза 1: Основная инфраструктура (2-3 недели)**
1. **Базовая структура Go приложения**
   - Основное приложение с HTTP сервером
   - Конфигурация и логирование
   - Подключение PostgreSQL и Redis

2. **Портирование логики обработки файлов**
   - encoding.go (Caesar cipher)
   - watermark.go (бинарные водяные знаки)
   - imaging.go (OpenCV + видимые водяные знаки)
   - processor.go (основная логика BatchUtils)

3. **Базовое API**
   - Webhook endpoint для WooCommerce
   - Очередь обработки (Redis)
   - File storage (локальная файловая система)

### **Фаза 2: Интеграция с WooCommerce (1-2 недели)**
1. **Webhook настройка**
   - PHP код для WooCommerce
   - Обработка заказов в Go сервисе
   - Тестирование интеграции

2. **Система метаданных**
   - Связь продуктов с папками фотографий
   - Конфигурация параметров кодирования
   - База данных заказов и статусов

### **Фаза 3: Admin Panel (2-3 недели)**
1. **Backend API**
   - Эндпоинты для получения статуса обработки
   - API для предпросмотра файлов
   - Система утверждения/отмены

2. **Frontend (React)**
   - Интерфейс мониторинга
   - Визуальная проверка результатов
   - Система уведомлений в реальном времени

### **Фаза 4: Download System (1-2 недели)**  
1. **Временные ссылки**
   - Генерация токенов
   - Валидация и ограничения
   - Автоматическая очистка истекших ссылок

2. **Безопасная передача файлов**
   - Стриминг больших файлов
   - Rate limiting
   - Логирование скачиваний

### **Фаза 5: Уведомления (1 неделя)**
1. **Email система**
   - SMTP настройка
   - Шаблоны писем
   - Уведомления админу и клиентам

2. **Telegram бот**
   - Создание бота
   - Интеграция с основным сервисом
   - Мгновенные уведомления

### **Фаза 6: Тестирование и оптимизация (1-2 недели)**
1. **Нагрузочное тестирование**
   - Обработка больших файлов
   - Конкурентная обработка заказов
   - Оптимизация производительности

2. **Мониторинг и логирование**
   - Система метрик
   - Алерты на ошибки
   - Backup стратегия

### **Общее время реализации: 9-14 недель**

Примечание: Добавлена "Фаза 0" для подготовки автоматической установки

---

## Оценка рисков и альтернативы

### **Потенциальные проблемы:**
1. **OpenCV на сервере** - может потребовать дополнительных библиотек
2. **Размер файлов** - большие видео могут замедлить обработку
3. **Конкурентность** - множество заказов одновременно

### **Решения:**
1. **OpenCV** - использовать Docker контейнер с предустановленными библиотеками
2. **Большие файлы** - асинхронная обработка + прогресс бар
3. **Масштабирование** - горизонтальное масштабирование worker'ов

### **План масштабирования (если понадобится):**
- **Вертикальное:** увеличение VPS ресурсов
- **Горизонтальное:** дополнительные worker VPS
- **Облачные сервисы:** для email (если SMTP не справляется)
- **CDN:** для быстрой доставки больших файлов
- **Локальный backup:** внешние HDD/SSD для резервных копий

---

## Следующие шаги

1. **Подготовка автоустановки** - создание Docker, скриптов и GitHub Actions
2. **Настройка тестового VPS** - проверка автоустановки
3. **Портирование логики обработки** - миграция с Kotlin на Go
4. **Интеграция с WooCommerce** - webhook и автоматизация
5. **Тестирование на реальных данных** - полный цикл от заказа до доставки
6. **Доработка Admin Panel** - визуальная проверка результатов
7. **Полное развертывание** - перевод на production

**Приоритетный подход:** Начать с Подготовки автоустановки и MVP обработки файлов, затем постепенно добавлять автоматизацию.

---

**Хотите обсудить какие-то конкретные аспекты или начать реализацию?**