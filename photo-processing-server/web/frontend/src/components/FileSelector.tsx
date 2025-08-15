import React, { useCallback, useState } from 'react';
import { useApp } from '../contexts/AppContext';

interface WebKitFile extends File {
  webkitRelativePath: string;
}

const FileSelector: React.FC = () => {
  const { state, updateSelectedPath, addLog, uploadAndSelect } = useApp();
  const [isDragOver, setIsDragOver] = useState(false);
  const isProcessing = state.processing.isProcessing;

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (!isProcessing) {
      setIsDragOver(true);
    }
  }, [isProcessing]);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);

    if (isProcessing) return;

    const items = e.dataTransfer.items;
    if (items.length > 0) {
      const item = items[0];
      if (item.kind === 'file') {
        const entry = item.webkitGetAsEntry();
        if (entry?.isDirectory) {
          // For web, we'll simulate folder selection by showing the folder name
          const folderName = entry.name;
          updateSelectedPath(`/uploads/${folderName}`);
          addLog(`Folder "${folderName}" selected for processing`);
        } else {
          addLog('Error: Please drag a folder, not individual files');
        }
      }
    }
  }, [isProcessing, updateSelectedPath, addLog]);

  const handleFileInputChange = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      // Extract folder name from webkitRelativePath of first file
      let folderName = '';
      const firstFile = files[0] as WebKitFile;
      if (firstFile.webkitRelativePath) {
        const parts = firstFile.webkitRelativePath.split('/');
        if (parts.length > 1) {
          folderName = parts[0];
        }
      }
      
      await uploadAndSelect(files, folderName);
    }
  }, [uploadAndSelect]);

  return (
    <div className="space-y-4">
      {/* Primary Action Button */}
      <button
        disabled={isProcessing}
        onClick={() => document.getElementById('file-input')?.click()}
        className="w-full flex items-center justify-center px-4 py-3 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white rounded-xl font-medium transition-all duration-200 disabled:cursor-not-allowed shadow-sm"
      >
        <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
        </svg>
        Choose Folder
      </button>

      {/* Hidden file input */}
      <input
        id="file-input"
        type="file"
        multiple
        {...({ webkitdirectory: "", directory: "" } as any)}
        accept=".jpg,.jpeg,.png,.mp4,.avi,.mov,.mkv,.txt"
        onChange={handleFileInputChange}
        className="hidden"
        disabled={isProcessing}
      />

      {/* Drag & Drop Zone */}
      <div
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        className={`
          relative w-full h-24 border-2 border-dashed rounded-xl flex flex-col items-center justify-center
          transition-all duration-200 cursor-pointer
          ${isDragOver 
            ? 'border-blue-500 bg-blue-50 dark:bg-blue-950/20' 
            : 'border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-800/50 hover:border-blue-400 hover:bg-blue-50/50 dark:hover:bg-blue-950/10'
          }
          ${isProcessing ? 'cursor-not-allowed opacity-50' : ''}
        `}
      >
        <svg className={`w-6 h-6 mb-1 ${isDragOver ? 'text-blue-600' : 'text-gray-400'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
        </svg>
        <span className={`text-sm font-medium ${isDragOver ? 'text-blue-700 dark:text-blue-300' : 'text-gray-600 dark:text-gray-400'}`}>
          {isDragOver ? 'Drop folder here' : 'Drag and drop folder here'}
        </span>
      </div>

      {/* Selected Path Display */}
      {state.processing.selectedPath && (
        <div className="p-4 bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-800 rounded-xl">
          <div className="flex items-start gap-3">
            <div className="flex-shrink-0 w-5 h-5 text-green-600 dark:text-green-400 mt-0.5">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-green-800 dark:text-green-200">
                Folder Selected
              </p>
              <p className="text-sm text-green-700 dark:text-green-300 break-all mt-1">
                {state.processing.selectedPath}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Supported Files Info */}
      <div className="p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
        <p className="text-xs text-gray-600 dark:text-gray-400 leading-relaxed">
          <span className="font-medium">Supported formats:</span> JPEG, PNG, MP4, AVI, MOV, MKV, TXT
        </p>
      </div>
    </div>
  );
};

export default FileSelector;