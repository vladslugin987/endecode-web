import React, { useState } from 'react';
import { useApp } from '../../contexts/AppContext';
import { AddTextSettings } from '../../types';

const AddTextDialog: React.FC = () => {
  const { toggleDialog, addTextToPhoto } = useApp();
  
  // State matching Kotlin AddTextDialog
  const [textToAdd, setTextToAdd] = useState('');
  const [photoNumber, setPhotoNumber] = useState('');
  const [showError, setShowError] = useState(false);

  const handleConfirm = async () => {
    const number = parseInt(photoNumber);
    if (!textToAdd.trim() || !number || number < 1) {
      setShowError(true);
      return;
    }

    const settings: AddTextSettings = {
      text: textToAdd.trim(),
      photoNumber: number
    };

    await addTextToPhoto(settings);
    toggleDialog('addText');
  };

  const handleCancel = () => {
    toggleDialog('addText');
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-md m-4 border border-gray-200 dark:border-gray-800">
        {/* Dialog Header */}
        <div className="p-4 border-b border-gray-200 dark:border-gray-800">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Add Text to Photo</h2>
        </div>

        <div className="p-4 space-y-4">
          {/* Text to Add Field */}
          <div>
            <label className="block text-sm text-gray-700 dark:text-gray-300 mb-1">Text to add</label>
            <input
              type="text"
              value={textToAdd}
              onChange={(e) => {
                setTextToAdd(e.target.value);
                setShowError(false);
              }}
              className={`w-full h-10 px-3 border rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent ${showError && !textToAdd.trim() ? 'border-red-500' : 'border-gray-300 dark:border-gray-700'} bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100`}
              placeholder="Enter text to add"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Text that will be added to the photo</p>
          </div>

          {/* Photo Number Field */}
          <div>
            <label className="block text-sm text-gray-700 dark:text-gray-300 mb-1">Photo number</label>
            <input
              type="text"
              value={photoNumber}
              onChange={(e) => {
                const value = e.target.value.replace(/\D/g, ''); // Only digits
                setPhotoNumber(value);
                setShowError(false);
              }}
              className={`w-full h-10 px-3 border rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent ${showError && !parseInt(photoNumber) ? 'border-red-500' : 'border-gray-300 dark:border-gray-700'} bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100`}
              placeholder="12"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
              Enter the photo number (e.g., for photo_12.jpg enter 12)
            </p>
          </div>

          {/* Error Message */}
          {showError && (
            <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded p-3">
              <p className="text-sm text-red-700 dark:text-red-300">Please enter valid data</p>
            </div>
          )}
        </div>

        {/* Dialog Actions */}
        <div className="p-4 border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800 flex justify-end gap-2">
          <button
            onClick={handleCancel}
            className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded hover:bg-gray-50 dark:hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            Cancel
          </button>
          <button
            onClick={handleConfirm}
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            Add Text
          </button>
        </div>
      </div>
    </div>
  );
};

export default AddTextDialog;