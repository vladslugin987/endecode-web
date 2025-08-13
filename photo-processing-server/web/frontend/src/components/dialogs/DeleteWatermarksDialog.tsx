import React from 'react';
import { useApp } from '../../contexts/AppContext';

const DeleteWatermarksDialog: React.FC = () => {
  const { toggleDialog, removeWatermarks } = useApp();

  const handleConfirm = async () => {
    await removeWatermarks();
    toggleDialog('deleteWatermarks');
  };

  const handleCancel = () => {
    toggleDialog('deleteWatermarks');
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-md m-4 border border-gray-200 dark:border-gray-800">
        {/* Dialog Header */}
        <div className="p-4 border-b border-gray-200 dark:border-gray-800">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Delete Watermarks</h2>
        </div>

        <div className="p-4">
          {/* Warning Text */}
          <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded p-4 mb-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-amber-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-amber-800 dark:text-amber-300">Warning</h3>
                <div className="mt-2 text-sm text-amber-700 dark:text-amber-200">
                  <p>This action will permanently remove all invisible watermarks from the selected files.</p>
                  <p className="mt-2 font-medium">This operation cannot be undone!</p>
                </div>
              </div>
            </div>
          </div>

          <p className="text-sm text-gray-600 dark:text-gray-300">
            Are you sure you want to delete all watermarks from the files in the selected folder?
          </p>
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
            className="px-4 py-2 text-sm font-medium text-white bg-red-600 border border-transparent 
                     rounded hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500"
          >
            Delete Watermarks
          </button>
        </div>
      </div>
    </div>
  );
};

export default DeleteWatermarksDialog;