# Анализ существующего UI и план веб-версии

## 📋 Анализ существующего Kotlin Compose UI

### 🏗️ Архитектура приложения

**Структура:**
```
Main.kt (Entry Point - 1024x768 window)
├── HomeScreen.kt (Main Layout)
│   ├── HomeViewModel.kt (Business Logic)
│   ├── Left Panel (40% width) - Controls
│   │   ├── FileSelector.kt (Drag & Drop)
│   │   ├── Control Buttons
│   │   └── Progress Indicator
│   └── Right Panel (60% width) - Console
│       └── ConsoleView.kt (Real-time logs)
└── Dialogs/
    ├── BatchCopyDialog.kt (Complex batch settings)
    ├── AddTextDialog.kt (Add text to photo)
    └── DeleteWatermarksDialog.kt (Confirmation)
```

### 🎛️ Детальная функциональность

#### **HomeScreen Layout (Row with 40/60 split):**

**Левая панель - Элементы управления:**
1. **FileSelector Component:**
   - Кнопка "Choose folder with files"
   - Drag & Drop зона (150dp height)
   - Visual feedback при перетаскивании
   - Отображение выбранного пути

2. **Main Action Buttons (Row):**
   - `DECRYPT` - расшифровка файлов
   - `ENCRYPT` - шифрование файлов
   - Disabled во время обработки

3. **Text Input:**
   - "Name to inject" - текст для кодирования
   - Single line, validation required

4. **Secondary Action Buttons (Row):**
   - `Batch Copy` - открывает BatchCopyDialog
   - `Add Text` - открывает AddTextDialog

5. **Additional Actions:**
   - `Delete Watermarks` - удаление водяных знаков
   - Checkbox: "Auto-clear console"

6. **Progress Indicator:**
   - LinearProgressIndicator (показывается только при isProcessing)
   - Animated progress updates

**Правая панель - Console:**
1. **Header (Row):**
   - "Console" title
   - "Info" button - показывает ASCII art + системную информацию
   - "Clear" button - очищает лог

2. **Log Area:**
   - LazyColumn с автоскроллом
   - Monospace font (FontFamily.Monospace)
   - SelectionContainer для копирования
   - Real-time updates через ConsoleState

#### **Диалоги:**

**1. BatchCopyDialog (600dp width, max 700dp height):**
- **Basic Settings Card:**
  - Number of copies (digits only)
  - Text base for encoding (auto-increment example)
  
- **Additional Options Card:**
  - ☐ Additional Swap File Encoding (swap 003 with 103)
  - ☐ Add visible watermark to photos
  - ☐ Create ZIP archives (No compression, No password)
  
- **Watermark Settings Card** (показывается если watermark включен):
  - Watermark text (опционально)
  - ☐ Use order number as photo number
  - Photo number (если не используется order number)

**2. AddTextDialog:**
- Text to add (required)
- Photo number (digits only, required)
- Validation с error states

**3. DeleteWatermarksDialog:**
- Simple confirmation
- Warning text о необратимости
- Red "Delete Watermarks" button

### 🎨 Дизайн система

**Цвета (Material 3 Light):**
- Primary: #1976D2 (синий)
- Secondary: #2196F3 (светло-синий)
- Surface: #FAFAFA
- Background: #F5F5F5

**Типография:**
- Title Large: 16sp
- Title Medium: 13sp
- Body Large: 13sp
- Body Medium: 11sp

**Размеры (Dimensions):**
- Button Height: 32dp
- Spacing Small: 4dp
- Spacing Medium: 8dp
- Spacing Large: 16dp

### 🔄 Состояния и логика (HomeViewModel)

**UI States:**
- `selectedPath: String?` - выбранная папка
- `nameToInject: String` - текст для кодирования
- `autoClearConsole: Boolean` - автоочистка консоли
- `isProcessing: Boolean` - флаг обработки
- `progress: Float` - прогресс (0.0-1.0)

**Main Operations:**
1. `encrypt()` - шифрует файлы с nameToInject
2. `decrypt()` - расшифровывает и показывает содержимое водяных знаков
3. `performBatchCopy()` - пакетное копирование с настройками
4. `addTextToPhoto()` - добавляет видимый текст к фото
5. `removeWatermarks()` - удаляет невидимые водяные знаки

**Features:**
- Асинхронные операции с Coroutines
- Progress callback для UI updates
- Error handling с логированием
- Auto-clear console опция

### 📊 Data Flow

```
User Action → HomeViewModel → Utils (Background Thread) → Progress Updates → ConsoleState.log() → UI Update
```

**Real-time Updates:**
- ConsoleState использует `mutableStateListOf<String>()`
- UI автоматически реагирует на изменения в `ConsoleState.logs`
- LazyColumn с `animateScrollToItem` для автоскролла

---

## 🌐 План веб-версии UI

### 🏗️ Архитектура веб-приложения

**Technology Stack:**
- **Frontend:** React 18 + TypeScript + Tailwind CSS
- **Backend:** Go HTTP Server + WebSocket
- **State Management:** React Context + useReducer
- **File Upload:** Drag & Drop + File API
- **Real-time:** WebSocket для логов и прогресса

**Structure:**
```
web/
├── frontend/                    # React приложение
│   ├── public/
│   ├── src/
│   │   ├── components/          # React компоненты
│   │   │   ├── FileSelector.tsx
│   │   │   ├── ConsoleView.tsx
│   │   │   ├── ControlPanel.tsx
│   │   │   └── dialogs/
│   │   │       ├── BatchCopyDialog.tsx
│   │   │       ├── AddTextDialog.tsx
│   │   │       └── DeleteWatermarksDialog.tsx
│   │   ├── hooks/               # Custom hooks
│   │   │   ├── useWebSocket.ts
│   │   │   ├── useFileUpload.ts
│   │   │   └── useProcessing.ts
│   │   ├── services/            # API services
│   │   │   ├── api.ts
│   │   │   └── websocket.ts
│   │   ├── types/               # TypeScript types
│   │   ├── utils/
│   │   └── styles/              # Tailwind config
│   ├── package.json
│   └── tailwind.config.js
└── backend/                     # Go HTTP handlers
    ├── handlers/
    │   ├── web.go               # Serve React app
    │   ├── api.go               # REST API endpoints
    │   ├── upload.go            # File upload handling
    │   └── websocket.go         # WebSocket connection
    └── static/                  # Built React app
```

### 🎯 Точное соответствие функциональности

#### **1. Layout Mapping:**

**Desktop → Web:**
```
Kotlin Compose Row(40/60)  →  CSS Grid (2-column: 40% 60%)
ElevatedCard               →  Tailwind card classes
MaterialTheme              →  CSS variables + Tailwind
Dimensions                 →  Tailwind spacing scale
```

#### **2. Component Mapping:**

**FileSelector.kt → FileSelector.tsx:**
```typescript
interface FileSelectorProps {
  selectedPath: string | null;
  onPathSelected: (path: string) => void;
  isProcessing: boolean;
}

// Features:
// - Drag & Drop API
// - Visual feedback (border animation)
// - File/Directory validation
// - Path display
```

**ConsoleView.kt → ConsoleView.tsx:**
```typescript
interface ConsoleViewProps {
  logs: string[];
  onClear: () => void;
  onShowInfo: () => void;
}

// Features:
// - Auto-scroll to bottom
// - Monospace font
// - Text selection support
// - Real-time updates via WebSocket
```

**HomeViewModel.kt → Custom Hooks:**
```typescript
// useProcessing.ts
interface ProcessingState {
  selectedPath: string | null;
  nameToInject: string;
  autoClearConsole: boolean;
  isProcessing: boolean;
  progress: number;
}

// Operations:
// - encrypt()
// - decrypt() 
// - performBatchCopy()
// - addTextToPhoto()
// - removeWatermarks()
```

#### **3. Dialog Mapping:**

**BatchCopyDialog.kt → BatchCopyDialog.tsx:**
```typescript
interface BatchCopySettings {
  numberOfCopies: number;
  baseText: string;
  addSwapEncoding: boolean;
  addVisibleWatermark: boolean;
  createZip: boolean;
  watermarkText?: string;
  photoNumber?: number;
}

// Exact form structure:
// - Basic Settings card
// - Additional Options checkboxes  
// - Conditional Watermark Settings
// - Validation & error states
```

### 🔗 Backend API Design

#### **REST Endpoints:**

```go
// File operations
POST   /api/upload              // Upload files/folder
GET    /api/files               // List uploaded files
DELETE /api/files/:id           // Delete uploaded files

// Processing operations  
POST   /api/encrypt             // Start encryption
POST   /api/decrypt             // Start decryption
POST   /api/batch-copy          // Start batch copy
POST   /api/add-text            // Add text to photo
POST   /api/remove-watermarks   // Remove watermarks

// Status and results
GET    /api/processing/:id      // Get processing status
GET    /api/download/:token     // Download processed files
```

#### **WebSocket Events:**

```typescript
// Client → Server
interface ClientMessage {
  type: 'subscribe' | 'unsubscribe';
  jobId?: string;
}

// Server → Client  
interface ServerMessage {
  type: 'log' | 'progress' | 'status' | 'complete' | 'error';
  data: {
    message?: string;      // For logs
    progress?: number;     // 0.0 - 1.0  
    status?: string;       // processing status
    error?: string;        // error message
    result?: any;          // final result
  };
}
```

### 🎨 Дизайн система (CSS Variables)

```css
/* colors.css - точная копия Kotlin theme */
:root {
  --color-primary: #1976D2;
  --color-secondary: #2196F3;  
  --color-surface: #FAFAFA;
  --color-background: #F5F5F5;
  --color-on-primary: #FFFFFF;
  --color-on-background: #1A1A1A;
  --color-on-surface: #1A1A1A;
  
  /* Typography - точные размеры */
  --text-title-large: 16px;
  --text-title-medium: 13px;
  --text-body-large: 13px;
  --text-body-medium: 11px;
  
  /* Dimensions - точные размеры */
  --button-height: 32px;
  --spacing-small: 4px;
  --spacing-medium: 8px; 
  --spacing-large: 16px;
}
```

### 📱 Responsive Considerations

**Desktop-first подход с breakpoints:**
```css
/* Desktop (1024x768 default) */
.main-layout {
  @apply grid grid-cols-[40%_60%] gap-2;
}

/* Tablet */
@media (max-width: 1024px) {
  .main-layout {
    @apply grid-cols-1 grid-rows-[auto_1fr];
  }
}

/* Mobile */  
@media (max-width: 768px) {
  .control-buttons {
    @apply flex-col space-y-2;
  }
}
```

### 🔄 State Management

**React Context Structure:**
```typescript
interface AppState {
  // Processing state
  processing: ProcessingState;
  
  // UI state  
  ui: {
    selectedPath: string | null;
    isProcessing: boolean;
    progress: number;
  };
  
  // Console state
  console: {
    logs: string[];
    autoClear: boolean;
  };
  
  // Dialog states
  dialogs: {
    batchCopy: boolean;
    addText: boolean;
    deleteWatermarks: boolean;
  };
}

// Actions - точно соответствуют ViewModel методам
type AppAction = 
  | { type: 'SET_SELECTED_PATH'; payload: string }
  | { type: 'SET_NAME_TO_INJECT'; payload: string }
  | { type: 'SET_PROCESSING'; payload: boolean }
  | { type: 'SET_PROGRESS'; payload: number }
  | { type: 'ADD_LOG'; payload: string }
  | { type: 'CLEAR_LOGS' }
  | { type: 'TOGGLE_DIALOG'; payload: keyof AppState['dialogs'] };
```

### 🚀 Development Plan

**Phase 1: Core Structure (1 week)**
1. Setup React + TypeScript + Tailwind
2. Create basic layout (40/60 grid)
3. Implement FileSelector component
4. Basic ConsoleView with mock data

**Phase 2: Go Backend Integration (1 week)**  
5. Create HTTP handlers in Go service
6. Implement file upload functionality
7. WebSocket connection for real-time updates
8. Connect frontend to backend APIs

**Phase 3: Processing Operations (1 week)**
9. Implement all processing operations
10. Add progress tracking and error handling
11. Create all dialog components
12. State management with Context

**Phase 4: Polish & Testing (1 week)**
13. Exact visual styling to match desktop
14. Mobile responsive design
15. Error handling and edge cases
16. Performance optimization

**Total: 4 weeks for complete web UI**

### 📊 Compatibility Matrix

| Feature | Desktop (Kotlin) | Web (React) | Status |
|---------|------------------|-------------|---------|
| File Selection | ✅ Drag & Drop | ✅ File API + DnD | Planned |
| Console Logging | ✅ Real-time | ✅ WebSocket | Planned |
| Progress Tracking | ✅ Coroutines | ✅ WebSocket | Planned |
| Batch Operations | ✅ Full | ✅ Full | Planned |
| Visual Design | ✅ Material 3 | ✅ Tailwind | Planned |
| Responsive | ❌ Fixed size | ✅ Mobile-ready | Enhanced |
| Cross-platform | ❌ Desktop only | ✅ Any browser | Enhanced |

---

## 🎯 Implementation Priority

**High Priority (MVP):**
1. ✅ File upload/selection 
2. ✅ Basic processing operations
3. ✅ Real-time console
4. ✅ Progress tracking

**Medium Priority:**
5. ✅ All dialogs functionality
6. ✅ Exact visual design
7. ✅ Error handling

**Low Priority (Nice to have):**
8. Mobile optimization
9. Keyboard shortcuts
10. Export/import settings
11. Processing history

**Веб-версия будет иметь 100% функциональную совместимость с desktop версией плюс дополнительные преимущества веб-платформы.**