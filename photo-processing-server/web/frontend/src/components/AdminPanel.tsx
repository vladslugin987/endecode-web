import React, { useEffect, useState } from 'react';
import { adminListJobs, adminJobDetails, JobSummary, adminApprove, adminImages, adminStats, adminLogs, AdminArchive } from '../services/api';

const AdminPanel: React.FC = () => {
  const [jobs, setJobs] = useState<JobSummary[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [details, setDetails] = useState<any | null>(null);
  const [approveToken, setApproveToken] = useState<string | null>(null);
  const [archives, setArchives] = useState<AdminArchive[]>([]);
  const [selectedArchive, setSelectedArchive] = useState<AdminArchive | null>(null);
  const [stats, setStats] = useState<{ images: number; videos: number; texts: number; zips: number; totalBytes: number } | null>(null);
  const [logs, setLogs] = useState<{ message: string; timestamp: string; level: string }[]>([]);
  const [zoomUseSample, setZoomUseSample] = useState(true);

  const load = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await adminListJobs();
      setJobs(res.jobs);
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
    const [d, archivesRes, st, lg] = await Promise.all([
      adminJobDetails(id),
      adminImages(id),
      adminStats(id),
      adminLogs(id),
    ]);
    setDetails(d);
    setArchives(archivesRes.archives);
    setSelectedArchive(archivesRes.archives.length > 0 ? archivesRes.archives[0] : null);
    setStats(st.stats);
    setLogs(lg.logs);
    setZoomUseSample(true);
    const r = d && (d.result ?? d.Result);
    if (r && r.approvedToken) {
      setApproveToken(r.approvedToken as string);
    } else {
      setApproveToken(null);
    }
  };

  const selectJob = async (id: string) => {
    setSelectedId(id);
    setDetails(null);
    setApproveToken(null);
    setArchives([]);
    setSelectedArchive(null);
    setStats(null);
    setLogs([]);
    setZoomUseSample(true);
    try {
      await loadJobData(id);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load details');
    }
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

  const downloadUrl = approveToken ? `${window.location.origin}/api/download/${approveToken}` : null;

  // Build watermark sample preview URL if available
  const sampleUrl = (() => {
    if (!selectedId || !details) return null;
    const r = (details.result ?? details.Result) as { watermarkSample?: any } | undefined;
    const s = r && (r.watermarkSample ?? (details.watermarkSample ?? undefined));
    if (!s) return null;
    if (s.path) return `/api/admin/jobs/${selectedId}/preview?path=${encodeURIComponent(s.path)}`;
    if (s.zip && s.entry) return `/api/admin/jobs/${selectedId}/preview?zip=${encodeURIComponent(s.zip)}&entry=${encodeURIComponent(s.entry)}`;
    return null;
  })();

  // Frontend fallback: if no explicit sample, use the first grid image preview
  const firstPreview = archives.length > 0 && selectedArchive ? selectedArchive.images.length > 0 ? selectedArchive.images[0].previewURL : null : null;
  const zoomSrc = zoomUseSample && sampleUrl ? sampleUrl : firstPreview;

  return (
    <div className="p-4 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Admin Panel</h1>
        <div className="space-x-2">
          <button onClick={load} className="h-8 px-3 bg-gray-100 hover:bg-gray-200 rounded">Refresh</button>
        </div>
      </div>

      {error && (
        <div className="p-3 bg-red-50 text-red-700 rounded">{error}</div>
      )}

      <div className="grid grid-cols-1 xl:grid-cols-3 gap-4">
        {/* Jobs list */}
        <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 xl:col-span-1">
          <div className="p-3 border-b border-gray-200 dark:border-gray-800 font-medium">Jobs</div>
          <div className="divide-y divide-gray-200 dark:divide-gray-800 max-h-[60vh] overflow-auto">
            {loading && <div className="p-3 text-sm text-gray-500">Loading...</div>}
            {!loading && jobs.length === 0 && (
              <div className="p-3 text-sm text-gray-500">No jobs yet</div>
            )}
            {jobs.map(job => (
              <button
                key={job.id}
                onClick={() => selectJob(job.id)}
                className={`w-full text-left p-3 hover:bg-gray-50 dark:hover:bg-gray-800 ${selectedId === job.id ? 'bg-gray-50 dark:bg-gray-800' : ''}`}
              >
                <div className="flex items-center justify-between">
                  <div>
                    <div className="font-mono text-sm">{job.id}</div>
                    <div className="text-xs text-gray-500">{new Date(job.startTime).toLocaleString()}</div>
                  </div>
                  <div className="text-sm">
                    <span className="mr-2">{job.status}</span>
                    <span className="text-gray-500">{Math.round(job.progress * 100)}%</span>
                  </div>
                </div>
              </button>
            ))}
          </div>
        </div>

        {/* Center: Preview grid with archive selector */}
        <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 xl:col-span-1">
          <div className="p-3 border-b border-gray-200 dark:border-gray-800 font-medium flex items-center justify-between">
            <span>Preview</span>
            <div className="space-x-2">
              <button
                onClick={approve}
                disabled={!selectedId}
                className="h-8 px-3 bg-blue-600 text-white rounded disabled:bg-gray-200 disabled:text-gray-500"
              >
                Approve
              </button>
              {downloadUrl && (
                <a href={downloadUrl} className="h-8 px-3 bg-emerald-600 text-white rounded inline-flex items-center" download>
                  Download
                </a>
              )}
            </div>
          </div>
          
          {/* Archive Selector */}
          {archives.length > 1 && (
            <div className="p-3 border-b border-gray-200 dark:border-gray-800">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Select Archive ({archives.length} available):
              </label>
              <select 
                value={selectedArchive?.name || ''} 
                onChange={(e) => {
                  const selected = archives.find(a => a.name === e.target.value);
                  setSelectedArchive(selected || null);
                }}
                className="w-full h-8 px-2 border border-gray-300 dark:border-gray-700 rounded text-sm bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100"
              >
                {archives.map((archive, idx) => (
                  <option key={idx} value={archive.name}>
                    {archive.name} ({archive.images.length} images) [{archive.type.toUpperCase()}]
                  </option>
                ))}
              </select>
            </div>
          )}
          
          <div className="p-3 max-h-[60vh] overflow-auto">
            {selectedId && archives.length === 0 && (
              <div className="text-sm text-gray-500">No archives found for preview</div>
            )}
            {archives.length > 0 && selectedArchive && selectedArchive.images.length === 0 && (
              <div className="text-sm text-gray-500">No images found in the selected archive</div>
            )}
            {archives.length > 0 && selectedArchive && selectedArchive.images.length > 0 && (
              <div>
                <div className="text-xs text-gray-600 dark:text-gray-400 mb-3">
                  Archive: <span className="font-medium">{selectedArchive.name}</span> 
                  ({selectedArchive.images.length} images, {selectedArchive.type.toUpperCase()})
                </div>
                <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
                  {selectedArchive.images.map((im, idx) => (
                    <div key={idx} className="flex flex-col">
                      <div className="aspect-square overflow-hidden rounded ring-1 ring-gray-200 dark:ring-gray-800">
                        <img src={im.previewURL} alt={im.name} className="w-full h-full object-cover" />
                      </div>
                      <div className="mt-1 text-xs text-gray-600 dark:text-gray-400 truncate" title={im.name}>{im.name}</div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Right: Zoomed watermark sample */}
        <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 xl:col-span-1">
          <div className="p-3 border-b border-gray-200 dark:border-gray-800 font-medium">Watermark zoom</div>
          <div className="p-3">
            {!zoomSrc && <div className="text-sm text-gray-500">No sample available</div>}
            {zoomSrc && (
              <div className="w-full h-[60vh] overflow-hidden rounded ring-1 ring-gray-200 dark:ring-gray-800 relative bg-[linear-gradient(45deg,rgba(255,255,255,0.04)_25%,transparent_25%,transparent_75%,rgba(255,255,255,0.04)_75%,rgba(255,255,255,0.04)),linear-gradient(45deg,rgba(255,255,255,0.04)_25%,transparent_25%,transparent_75%,rgba(255,255,255,0.04)_75%,rgba(255,255,255,0.04))] bg-[length:20px_20px,20px_20px] bg-[position:0_0,10px_10px] dark:bg-[linear-gradient(45deg,rgba(0,0,0,0.2)_25%,transparent_25%,transparent_75%,rgba(0,0,0,0.2)_75%,rgba(0,0,0,0.2)),linear-gradient(45deg,rgba(0,0,0,0.2)_25%,transparent_25%,transparent_75%,rgba(0,0,0,0.2)_75%,rgba(0,0,0,0.2))]">
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

        {/* Stats */}
        <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
          <div className="p-3 border-b border-gray-200 dark:border-gray-800 font-medium">Stats</div>
          <div className="p-3 text-sm">
            {!stats && <div className="text-gray-500">Select a job to view stats</div>}
            {stats && (
              <ul className="space-y-1">
                <li>Images: {stats.images}</li>
                <li>Videos: {stats.videos}</li>
                <li>Texts: {stats.texts}</li>
                <li>ZIPs: {stats.zips}</li>
                <li>Total size: {(stats.totalBytes / (1024 * 1024)).toFixed(2)} MB</li>
              </ul>
            )}
          </div>
        </div>

        {/* Logs */}
        <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 xl:col-span-2">
          <div className="p-3 border-b border-gray-200 dark:border-gray-800 font-medium">Logs</div>
          <div className="p-3 text-xs font-mono whitespace-pre-wrap break-words max-h-80 overflow-auto">
            {!selectedId && <div className="text-gray-500">Select a job to view logs</div>}
            {selectedId && logs.length === 0 && <div className="text-gray-500">No logs</div>}
            {logs.map((l, i) => (
              <div key={i} className="flex items-start gap-2">
                <span className="text-gray-400">{new Date(l.timestamp).toLocaleTimeString()}</span>
                <span className={l.level === 'ERROR' ? 'text-red-600' : ''}>{l.message}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default AdminPanel; 