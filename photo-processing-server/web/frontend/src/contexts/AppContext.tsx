import React, { createContext, useContext, useReducer, useCallback, useEffect, useRef, useState } from 'react';
import { AppState, AppAction, ProcessingState, BatchCopySettings, AddTextSettings } from '../types';
import { useWebSocket } from '../hooks/useWebSocket';
import * as api from '../services/api';

// Initial state matching Kotlin HomeViewModel defaults
const initialState: AppState = {
  processing: {
    selectedPath: null,
    nameToInject: '',
    autoClearConsole: false,
    isProcessing: false,
    progress: 0,
  },
  console: {
    logs: [],
    autoClear: false,
  },
  dialogs: {
    batchCopy: false,
    addText: false,
    deleteWatermarks: false,
  },
  files: {
    uploaded: [],
  },
};

// Reducer function (replaces HomeViewModel state management)
function appReducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case 'SET_SELECTED_PATH':
      return {
        ...state,
        processing: {
          ...state.processing,
          selectedPath: action.payload,
        },
      };
      
    case 'SET_NAME_TO_INJECT':
      return {
        ...state,
        processing: {
          ...state.processing,
          nameToInject: action.payload,
        },
      };
      
    case 'SET_AUTO_CLEAR_CONSOLE':
      return {
        ...state,
        processing: {
          ...state.processing,
          autoClearConsole: action.payload,
        },
        console: {
          ...state.console,
          autoClear: action.payload,
        },
      };
      
    case 'SET_PROCESSING':
      return {
        ...state,
        processing: {
          ...state.processing,
          isProcessing: action.payload,
          progress: action.payload ? state.processing.progress : 0,
        },
      };
      
    case 'SET_PROGRESS':
      return {
        ...state,
        processing: {
          ...state.processing,
          progress: action.payload,
        },
      };
      
    case 'ADD_LOG':
      return {
        ...state,
        console: {
          ...state.console,
          logs: [...state.console.logs, action.payload],
        },
      };
      
    case 'CLEAR_LOGS':
      return {
        ...state,
        console: {
          ...state.console,
          logs: [],
        },
      };
      
    case 'TOGGLE_DIALOG':
      return {
        ...state,
        dialogs: {
          ...state.dialogs,
          [action.payload]: !state.dialogs[action.payload],
        },
      };
      
    case 'SET_UPLOADED_FILES':
      return {
        ...state,
        files: {
          ...state.files,
          uploaded: action.payload,
        },
      };
      
    case 'RESET_STATE':
      return initialState;
      
    default:
      return state;
  }
}

// Context interface
interface AppContextType {
  state: AppState;
  dispatch: React.Dispatch<AppAction>;
  
  // Actions matching Kotlin HomeViewModel methods
  uploadAndSelect: (files: FileList, folderName?: string) => Promise<void>;
  updateSelectedPath: (path: string | null) => void;
  updateNameToInject: (name: string) => void;
  updateAutoClearConsole: (value: boolean) => void;
  addLog: (message: string) => void;
  clearLogs: () => void;
  showInfo: () => void;
  
  // Processing operations (matching HomeViewModel)
  encrypt: () => Promise<void>;
  decrypt: () => Promise<void>;
  performBatchCopy: (settings: BatchCopySettings) => Promise<void>;
  addTextToPhoto: (settings: AddTextSettings) => Promise<void>;
  removeWatermarks: () => Promise<void>;
  
  // Dialog management
  toggleDialog: (dialog: keyof AppState['dialogs']) => void;
  
  // Download
  lastDownloadTokenRef: React.MutableRefObject<string | null>;
}

const AppContext = createContext<AppContextType | undefined>(undefined);

// Custom hook to use the context
export const useApp = () => {
  const context = useContext(AppContext);
  if (context === undefined) {
    throw new Error('useApp must be used within an AppProvider');
  }
  return context;
};

// Provider component
export const AppProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [state, dispatch] = useReducer(appReducer, initialState);
  const lastJobIdRef = useRef<string | null>(null);
  const lastDownloadTokenRef = useRef<string | null>(null);
  const clickingRef = useRef<boolean>(false);
  
  // WebSocket for real-time updates (matches Kotlin ConsoleState real-time logging)
  const { sendMessage } = useWebSocket({
    onMessage: (message) => {
      switch (message.type) {
        case 'log':
          if (message.data.message) {
            dispatch({ type: 'ADD_LOG', payload: message.data.message });
          }
          break;
        case 'progress':
          if (typeof message.data.progress === 'number') {
            dispatch({ type: 'SET_PROGRESS', payload: message.data.progress });
          }
          break;
        case 'complete':
          dispatch({ type: 'SET_PROCESSING', payload: false });
          dispatch({ type: 'ADD_LOG', payload: 'Operation completed successfully' });
          break;
        case 'error':
          dispatch({ type: 'SET_PROCESSING', payload: false });
          if (message.data.error) {
            dispatch({ type: 'ADD_LOG', payload: `Error: ${message.data.error}` });
          }
          break;
      }
    },
    onOpen: () => {
      if (state.processing.isProcessing && lastJobIdRef.current) {
        sendMessage({ type: 'subscribe', jobId: lastJobIdRef.current });
        dispatch({ type: 'ADD_LOG', payload: `Re-subscribed to job: ${lastJobIdRef.current}` });
      }
    },
    onClose: () => {}
  });

  const trackJob = useCallback((jobId: string) => {
    lastJobIdRef.current = jobId;
    sendMessage({ type: 'subscribe', jobId });
    dispatch({ type: 'ADD_LOG', payload: `Subscribed to job: ${jobId}` });
  }, [sendMessage]);

  const validatePath = useCallback((): boolean => {
    if (!state.processing.selectedPath) {
      dispatch({ type: 'ADD_LOG', payload: 'Error: No folder selected' });
      return false;
    }
    return true;
  }, [state.processing.selectedPath]);

  const validateInput = useCallback((): boolean => {
    if (!validatePath()) return false;
    if (state.processing.nameToInject.trim() === '') {
      dispatch({ type: 'ADD_LOG', payload: 'Error: Name to inject is empty' });
      return false;
    }
    return true;
  }, [validatePath, state.processing.nameToInject]);

  // Debounce helper for action buttons
  const guardClick = useCallback(async (fn: () => Promise<void>) => {
    if (clickingRef.current) return;
    clickingRef.current = true;
    try {
      await fn();
    } finally {
      setTimeout(() => { clickingRef.current = false; }, 300);
    }
  }, []);

  const launchOperation = useCallback(async (operation: () => Promise<void>) => {
    try {
      dispatch({ type: 'SET_PROCESSING', payload: true });
      dispatch({ type: 'SET_PROGRESS', payload: 0 });
      if (state.processing.autoClearConsole) {
        dispatch({ type: 'CLEAR_LOGS' });
      }
      await operation();
    } catch (error) {
      console.error('Operation failed:', error);
      const message = error instanceof Error ? error.message : 'Unknown error occurred';
      dispatch({ type: 'ADD_LOG', payload: `Error: ${message}` });
      dispatch({ type: 'SET_PROCESSING', payload: false });
      dispatch({ type: 'SET_PROGRESS', payload: 0 });
    }
  }, [state.processing.autoClearConsole]);

  const updateSelectedPath = useCallback((path: string | null) => {
    dispatch({ type: 'SET_SELECTED_PATH', payload: path });
    if (state.processing.autoClearConsole) {
      dispatch({ type: 'CLEAR_LOGS' });
    }
    if (path) {
      dispatch({ type: 'ADD_LOG', payload: '***** Folder successfully selected! *****' });
      dispatch({ type: 'ADD_LOG', payload: '' });
    }
  }, [state.processing.autoClearConsole]);

  const uploadAndSelect = useCallback(async (files: FileList, folderName?: string) => {
    try {
      if (!files || files.length === 0) {
        dispatch({ type: 'ADD_LOG', payload: 'Error: No files selected' });
        return;
      }
      dispatch({ type: 'ADD_LOG', payload: `Uploading ${files.length} file(s)...` });
      const response = await api.uploadFiles(files, folderName);
      if (response.success && response.jobId) {
        updateSelectedPath(response.jobId);
        // Use original folder name if available, otherwise fall back to jobId
        const displayName = (response as any).originalName || response.jobId;
        dispatch({ type: 'ADD_LOG', payload: `Files uploaded: ${displayName}` });
        dispatch({ type: 'ADD_LOG', payload: `Upload path: ${response.jobId}` });
        // Capture token if present for later download
        // @ts-ignore
        if ((response as any).token) {
          // @ts-ignore
          lastDownloadTokenRef.current = (response as any).token;
        }
      } else {
        dispatch({ type: 'ADD_LOG', payload: `Upload failed: ${response.error || 'Unknown error'}` });
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unknown upload error';
      dispatch({ type: 'ADD_LOG', payload: `Upload error: ${message}` });
    }
  }, [updateSelectedPath]);

  const updateNameToInject = useCallback((name: string) => {
    dispatch({ type: 'SET_NAME_TO_INJECT', payload: name });
  }, []);

  const updateAutoClearConsole = useCallback((value: boolean) => {
    dispatch({ type: 'SET_AUTO_CLEAR_CONSOLE', payload: value });
  }, []);

  const addLog = useCallback((message: string) => {
    dispatch({ type: 'ADD_LOG', payload: message });
  }, []);

  const clearLogs = useCallback(() => {
    dispatch({ type: 'CLEAR_LOGS' });
  }, []);

  const showInfo = useCallback(() => {
    const width = 70;
    const line = '='.repeat(width);
    const logs = [
      line,
      `███████╗███╗   ██╗██████╗ ███████╗ ██████╗ ██████╗ ██████╗ ███████╗`,
      `██╔════╝████╗  ██║██╔══██╗██╔════╝██╔════╝██╔═══██╗██╔══██╗██╔════╝`,
      `█████╗  ██╔██╗ ██║██║  ██║█████╗  ██║     ██║   ██║██║  ██║█████╗  `,
      `██╔══╝  ██║╚██╗██║██║  ██║██╔══╝  ██║     ██║   ██║██║  ██║██╔══╝  `,
      `███████╗██║ ╚████║██████╔╝███████╗╚██████╗╚██████╔╝██████╔╝███████╗`,
      `╚══════╝╚═╝  ╚═══╝╚═════╝ ╚══════╝ ╚═════╝ ╚═════╝ ╚═════╝ ╚══════╝`,
      line,
    ];
    logs.forEach(log => dispatch({ type: 'ADD_LOG', payload: log }));
  }, []);

  const encrypt = useCallback(async () => {
    if (state.processing.isProcessing) return;
    if (!validateInput()) return;
    await guardClick(async () => {
      await launchOperation(async () => {
        const response = await api.encrypt({
          selectedPath: state.processing.selectedPath!,
          nameToInject: state.processing.nameToInject,
        });
        if (response.jobId) {
          trackJob(response.jobId);
        }
      });
    });
  }, [state.processing.isProcessing, validateInput, launchOperation, state.processing.selectedPath, state.processing.nameToInject, trackJob, guardClick]);

  const decrypt = useCallback(async () => {
    if (state.processing.isProcessing) return;
    if (!validatePath()) return;
    await guardClick(async () => {
      await launchOperation(async () => {
        const response = await api.decrypt({ selectedPath: state.processing.selectedPath! });
        if (response.jobId) {
          trackJob(response.jobId);
        }
      });
    });
  }, [state.processing.isProcessing, validatePath, launchOperation, state.processing.selectedPath, trackJob, guardClick]);

  const performBatchCopy = useCallback(async (settings: BatchCopySettings) => {
    if (!validatePath()) return;
    await guardClick(async () => {
      await launchOperation(async () => {
        dispatch({ type: 'ADD_LOG', payload: 'Starting batch copy process...' });
        const response = await api.batchCopy({ selectedPath: state.processing.selectedPath!, settings });
        if (response.jobId) {
          sendMessage({ type: 'subscribe', jobId: response.jobId });
        }
      });
    });
  }, [validatePath, launchOperation, state.processing.selectedPath, sendMessage, guardClick]);

  const addTextToPhoto = useCallback(async (settings: AddTextSettings) => {
    if (!validatePath()) return;
    await guardClick(async () => {
      await launchOperation(async () => {
        dispatch({ type: 'ADD_LOG', payload: 'Adding text to photo...' });
        const response = await api.addText({ selectedPath: state.processing.selectedPath!, settings });
        if (response.jobId) {
          sendMessage({ type: 'subscribe', jobId: response.jobId });
        }
      });
    });
  }, [validatePath, launchOperation, state.processing.selectedPath, sendMessage, guardClick]);

  const removeWatermarks = useCallback(async () => {
    if (!validatePath()) return;
    await guardClick(async () => {
      await launchOperation(async () => {
        dispatch({ type: 'ADD_LOG', payload: 'Starting watermark removal process...' });
        const response = await api.removeWatermarks({ selectedPath: state.processing.selectedPath! });
        if (response.jobId) {
          sendMessage({ type: 'subscribe', jobId: response.jobId });
        }
      });
    });
  }, [validatePath, launchOperation, state.processing.selectedPath, sendMessage, guardClick]);

  const toggleDialog = useCallback((dialog: keyof AppState['dialogs']) => {
    dispatch({ type: 'TOGGLE_DIALOG', payload: dialog });
  }, []);

  const value = {
    state,
    dispatch,
    uploadAndSelect,
    updateSelectedPath,
    updateNameToInject,
    updateAutoClearConsole,
    addLog,
    clearLogs,
    showInfo,
    encrypt,
    decrypt,
    performBatchCopy,
    addTextToPhoto,
    removeWatermarks,
    toggleDialog,
    lastDownloadTokenRef,
  } as AppContextType;

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
};