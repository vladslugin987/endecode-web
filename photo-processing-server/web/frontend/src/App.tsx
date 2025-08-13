import React, { useMemo, useState, useEffect } from 'react';
import { AppProvider } from './contexts/AppContext';
import HomeScreen from './components/HomeScreen';
import { ThemeProvider } from './contexts/ThemeContext';
import AdminPanel from './components/AdminPanel';

function App() {
  const initialView: 'home' | 'admin' = useMemo(() => {
    const path = typeof window !== 'undefined' ? window.location.pathname : '/';
    return path.startsWith('/admin') ? 'admin' : 'home';
  }, []);
  const [view, setView] = useState<'home' | 'admin'>(initialView);

  useEffect(() => {
    // Keep URL in sync without reload
    const targetPath = view === 'admin' ? '/admin' : '/';
    if (window.location.pathname !== targetPath) {
      window.history.replaceState({}, '', targetPath);
    }
  }, [view]);

  return (
    <ThemeProvider>
      <AppProvider>
        <div className="min-h-screen bg-background text-on-background dark:bg-gray-900 dark:text-gray-100">
          <div className="w-full border-b border-gray-200 dark:border-gray-800 bg-white/70 dark:bg-gray-900/70 backdrop-blur supports-[backdrop-filter]:bg-white/50 supports-[backdrop-filter]:dark:bg-gray-900/50">
            <div className="max-w-7xl mx-auto px-3 h-12 flex items-center justify-between">
              <div className="font-semibold">ENDECode Console</div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setView('home')}
                  className={`h-8 px-3 rounded ${view === 'home' ? 'bg-blue-600 text-white' : 'bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-200'}`}
                >
                  Home
                </button>
                <button
                  onClick={() => setView('admin')}
                  className={`h-8 px-3 rounded ${view === 'admin' ? 'bg-blue-600 text-white' : 'bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-200'}`}
                >
                  Admin
                </button>
              </div>
            </div>
          </div>
          {view === 'admin' ? <AdminPanel /> : <HomeScreen />}
        </div>
      </AppProvider>
    </ThemeProvider>
  );
}

export default App;