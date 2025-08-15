import React, { useEffect, useState, useMemo } from 'react';
import { adminListJobs, adminJobDetails, JobSummary, adminApprove, adminImages, adminStats, adminLogs, AdminArchive } from '../services/api';

interface JobStats {
  images: number;
  videos: number;
  texts: number;
  zips: number;
  totalBytes: number;
}

interface LogEntry {
  message: string;
  timestamp: string;
  level: string;
}

const AdminPanel: React.FC = () => {
  const [jobs, setJobs] = useState<JobSummary[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadingDetails, setLoadingDetails] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [details, setDetails] = useState<any | null>(null);
  const [approveToken, setApproveToken] = useState<string | null>(null);
  const [archives, setArchives] = useState<AdminArchive[]>([]);
  const [selectedArchive, setSelectedArchive] = useState<AdminArchive | null>(null);
  const [stats, setStats] = useState<JobStats | null>(null);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [zoomUseSample, setZoomUseSample] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [currentView, setCurrentView] = useState<'overview' | 'details'>('overview');

  // Фильтрованные задачи
  const filteredJobs = useMemo(() => {
    return jobs.filter(job => {
      const matchesSearch = job.id.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesStatus = statusFilter === 'all' || job.status === statusFilter;
      return matchesSearch && matchesStatus;
    });
  }, [jobs, searchTerm, statusFilter]);

  // Статистика задач
  const jobsOverview = useMemo(() => {
    const total = jobs.length;
    const completed = jobs.filter(j => j.status === 'completed').length;
    const processing = jobs.filter(j => j.status === 'processing').length;
    const failed = jobs.filter(j => j.status === 'failed' || j.status === 'error').length;
    const pending = jobs.filter(j => j.status === 'pending').length;
    
    return { total, completed, processing, failed, pending };
  }, [jobs]);

  const load = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await adminListJobs();
      setJobs(res.jobs || []);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load jobs');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const loadJobData = async (id: string) => {
    try {
      setLoadingDetails(true);
      const [d, archivesRes, st, lg] = await Promise.all([
        adminJobDetails(id),
        adminImages(id),
        adminStats(id),
        adminLogs(id),
      ]);
      
      setDetails(d);
      setArchives(archivesRes.archives || []);
      setSelectedArchive(archivesRes.archives?.length > 0 ? archivesRes.archives[0] : null);
      setStats(st.stats);
      setLogs(lg.logs || []);
      setZoomUseSample(true);
      
      const r = d && (d.result ?? d.Result);
      if (r && r.approvedToken) {
        setApproveToken(r.approvedToken as string);
      } else {
        setApproveToken(null);
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load job details');
    } finally {
      setLoadingDetails(false);
    }
  };

  const selectJob = async (id: string) => {
    if (selectedId === id) return; // Already selected
    
    setSelectedId(id);
    setCurrentView('details');
    setDetails(null);
    setApproveToken(null);
    setArchives([]);
    setSelectedArchive(null);
    setStats(null);
    setLogs([]);
    setZoomUseSample(true);
    
    await loadJobData(id);
  };

  const approve = async () => {
    if (!selectedId) return;
    try {
      const r = await adminApprove(selectedId);
      setApproveToken(r.token);
      await loadJobData(selectedId);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Approve failed');
    }
  };

  const goBack = () => {
    setCurrentView('overview');
    setSelectedId(null);
    setDetails(null);
    setApproveToken(null);
    setArchives([]);
    setSelectedArchive(null);
    setStats(null);
    setLogs([]);
  };

  const downloadUrl = approveToken ? `${window.location.origin}/api/download/${approveToken}` : null;

  // Build watermark sample preview URL
  const sampleUrl = (() => {
    if (!selectedId || !details) return null;
    const r = (details.result ?? details.Result) as { watermarkSample?: any } | undefined;
    const s = r && (r.watermarkSample ?? (details.watermarkSample ?? undefined));
    if (!s) return null;
    if (s.path) return `/api/admin/jobs/${selectedId}/preview?path=${encodeURIComponent(s.path)}`;
    if (s.zip && s.entry) return `/api/admin/jobs/${selectedId}/preview?zip=${encodeURIComponent(s.zip)}&entry=${encodeURIComponent(s.entry)}`;
    return null;
  })();

  const firstPreview = archives.length > 0 && selectedArchive ? selectedArchive.images.length > 0 ? selectedArchive.images[0].previewURL : null : null;
  const zoomSrc = zoomUseSample && sampleUrl ? sampleUrl : firstPreview;

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
      case 'processing': return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200';
      case 'failed': case 'error': return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
      case 'pending': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200';
      default: return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed': return '✅';
      case 'processing': return '⏳';
      case 'failed': case 'error': return '❌';
      case 'pending': return '⏸️';
      default: return '❓';
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  // Overview Layout
  if (currentView === 'overview') {
    return (
      <div className="flex-1 p-6 bg-gray-50 dark:bg-gray-950">
        <div className="max-w-7xl mx-auto">
          {/* Header */}
          <div className="mb-8">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100 mb-2">
                  Admin Dashboard
                </h1>
                <p className="text-gray-600 dark:text-gray-400">
                  Manage and monitor photo processing jobs
                </p>
              </div>
              <button
                onClick={load}
                disabled={loading}
                className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors duration-200 disabled:opacity-50"
              >
                {loading ? (
                  <>
                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Refreshing...
                  </>
                ) : (
                  <>
                    Refresh
                  </>
                )}
              </button>
            </div>
          </div>

          {error && (
            <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
              <div className="flex items-center">
                <svg className="w-5 h-5 text-red-600 dark:text-red-400 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span className="text-red-700 dark:text-red-300">{error}</span>
              </div>
            </div>
          )}

          {/* Stats Cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6 mb-8">
            <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-6">
              <div className="flex items-center">
                <div className="p-2 bg-gray-100 dark:bg-gray-800 rounded-lg">
                  <svg className="w-6 h-6 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                </div>
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Total Jobs</p>
                  <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{jobsOverview.total}</p>
                </div>
              </div>
            </div>
            
            <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-6">
              <div className="flex items-center">
                <div className="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
                  <svg className="w-6 h-6 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Completed</p>
                  <p className="text-2xl font-bold text-green-600 dark:text-green-400">{jobsOverview.completed}</p>
                </div>
              </div>
            </div>
            
            <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-6">
              <div className="flex items-center">
                <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
                  <svg className="w-6 h-6 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Processing</p>
                  <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">{jobsOverview.processing}</p>
                </div>
              </div>
            </div>
            
            <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-6">
              <div className="flex items-center">
                <div className="p-2 bg-yellow-100 dark:bg-yellow-900 rounded-lg">
                  <svg className="w-6 h-6 text-yellow-600 dark:text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Pending</p>
                  <p className="text-2xl font-bold text-yellow-600 dark:text-yellow-400">{jobsOverview.pending}</p>
                </div>
              </div>
            </div>
            
            <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-6">
              <div className="flex items-center">
                <div className="p-2 bg-red-100 dark:bg-red-900 rounded-lg">
                  <svg className="w-6 h-6 text-red-600 dark:text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </div>
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Failed</p>
                  <p className="text-2xl font-bold text-red-600 dark:text-red-400">{jobsOverview.failed}</p>
                </div>
              </div>
            </div>
          </div>

          {/* Search and Filters */}
          <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-6 mb-6">
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="flex-1">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Search Jobs
                </label>
                <div className="relative">
                  <input
                    type="text"
                    placeholder="Search by job ID..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-500 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                  <span className="absolute left-3 top-2.5 text-gray-400">🔍</span>
                  <svg className="w-4 h-4 absolute left-3 top-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                  </svg>
                </div>
              </div>
              
              <div className="sm:w-48">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Filter by Status
                </label>
                <select
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="all">All Status</option>
                  <option value="completed">Completed</option>
                  <option value="processing">Processing</option>
                  <option value="pending">Pending</option>
                  <option value="failed">Failed</option>
                  <option value="error">Error</option>
                </select>
              </div>
            </div>
          </div>

          {/* Jobs Table */}
          <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                Processing Jobs {filteredJobs.length > 0 && `(${filteredJobs.length})`}
              </h2>
            </div>

            {loading ? (
              <div className="flex items-center justify-center py-12">
                <div className="text-center">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
                  <p className="text-gray-600 dark:text-gray-400">Loading jobs...</p>
                </div>
              </div>
            ) : filteredJobs.length === 0 ? (
              <div className="text-center py-12">
                <div className="w-16 h-16 mx-auto mb-4 text-gray-400">
                  <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                </div>
                <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
                  {jobs.length === 0 ? 'No jobs yet' : 'No jobs match your search'}
                </h3>
                <p className="text-gray-600 dark:text-gray-400 mb-6">
                  {jobs.length === 0 
                    ? 'Jobs will appear here when users start processing files'
                    : 'Try adjusting your search or filter criteria'
                  }
                </p>
                {searchTerm || statusFilter !== 'all' ? (
                  <button
                    onClick={() => {
                      setSearchTerm('');
                      setStatusFilter('all');
                    }}
                    className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors duration-200"
                  >
                    Clear Filters
                  </button>
                ) : null}
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-gray-50 dark:bg-gray-800">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Job
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Status
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Progress
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Started
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                    {filteredJobs.map((job) => (
                      <tr key={job.id} className="hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors duration-200">
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="flex items-center">
                            <div className="h-10 w-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center text-white text-sm font-medium">
                              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                              </svg>
                            </div>
                            <div className="ml-4">
                              <div className="text-sm font-medium text-gray-900 dark:text-gray-100 font-mono">
                                {job.id.slice(0, 8)}...
                              </div>
                              <div className="text-xs text-gray-500 dark:text-gray-400">
                                {job.id}
                              </div>
                            </div>
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getStatusColor(job.status)}`}>
                            {job.status}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="flex items-center">
                            <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2 mr-2">
                              <div
                                className={`h-2 rounded-full transition-all duration-300 ${
                                  job.status === 'completed' ? 'bg-green-500' :
                                  job.status === 'failed' || job.status === 'error' ? 'bg-red-500' :
                                  'bg-blue-500'
                                }`}
                                style={{ width: `${Math.round(job.progress * 100)}%` }}
                              ></div>
                            </div>
                            <span className="text-sm text-gray-600 dark:text-gray-400 min-w-0">
                              {Math.round(job.progress * 100)}%
                            </span>
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600 dark:text-gray-400">
                          {formatDate(job.startTime)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm">
                          <button
                            onClick={() => selectJob(job.id)}
                            className="inline-flex items-center px-3 py-1 bg-blue-600 text-white text-xs font-medium rounded-md hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors duration-200"
                          >
                            View Details
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }

  // Details Layout
  return (
    <div className="flex-1 p-6 bg-gray-50 dark:bg-gray-950">
      <div className="max-w-7xl mx-auto">
        {/* Header with Breadcrumb */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <nav className="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400 mb-2">
                <button 
                  onClick={goBack}
                  className="hover:text-blue-600 dark:hover:text-blue-400 transition-colors duration-200"
                >
                  Admin Dashboard
                </button>
                <span>/</span>
                <span className="font-medium text-gray-900 dark:text-gray-100">Job Details</span>
              </nav>
              <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">
                Job: {selectedId?.slice(0, 8)}...
              </h1>
              {details && (
                <div className="flex items-center gap-2 mt-2">
                  <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getStatusColor(details.status || 'unknown')}`}>
                    {details.status || 'Unknown'}
                  </span>
                  {details.progress !== undefined && (
                    <span className="text-sm text-gray-600 dark:text-gray-400">
                      {Math.round(details.progress * 100)}% complete
                    </span>
                  )}
                </div>
              )}
            </div>
            <div className="flex items-center gap-3">
              <button
                onClick={goBack}
                className="inline-flex items-center px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors duration-200"
              >
                <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                </svg>
                Back to Overview
              </button>
              <button
                onClick={approve}
                disabled={!selectedId || loadingDetails}
                className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
                {loadingDetails ? 'Loading...' : 'Approve'}
              </button>
              {downloadUrl && (
                <a
                  href={downloadUrl}
                  download
                  className="inline-flex items-center px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 focus:ring-2 focus:ring-green-500 focus:ring-offset-2 transition-colors duration-200"
                >
                  <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  Download
                </a>
              )}
            </div>
          </div>
        </div>

        {error && (
          <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
            <div className="flex items-center">
              <svg className="w-5 h-5 text-red-600 dark:text-red-400 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-red-700 dark:text-red-300">{error}</span>
            </div>
          </div>
        )}

        {loadingDetails ? (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
              <p className="text-lg text-gray-600 dark:text-gray-400">Loading job details...</p>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
            {/* Left: Job Info & Stats */}
            <div className="space-y-6">
              {/* Job Info */}
              <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-6">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Job Information</h3>
                <div className="space-y-3">
                  <div>
                    <span className="text-sm text-gray-600 dark:text-gray-400">Job ID:</span>
                    <p className="font-mono text-sm text-gray-900 dark:text-gray-100">{selectedId}</p>
                  </div>
                  {details?.startTime && (
                    <div>
                      <span className="text-sm text-gray-600 dark:text-gray-400">Started:</span>
                      <p className="text-sm text-gray-900 dark:text-gray-100">{formatDate(details.startTime)}</p>
                    </div>
                  )}
                  {details?.message && (
                    <div>
                      <span className="text-sm text-gray-600 dark:text-gray-400">Message:</span>
                      <p className="text-sm text-gray-900 dark:text-gray-100">{details.message}</p>
                    </div>
                  )}
                </div>
              </div>

              {/* Stats */}
              {stats && (
                <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 p-6">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Statistics</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="text-center p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
                      <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">{stats.images}</p>
                      <p className="text-xs text-blue-600 dark:text-blue-400">Images</p>
                    </div>
                    <div className="text-center p-3 bg-purple-50 dark:bg-purple-900/20 rounded-lg">
                      <p className="text-2xl font-bold text-purple-600 dark:text-purple-400">{stats.videos}</p>
                      <p className="text-xs text-purple-600 dark:text-purple-400">Videos</p>
                    </div>
                    <div className="text-center p-3 bg-green-50 dark:bg-green-900/20 rounded-lg">
                      <p className="text-2xl font-bold text-green-600 dark:text-green-400">{stats.texts}</p>
                      <p className="text-xs text-green-600 dark:text-green-400">Texts</p>
                    </div>
                    <div className="text-center p-3 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
                      <p className="text-2xl font-bold text-yellow-600 dark:text-yellow-400">{stats.zips}</p>
                      <p className="text-xs text-yellow-600 dark:text-yellow-400">ZIPs</p>
                    </div>
                  </div>
                  <div className="mt-4 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg text-center">
                    <p className="text-lg font-bold text-gray-900 dark:text-gray-100">{formatBytes(stats.totalBytes)}</p>
                    <p className="text-xs text-gray-600 dark:text-gray-400">Total Size</p>
                  </div>
                </div>
              )}
            </div>

            {/* Center: Preview Grid */}
            <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
              <div className="p-6 border-b border-gray-200 dark:border-gray-700">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Image Preview</h3>
                {archives.length > 1 && (
                  <div className="mt-4">
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Archive ({archives.length} available):
                    </label>
                    <select 
                      value={selectedArchive?.name || ''} 
                      onChange={(e) => {
                        const selected = archives.find(a => a.name === e.target.value);
                        setSelectedArchive(selected || null);
                      }}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    >
                      {archives.map((archive, idx) => (
                        <option key={idx} value={archive.name}>
                          {archive.name} ({archive.images.length} images) [{archive.type.toUpperCase()}]
                        </option>
                      ))}
                    </select>
                  </div>
                )}
              </div>
              
              <div className="p-6">
                {archives.length === 0 ? (
                  <div className="text-center py-8">
                    <div className="w-12 h-12 mx-auto mb-2 text-gray-400">
                      <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                      </svg>
                    </div>
                    <p className="text-gray-600 dark:text-gray-400">No archives found</p>
                  </div>
                ) : selectedArchive && selectedArchive.images.length === 0 ? (
                  <div className="text-center py-8">
                    <div className="w-12 h-12 mx-auto mb-2 text-gray-400">
                      <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                      </svg>
                    </div>
                    <p className="text-gray-600 dark:text-gray-400">No images in selected archive</p>
                  </div>
                ) : selectedArchive && selectedArchive.images.length > 0 ? (
                  <div>
                    <div className="text-xs text-gray-600 dark:text-gray-400 mb-4 p-2 bg-gray-50 dark:bg-gray-800 rounded">
                      <strong>{selectedArchive.name}</strong> ({selectedArchive.images.length} images, {selectedArchive.type.toUpperCase()})
                    </div>
                    <div className="grid grid-cols-2 lg:grid-cols-3 gap-3 max-h-96 overflow-y-auto">
                      {selectedArchive.images.map((img, idx) => (
                        <div key={idx} className="group">
                          <div className="aspect-square overflow-hidden rounded-lg ring-1 ring-gray-200 dark:ring-gray-700 group-hover:ring-blue-500 transition-all duration-200">
                            <img 
                              src={img.previewURL} 
                              alt={img.name} 
                              className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200" 
                            />
                          </div>
                          <p className="mt-1 text-xs text-gray-600 dark:text-gray-400 truncate" title={img.name}>
                            {img.name}
                          </p>
                        </div>
                      ))}
                    </div>
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <div className="w-12 h-12 mx-auto mb-2 text-gray-400">
                      <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                      </svg>
                    </div>
                    <p className="text-gray-600 dark:text-gray-400">Select a job to view images</p>
                  </div>
                )}
              </div>
            </div>

            {/* Right: Watermark Zoom & Logs */}
            <div className="space-y-6">
              {/* Watermark Zoom */}
              <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
                <div className="p-6 border-b border-gray-200 dark:border-gray-700">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Watermark Zoom</h3>
                </div>
                <div className="p-6">
                  {!zoomSrc ? (
                    <div className="text-center py-8">
                      <div className="w-12 h-12 mx-auto mb-2 text-gray-400">
                        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                        </svg>
                      </div>
                      <p className="text-gray-600 dark:text-gray-400">No sample available</p>
                    </div>
                  ) : (
                    <div className="aspect-square overflow-hidden rounded-lg ring-1 ring-gray-200 dark:ring-gray-700 relative bg-[linear-gradient(45deg,rgba(255,255,255,0.04)_25%,transparent_25%,transparent_75%,rgba(255,255,255,0.04)_75%,rgba(255,255,255,0.04)),linear-gradient(45deg,rgba(255,255,255,0.04)_25%,transparent_25%,transparent_75%,rgba(255,255,255,0.04)_75%,rgba(255,255,255,0.04))] bg-[length:20px_20px,20px_20px] bg-[position:0_0,10px_10px] dark:bg-[linear-gradient(45deg,rgba(0,0,0,0.2)_25%,transparent_25%,transparent_75%,rgba(0,0,0,0.2)_75%,rgba(0,0,0,0.2)),linear-gradient(45deg,rgba(0,0,0,0.2)_25%,transparent_25%,transparent_75%,rgba(0,0,0,0.2)_75%,rgba(0,0,0,0.2))]">
                      <img
                        src={zoomSrc}
                        alt="watermark sample"
                        onError={() => setZoomUseSample(false)}
                        className="absolute right-0 bottom-0 object-contain origin-bottom-right [image-rendering:pixelated]"
                        style={{ transform: 'scale(6)' }}
                      />
                    </div>
                  )}
                </div>
              </div>

              {/* Logs */}
              <div className="bg-white dark:bg-gray-900 rounded-xl shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
                <div className="p-6 border-b border-gray-200 dark:border-gray-700">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Logs</h3>
                </div>
                <div className="p-6">
                  {logs.length === 0 ? (
                    <div className="text-center py-4">
                      <div className="w-8 h-8 mx-auto mb-2 text-gray-400">
                        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                        </svg>
                      </div>
                      <p className="text-gray-600 dark:text-gray-400 text-sm">No logs available</p>
                    </div>
                  ) : (
                    <div className="max-h-64 overflow-y-auto space-y-2">
                      {logs.map((log, i) => (
                        <div key={i} className="flex items-start gap-2 p-2 rounded bg-gray-50 dark:bg-gray-800">
                          <span className="text-xs text-gray-500 dark:text-gray-400 font-mono min-w-0 flex-shrink-0">
                            {new Date(log.timestamp).toLocaleTimeString()}
                          </span>
                          <span className={`text-xs font-mono ${log.level === 'ERROR' ? 'text-red-600 dark:text-red-400' : 'text-gray-700 dark:text-gray-300'}`}>
                            {log.message}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default AdminPanel; 