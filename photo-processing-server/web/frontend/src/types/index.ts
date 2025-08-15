// Types matching Kotlin models and HomeViewModel states

// User and authentication types
export interface User {
  id: string;
  email: string;
  nickname: string;
  role: 'user' | 'admin';
  createdAt: string;
  lastLogin?: string;
}

export interface UserStats {
  totalUsers: number;
  adminUsers: number;
  activeUsers: number;
  registeredToday: number;
}

// Processing state (matches HomeViewModel)
export interface ProcessingState {
  selectedPath: string | null;
  nameToInject: string;
  autoClearConsole: boolean;
  isProcessing: boolean;
  progress: number; // 0.0 to 1.0
}

// Batch copy settings (matches BatchCopyDialog.kt parameters)
export interface BatchCopySettings {
  numberOfCopies: number;
  baseText: string;
  addSwapEncoding: boolean;
  addVisibleWatermark: boolean;
  createZip: boolean;
  watermarkText?: string;
  photoNumber?: number;
  useOrderNumberAsPhotoNumber?: boolean;
}

// Add text settings (matches AddTextDialog.kt)
export interface AddTextSettings {
  text: string;
  photoNumber: number;
}

// API request/response types
export interface EncryptRequest {
  selectedPath: string;
  nameToInject: string;
}

export interface DecryptRequest {
  selectedPath: string;
}

export interface BatchCopyRequest {
  selectedPath: string;
  settings: BatchCopySettings;
}

export interface AddTextRequest {
  selectedPath: string;
  settings: AddTextSettings;
}

export interface RemoveWatermarksRequest {
  selectedPath: string;
}

// WebSocket message types
export interface WebSocketMessage {
  type: 'log' | 'progress' | 'status' | 'complete' | 'error';
  data: {
    message?: string;      // For log messages
    progress?: number;     // Progress 0.0-1.0
    status?: string;       // Status update
    error?: string;        // Error message
    result?: any;          // Final result data
  };
}

export interface ClientMessage {
  type: 'subscribe' | 'unsubscribe';
  jobId?: string;
}

// Processing job status (matches Go models)
export interface ProcessingJob {
  id: string;
  orderID: string;
  sourcePath: string;
  numCopies: number;
  baseText: string;
  addSwap: boolean;
  addWatermark: boolean;
  createZip: boolean;
  watermarkText?: string;
  photoNumber?: number;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  createdAt: string;
  updatedAt: string;
}

// File upload types
export interface FileUpload {
  file: File;
  path: string;
  size: number;
}

export interface UploadedFile {
  id: string;
  name: string;
  path: string;
  size: number;
  type: 'file' | 'directory';
  uploadedAt: string;
}

// Console log entry (matches Kotlin ConsoleState)
export interface LogEntry {
  message: string;
  timestamp: string;
}

// Dialog state
export interface DialogState {
  batchCopy: boolean;
  addText: boolean;
  deleteWatermarks: boolean;
}

// App context state (matches HomeViewModel functionality)
export interface AppState {
  processing: ProcessingState;
  console: {
    logs: string[];
    autoClear: boolean;
  };
  dialogs: DialogState;
  files: {
    uploaded: UploadedFile[];
  };
}

// App actions (matches HomeViewModel methods)
export type AppAction =
  | { type: 'SET_SELECTED_PATH'; payload: string | null }
  | { type: 'SET_NAME_TO_INJECT'; payload: string }
  | { type: 'SET_AUTO_CLEAR_CONSOLE'; payload: boolean }
  | { type: 'SET_PROCESSING'; payload: boolean }
  | { type: 'SET_PROGRESS'; payload: number }
  | { type: 'ADD_LOG'; payload: string }
  | { type: 'CLEAR_LOGS' }
  | { type: 'TOGGLE_DIALOG'; payload: keyof DialogState }
  | { type: 'SET_UPLOADED_FILES'; payload: UploadedFile[] }
  | { type: 'RESET_STATE' };

// API response types
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  jobId?: string;
}

export interface ProcessingResponse {
  jobId: string;
  status: string;
  message: string;
}

// Error types
export interface ApiError {
  message: string;
  code?: number;
  details?: any;
}

// Component props interfaces
export interface FileSelectorProps {
  selectedPath: string | null;
  onPathSelected: (path: string) => void;
  isProcessing: boolean;
}

export interface ConsoleViewProps {
  logs: string[];
  onClear: () => void;
  onShowInfo: () => void;
}

export interface ControlPanelProps {
  processing: ProcessingState;
  onUpdateProcessing: (updates: Partial<ProcessingState>) => void;
  onEncrypt: () => void;
  onDecrypt: () => void;
  onOpenBatchDialog: () => void;
  onOpenAddTextDialog: () => void;
  onOpenDeleteWatermarksDialog: () => void;
}

export interface ProgressIndicatorProps {
  progress: number;
  isVisible: boolean;
}

// Text position enum (matches Kotlin ImageUtils.kt TextPosition)
export enum TextPosition {
  TopLeft = 'top_left',
  TopRight = 'top_right',
  Center = 'center',
  BottomLeft = 'bottom_left',
  BottomRight = 'bottom_right'
}

// File type detection (matches Kotlin FileUtils.kt)
export interface FileTypeInfo {
  isSupported: boolean;
  isImage: boolean;
  isVideo: boolean;
  isText: boolean;
  extension: string;
}

// Validation result
export interface ValidationResult {
  isValid: boolean;
  errors: string[];
}

// Subscription types
export interface Subscription {
  id: string;
  user_id: string;
  plan_type: string;
  status: 'active' | 'expired' | 'cancelled' | 'pending';
  start_date: string;
  end_date: string;
  created_at: string;
  updated_at: string;
  last_payment_id?: string;
  last_payment_date?: string;
  next_payment_date?: string;
  auto_renewal: boolean;
}

export interface SubscriptionPlan {
  id: string;
  name: string;
  description: string;
  price_usd: number;
  price_crypto?: string;
  duration_days: number;
  max_processing_jobs: number; // -1 for unlimited
  max_file_size: number; // in bytes
  features: string[];
  active: boolean;
}

export interface Payment {
  id: string;
  user_id: string;
  subscription_id: string;
  amount: number;
  currency: string;
  payment_method: 'crypto' | 'card';
  status: 'pending' | 'completed' | 'failed' | 'expired';
  crypto_address?: string;
  crypto_amount?: string;
  transaction_hash?: string;
  external_payment_id?: string;
  created_at: string;
  updated_at: string;
  expires_at?: string;
  paid_at?: string;
}

export interface UserUsage {
  id: string;
  user_id: string;
  month: string;
  processing_jobs: number;
  files_processed: number;
  storage_used_bytes: number;
  last_reset_date: string;
  created_at: string;
  updated_at: string;
}

export interface CryptoCurrency {
  code: string;
  name: string;
  symbol: string;
}