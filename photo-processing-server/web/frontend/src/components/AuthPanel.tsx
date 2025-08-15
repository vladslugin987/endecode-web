import React, { useEffect, useState } from 'react';
import { login, register, logout, me } from '../services/api';
import { User } from '../types';

interface AuthPanelProps {
  onUserChange?: (user: User | null) => void;
}

const AuthPanel: React.FC<AuthPanelProps> = ({ onUserChange }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [user, setUser] = useState<User | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    (async () => {
      try {
        const res = await me() as User;
        if (res?.id) {
          setUser(res);
          onUserChange?.(res);
        }
      } catch {}
    })();
  }, [onUserChange]);

  const validEmail = email.trim().length > 3 && email.includes('@');
  const validPassword = password.length >= 6;

  const refreshMe = async () => {
    try {
      const res = await me() as User;
      if (res?.id) {
        setUser(res);
        onUserChange?.(res);
      } else {
        setUser(null);
        onUserChange?.(null);
      }
    } catch { 
      setUser(null);
      onUserChange?.(null);
    }
  };

  const onLogin = async () => {
    if (loading) return;
    setError(null);
    setLoading(true);
    
    if (!validEmail) { setError('Enter a valid email'); setLoading(false); return; }
    if (!validPassword) { setError('Password must be at least 6 characters'); setLoading(false); return; }
    
    try {
      await login(email.trim(), password);
      // Небольшая задержка для обработки cookie
      setTimeout(async () => {
        await refreshMe();
        setEmail('');
        setPassword('');
        setLoading(false);
      }, 100);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Login failed');
      setLoading(false);
    }
  };

  const onRegister = async () => {
    if (loading) return;
    setError(null);
    setLoading(true);
    
    if (!validEmail) { setError('Enter a valid email'); setLoading(false); return; }
    if (!validPassword) { setError('Password must be at least 6 characters'); setLoading(false); return; }
    
    try {
      await register(email.trim(), password);
      // Небольшая задержка для обработки cookie
      setTimeout(async () => {
        await refreshMe();
        setEmail('');
        setPassword('');
        setLoading(false);
      }, 100);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Register failed');
      setLoading(false);
    }
  };

  const onLogout = async () => {
    try {
      await logout();
      setUser(null);
      onUserChange?.(null);
    } catch (e) {
      console.error('Logout error:', e);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && validEmail && validPassword && !loading) {
      onLogin();
    }
  };

  return (
    <div className="flex items-center gap-3">
      {user ? (
        <>
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600 dark:text-gray-300">
              Welcome, <span className="font-medium text-gray-900 dark:text-gray-100">{user.nickname}</span>
            </span>
            {user.role === 'admin' && (
              <span className="inline-flex items-center px-2 py-0.5 text-xs font-medium bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-200 rounded-full">
                Admin
              </span>
            )}
          </div>
          <button 
            type="button" 
            onClick={onLogout} 
            className="inline-flex items-center px-3 py-1.5 text-sm font-medium text-gray-700 dark:text-gray-200 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md transition-colors duration-200"
          >
            Logout
          </button>
        </>
      ) : (
        <div className="flex items-center gap-3">
          <div className="flex gap-2">
            <input 
              className="h-9 px-3 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors duration-200" 
              placeholder="email" 
              type="email"
              value={email} 
              onChange={e=>setEmail(e.target.value)}
              onKeyPress={handleKeyPress}
              disabled={loading}
            />
            <input 
              className="h-9 px-3 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors duration-200" 
              placeholder="password (>=6)" 
              type="password" 
              value={password} 
              onChange={e=>setPassword(e.target.value)}
              onKeyPress={handleKeyPress}
              disabled={loading}
            />
          </div>
          <div className="flex gap-2">
            <button 
              type="button" 
              onClick={onLogin} 
              disabled={!validEmail || !validPassword || loading} 
              className={`inline-flex items-center px-4 py-1.5 text-sm font-medium rounded-md transition-colors duration-200 ${
                (!validEmail || !validPassword || loading) 
                  ? 'bg-gray-200 text-gray-500 dark:bg-gray-700 dark:text-gray-400 cursor-not-allowed' 
                  : 'bg-blue-600 text-white hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'
              }`}
            >
              {loading ? (
                <>
                  <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-current" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Logging in...
                </>
              ) : (
                'Login'
              )}
            </button>
            <button 
              type="button" 
              onClick={onRegister} 
              disabled={!validEmail || !validPassword || loading} 
              className={`inline-flex items-center px-4 py-1.5 text-sm font-medium rounded-md border transition-colors duration-200 ${
                (!validEmail || !validPassword || loading) 
                  ? 'bg-gray-200 text-gray-500 dark:bg-gray-700 dark:text-gray-400 border-gray-300 dark:border-gray-600 cursor-not-allowed' 
                  : 'bg-white text-gray-700 dark:bg-gray-800 dark:text-gray-200 border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'
              }`}
            >
              {loading ? 'Processing...' : 'Register'}
            </button>
          </div>
          {error && (
            <div className="flex items-center">
              <span className="text-xs text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-2 py-1 rounded">
                {error}
              </span>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default AuthPanel; 