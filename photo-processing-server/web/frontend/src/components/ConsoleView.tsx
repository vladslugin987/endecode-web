import React from 'react';
import { useApp } from '../contexts/AppContext';
import { useTheme } from '../contexts/ThemeContext';

const ConsoleView: React.FC = () => {
  const { state, clearLogs, showInfo } = useApp();
  const { isDark, toggleDark } = useTheme();

  return (
    <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-3 flex flex-col h-full">
      <div className="flex items-center justify-between mb-2">
        <div className="text-sm font-semibold text-gray-700 dark:text-gray-200">Console</div>
        <div className="flex items-center gap-2">
          <button
            onClick={showInfo}
            className="h-8 px-3 bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 text-gray-800 dark:text-gray-100 rounded-lg text-xs font-semibold"
          >Info</button>
          <button
            onClick={clearLogs}
            className="h-8 px-3 bg-red-100 hover:bg-red-200 text-red-800 rounded-lg text-xs font-semibold"
          >Clear</button>
          <button
            onClick={toggleDark}
            className="h-8 px-3 bg-blue-100 hover:bg-blue-200 text-blue-800 rounded-lg text-xs font-semibold"
            title="Toggle dark mode"
          >{isDark ? 'Light' : 'Dark'}</button>
        </div>
      </div>
      <div className="flex-1 overflow-auto bg-gray-50 dark:bg-gray-950 border border-gray-200 dark:border-gray-800 rounded-lg p-2 text-xs text-gray-800 dark:text-gray-100">
        {state.console.logs.map((log, idx) => (
          <div key={idx} className="whitespace-pre-wrap break-words">{log}</div>
        ))}
      </div>
    </div>
  );
};

export default ConsoleView;