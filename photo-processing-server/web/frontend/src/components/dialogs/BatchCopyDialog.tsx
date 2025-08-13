import React, { useState } from 'react';
import { useApp } from '../../contexts/AppContext';
import { BatchCopySettings } from '../../types';

const BatchCopyDialog: React.FC = () => {
  const { toggleDialog, performBatchCopy } = useApp();
  
  // State matching Kotlin BatchCopyDialog
  const [numberOfCopies, setNumberOfCopies] = useState('');
  const [baseText, setBaseText] = useState('ORDER');
  const [addSwapEncoding, setAddSwapEncoding] = useState(false);
  const [addVisibleWatermark, setAddVisibleWatermark] = useState(false);
  const [createZip, setCreateZip] = useState(false);
  const [watermarkText, setWatermarkText] = useState('');
  const [useOrderNumber, setUseOrderNumber] = useState(true);
  const [photoNumber, setPhotoNumber] = useState('');
  const [showError, setShowError] = useState(false);

  const handleConfirm = async () => {
    const copies = parseInt(numberOfCopies);
    if (!copies || copies < 1 || baseText.trim() === '') {
      setShowError(true);
      return;
    }

    const settings: BatchCopySettings = {
      numberOfCopies: copies,
      baseText: baseText.trim(),
      addSwapEncoding,
      addVisibleWatermark,
      createZip,
      watermarkText: addVisibleWatermark ? watermarkText : undefined,
      photoNumber: addVisibleWatermark && !useOrderNumber ? parseInt(photoNumber) || undefined : undefined,
      useOrderNumberAsPhotoNumber: addVisibleWatermark ? useOrderNumber : undefined
    };

    await performBatchCopy(settings);
    toggleDialog('batchCopy');
  };

  const handleCancel = () => {
    toggleDialog('batchCopy');
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto m-4 border border-gray-200 dark:border-gray-800">
        {/* Dialog Header */}
        <div className="p-4 border-b border-gray-200 dark:border-gray-800">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Batch Copy Settings</h2>
        </div>

        <div className="p-4 space-y-4">
          {/* Basic Settings Card */}
          <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 space-y-3">
            <h3 className="text-sm font-medium text-gray-900 dark:text-gray-100">Basic Settings</h3>
            
            <div>
              <label className="block text-sm text-gray-700 dark:text-gray-300 mb-1">Number of copies</label>
              <input
                type="text"
                value={numberOfCopies}
                onChange={(e) => {
                  const value = e.target.value.replace(/\D/g, ''); // Only digits
                  setNumberOfCopies(value);
                  setShowError(false);
                }}
                className="w-full h-8 px-2 border border-gray-300 dark:border-gray-700 rounded text-sm
                         bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100
                         focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Enter number of copies"
              />
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">How many copies to create</p>
            </div>

            <div>
              <label className="block text-sm text-gray-700 dark:text-gray-300 mb-1">Text base for encoding</label>
              <input
                type="text"
                value={baseText}
                onChange={(e) => {
                  setBaseText(e.target.value);
                  setShowError(false);
                }}
                className="w-full h-8 px-2 border border-gray-300 dark:border-gray-700 rounded text-sm
                         bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100
                         focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="ORDER"
              />
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                Example: ORDER 001 (number will auto-increment for each copy)
              </p>
            </div>
          </div>

          {/* Additional Options Card */}
          <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 space-y-3">
            <h3 className="text-sm font-medium text-gray-900 dark:text-gray-100">Additional Options</h3>
            
            <label className="flex items-center text-sm text-gray-700 dark:text-gray-300">
              <input
                type="checkbox"
                checked={addSwapEncoding}
                onChange={(e) => setAddSwapEncoding(e.target.checked)}
                className="mr-2 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 dark:border-gray-600 rounded"
              />
              Additional Swap File Encoding (e.g., swap 003 with 103)
            </label>

            <label className="flex items-center text-sm text-gray-700 dark:text-gray-300">
              <input
                type="checkbox"
                checked={addVisibleWatermark}
                onChange={(e) => setAddVisibleWatermark(e.target.checked)}
                className="mr-2 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 dark:border-gray-600 rounded"
              />
              Add visible watermark to photos
            </label>

            <label className="flex items-center text-sm text-gray-700 dark:text-gray-300">
              <input
                type="checkbox"
                checked={createZip}
                onChange={(e) => setCreateZip(e.target.checked)}
                className="mr-2 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 dark:border-gray-600 rounded"
              />
              Create ZIP archives (No compression, No password)
            </label>
          </div>

          {/* Watermark Settings Card - Show only if watermark is enabled */}
          {addVisibleWatermark && (
            <div className="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-4 space-y-3">
              <h3 className="text-sm font-medium text-gray-900 dark:text-gray-100">Watermark Settings</h3>
              
              <div>
                <label className="block text-sm text-gray-700 dark:text-gray-300 mb-1">Watermark text</label>
                <input
                  type="text"
                  value={watermarkText}
                  onChange={(e) => setWatermarkText(e.target.value)}
                  className="w-full h-8 px-2 border border-gray-300 dark:border-gray-700 rounded text-sm
                           bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100
                           focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder=""
                />
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  Leave empty to use folder number
                </p>
              </div>

              <label className="flex items-center text-sm text-gray-700 dark:text-gray-300">
                <input
                  type="checkbox"
                  checked={useOrderNumber}
                  onChange={(e) => setUseOrderNumber(e.target.checked)}
                  className="mr-2 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 dark:border-gray-600 rounded"
                />
                Use order number as photo number
              </label>

              {!useOrderNumber && (
                <div>
                  <label className="block text-sm text-gray-700 dark:text-gray-300 mb-1">Photo number</label>
                  <input
                    type="text"
                    value={photoNumber}
                    onChange={(e) => {
                      const value = e.target.value.replace(/\D/g, ''); // Only digits
                      setPhotoNumber(value);
                    }}
                    className="w-full h-8 px-2 border border-gray-300 dark:border-gray-700 rounded text-sm
                             bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100
                             focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    placeholder=""
                  />
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    Enter specific photo number for watermark
                  </p>
                </div>
              )}
            </div>
          )}

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
            className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 
                     rounded hover:bg-gray-50 dark:hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            Cancel
          </button>
          <button
            onClick={handleConfirm}
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent 
                     rounded hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            Start
          </button>
        </div>
      </div>
    </div>
  );
};

export default BatchCopyDialog;