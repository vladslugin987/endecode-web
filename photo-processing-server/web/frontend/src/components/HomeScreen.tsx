import React from 'react';
import { useApp } from '../contexts/AppContext';
import FileSelector from './FileSelector';
import ConsoleView from './ConsoleView';
import ControlPanel from './ControlPanel';
import BatchCopyDialog from './dialogs/BatchCopyDialog';
import AddTextDialog from './dialogs/AddTextDialog';
import DeleteWatermarksDialog from './dialogs/DeleteWatermarksDialog';

const HomeScreen: React.FC = () => {
  const { state } = useApp();

  return (
    <div className="flex flex-col h-screen bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-950">
      {/* Main content area with 40/60 split */}
      <div className="flex-1 grid grid-cols-[40%_60%] gap-3 p-3 min-h-0">
        {/* Left Panel - Controls */}
        <div className="flex flex-col gap-3">
          <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-3">
            <FileSelector />
          </div>
          <ControlPanel />
        </div>
        {/* Right Panel - Console */}
        <div className="flex flex-col min-h-0">
          <ConsoleView />
        </div>
      </div>

      {/* Dialogs */}
      {state.dialogs.batchCopy && <BatchCopyDialog />}
      {state.dialogs.addText && <AddTextDialog />}
      {state.dialogs.deleteWatermarks && <DeleteWatermarksDialog />}
    </div>
  );
};

export default HomeScreen;