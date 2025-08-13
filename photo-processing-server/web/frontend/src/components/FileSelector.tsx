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
    <div className="space-y-1">
      {/* Choose Folder Button */}
      <button
        disabled={isProcessing}
        onClick={() => document.getElementById('file-input')?.click()}
        className="w-full h-8 px-3 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400
                 text-white rounded text-sm font-medium transition-colors
                 disabled:cursor-not-allowed"
      >
        Choose folder with files
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
          w-full h-10 border-2 border-dashed rounded-lg flex items-center justify-center
          transition-all duration-200 text-sm font-medium
          ${isDragOver 
            ? 'border-blue-500 bg-blue-50 text-blue-700' 
            : 'border-gray-300 bg-gray-50 text-gray-500'
          }
          ${isProcessing ? 'cursor-not-allowed opacity-50' : 'cursor-pointer hover:border-blue-400 hover:bg-blue-25'}
        `}
      >
        {isDragOver ? 'Drop folder here' : 'Drag and drop folder here'}
      </div>

      {/* Selected Path Display */}
      {state.processing.selectedPath && (
        <div className="p-2 bg-green-50 border border-green-200 rounded text-sm">
          <div className="text-green-800 font-medium">Selected:</div>
          <div className="text-green-700 text-xs break-all mt-1">
            {state.processing.selectedPath}
          </div>
        </div>
      )}

      {/* Help Text */}
      <p className="text-xs text-gray-500">
        Supported files: .jpg, .jpeg, .png, .mp4, .avi, .mov, .mkv, .txt
      </p>
    </div>
  );
};

export default FileSelector;