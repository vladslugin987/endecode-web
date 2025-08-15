import React, { useMemo, useState, useEffect } from 'react';
import { AppProvider } from './contexts/AppContext';
import HomeScreen from './components/HomeScreen';
import { ThemeProvider } from './contexts/ThemeContext';
import AdminPanel from './components/AdminPanel';
import AuthPanel from './components/AuthPanel';
import { me, adminUsers, adminUserStats, adminExtendSubscription } from './services/api';
import { User, UserStats } from './types';

// Temporary inline SuperAdminPanel component
const SuperAdminPanel: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [stats, setStats] = useState<UserStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedUser, setSelectedUser] = useState<string | null>(null);
  const [extendDays, setExtendDays] = useState<number>(30);
  const [extendPlan, setExtendPlan] = useState<string>('basic');
  const [extending, setExtending] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const [usersResponse, statsResponse] = await Promise.all([
        adminUsers(),
        adminUserStats(),
      ]);
      
      setUsers(usersResponse.users);
      setStats(statsResponse);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'Never';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getActivityStatus = (lastLogin?: string) => {
    if (!lastLogin) return 'inactive';
    const lastLoginDate = new Date(lastLogin);
    const thirtyDaysAgo = new Date();
    thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);
    return lastLoginDate > thirtyDaysAgo ? 'active' : 'inactive';
  };

  const handleExtendSubscription = async () => {
    if (!selectedUser || extending) return;

    try {
      setExtending(true);
      setError(null);
      
      const result = await adminExtendSubscription(selectedUser, extendPlan, extendDays);
      
      // Reload data to reflect changes
      await loadData();
      setSelectedUser(null);
      
      // Show success message briefly
      const successMessage = result.message;
      setError(null);
      setTimeout(() => {
        // Clear any success state if needed
      }, 3000);
      
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to extend subscription');
    } finally {
      setExtending(false);
    }
  };

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-red-600 mx-auto mb-4"></div>
          <p>Loading Super Admin Panel...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-600 mb-4">⚠️ Error</div>
          <p className="text-gray-600 dark:text-gray-400 mb-4">{error}</p>
          <button 
            onClick={loadData}
            className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 p-6 bg-gradient-to-br from-red-50 to-pink-100 dark:from-gray-900 dark:to-gray-950">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100 mb-2">
            Super Admin Panel
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            System administration and user management
          </p>
        </div>

        {/* Statistics Cards */}
        {stats && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-6">
              <div className="flex items-center">
                <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
                  <svg className="w-6 h-6 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
                  </svg>
                </div>
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Total Users</p>
                  <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.totalUsers}</p>
                </div>
              </div>
            </div>
            <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-6">
              <div className="flex items-center">
                <div className="p-2 bg-red-100 dark:bg-red-900 rounded-lg">
                  <svg className="w-6 h-6 text-red-600 dark:text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                  </svg>
                </div>
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Admins</p>
                  <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.adminUsers}</p>
                </div>
              </div>
            </div>
            <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-6">
              <div className="flex items-center">
                <div className="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
                  <svg className="w-6 h-6 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </div>
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Active (30d)</p>
                  <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.activeUsers}</p>
                </div>
              </div>
            </div>
            <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-6">
              <div className="flex items-center">
                <div className="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
                  <svg className="w-6 h-6 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                </div>
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">New Today</p>
                  <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.registeredToday}</p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Users Table */}
        <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">All Users</h2>
              <button 
                onClick={loadData}
                className="px-3 py-1 text-sm bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300 rounded"
              >
                Refresh
              </button>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 dark:bg-gray-800">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">User</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Role</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Registered</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Last Login</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {users.map((user) => (
                  <tr key={user.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <div className="h-8 w-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center text-white text-sm font-medium">
                          {user.nickname.charAt(0).toUpperCase()}
                        </div>
                        <div className="ml-3">
                          <div className="text-sm font-medium text-gray-900 dark:text-gray-100">{user.nickname}</div>
                          <div className="text-xs text-gray-500 dark:text-gray-400">{user.email}</div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                        user.role === 'admin' 
                          ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200' 
                          : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
                      }`}>
                        {user.role}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                        getActivityStatus(user.lastLogin) === 'active' 
                          ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' 
                          : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
                      }`}>
                        {getActivityStatus(user.lastLogin) === 'active' ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
                      {formatDate(user.createdAt)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
                      {formatDate(user.lastLogin)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      <button
                        onClick={() => setSelectedUser(user.id)}
                        className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300"
                      >
                        Extend Subscription
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {users.length === 0 && (
            <div className="text-center py-12">
              <div className="text-4xl mb-4">
                <svg className="w-16 h-16 mx-auto text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
                </svg>
              </div>
              <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-gray-100">No users found</h3>
              <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">No registered users in the system yet.</p>
            </div>
          )}
        </div>

        {/* Subscription Management Modal */}
        {selectedUser && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-xl max-w-md w-full mx-4 p-6">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
                Extend User Subscription
              </h3>
              
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Plan Type
                </label>
                <select
                  value={extendPlan}
                  onChange={(e) => setExtendPlan(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                >
                  <option value="free">Free Plan</option>
                  <option value="basic">Basic Plan ($29.99/month)</option>
                  <option value="pro">Professional Plan ($89.99/month)</option>
                  <option value="enterprise">Enterprise Plan ($299.99/month)</option>
                </select>
              </div>

              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Days to Add
                </label>
                <input
                  type="number"
                  min="1"
                  max="365"
                  value={extendDays}
                  onChange={(e) => setExtendDays(parseInt(e.target.value) || 30)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                />
              </div>

              <div className="flex gap-3">
                <button
                  onClick={() => setSelectedUser(null)}
                  className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800"
                >
                  Cancel
                </button>
                <button
                  onClick={handleExtendSubscription}
                  disabled={extending}
                  className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white rounded-lg"
                >
                  {extending ? 'Extending...' : 'Extend'}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

function App() {
  const initialView: 'home' | 'admin' | 'superadmin' = useMemo(() => {
    const path = typeof window !== 'undefined' ? window.location.pathname : '/';
    if (path.startsWith('/superadmin')) return 'superadmin';
    if (path.startsWith('/admin')) return 'admin';
    return 'home';
  }, []);
  
  const [view, setView] = useState<'home' | 'admin' | 'superadmin'>(initialView);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Keep URL in sync without reload
    const targetPath = view === 'admin' ? '/admin' : view === 'superadmin' ? '/superadmin' : '/';
    if (window.location.pathname !== targetPath) {
      window.history.replaceState({}, '', targetPath);
    }
  }, [view]);

  useEffect(() => {
    // Check authentication status
    (async () => {
      try {
        const res = await me() as User;
        if (res?.id) setUser(res);
      } catch {}
      setLoading(false);
    })();
  }, []);

  const handleUserChange = (newUser: User | null) => {
    setUser(newUser);
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-background text-on-background dark:bg-gray-900 dark:text-gray-100 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-lg text-gray-600 dark:text-gray-400">Loading ENDECode...</p>
        </div>
      </div>
    );
  }

  if (!user) {
    return (
      <ThemeProvider>
        <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-950 flex items-center justify-center">
          <div className="max-w-lg w-full mx-4">
            <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-xl ring-1 ring-gray-200 dark:ring-gray-800 p-8">
              <div className="text-center mb-8">
                <div className="mx-auto h-16 w-16 bg-gradient-to-br from-blue-600 to-purple-600 rounded-xl flex items-center justify-center mb-4">
                  <span className="text-2xl font-bold text-white">EC</span>
                </div>
                <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100 mb-2">ENDECode</h1>
                <p className="text-gray-600 dark:text-gray-400">Professional Photo Processing System</p>
              </div>
              
              <div className="space-y-6">
                <AuthPanel onUserChange={handleUserChange} />
              </div>
              
              <div className="mt-8 pt-6 border-t border-gray-200 dark:border-gray-700">
                <p className="text-xs text-gray-500 dark:text-gray-400 text-center">
                  Please log in to access the photo processing tools
                </p>
              </div>
            </div>
          </div>
        </div>
      </ThemeProvider>
    );
  }

  return (
    <ThemeProvider>
      <AppProvider>
        <div className="min-h-screen bg-background text-on-background dark:bg-gray-900 dark:text-gray-100">
          <div className="w-full border-b border-gray-200 dark:border-gray-800 bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm supports-[backdrop-filter]:bg-white/60 supports-[backdrop-filter]:dark:bg-gray-900/60">
            <div className="max-w-7xl mx-auto px-4 h-14 flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className="h-8 w-8 bg-gradient-to-br from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
                  <span className="text-sm font-bold text-white">EC</span>
                </div>
                <div className="font-semibold text-gray-900 dark:text-gray-100">ENDECode Console</div>
              </div>
              <div className="flex items-center gap-3">
                <div className="flex rounded-lg border border-gray-200 dark:border-gray-700 p-1 bg-gray-50 dark:bg-gray-800">
                  <button
                    onClick={() => setView('home')}
                    className={`px-3 py-1.5 text-sm font-medium rounded-md transition-colors duration-200 ${
                      view === 'home' 
                        ? 'bg-blue-600 text-white shadow-sm' 
                        : 'text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100'
                    }`}
                  >
                    Home
                  </button>
                  <button
                    onClick={() => setView('admin')}
                    className={`px-3 py-1.5 text-sm font-medium rounded-md transition-colors duration-200 ${
                      view === 'admin' 
                        ? 'bg-blue-600 text-white shadow-sm' 
                        : 'text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100'
                    }`}
                  >
                    Admin
                  </button>
                  {user.role === 'admin' && (
                    <button
                      onClick={() => setView('superadmin')}
                      className={`px-3 py-1.5 text-sm font-medium rounded-md transition-colors duration-200 ${
                        view === 'superadmin' 
                          ? 'bg-red-600 text-white shadow-sm' 
                          : 'text-red-700 dark:text-red-300 hover:text-red-900 dark:hover:text-red-100'
                      }`}
                    >
                      Super Admin
                    </button>
                  )}
                </div>
                <AuthPanel onUserChange={handleUserChange} />
              </div>
            </div>
          </div>
          {view === 'superadmin' ? <SuperAdminPanel /> : view === 'admin' ? <AdminPanel /> : <HomeScreen />}
        </div>
      </AppProvider>
    </ThemeProvider>
  );
}

export default App;