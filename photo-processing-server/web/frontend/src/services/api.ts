import { BatchCopySettings, AddTextSettings } from '../types';

const API_BASE = `${window.location.origin}/api`;

interface ApiResponse {
  success: boolean;
  message?: string;
  jobId?: string;
  error?: string;
}

interface EncryptRequest {
  selectedPath: string;
  nameToInject: string;
}

interface DecryptRequest {
  selectedPath: string;
}

interface BatchCopyRequest {
  selectedPath: string;
  settings: BatchCopySettings;
}

interface AddTextRequest {
  selectedPath: string;
  settings: AddTextSettings;
}

interface RemoveWatermarksRequest {
  selectedPath: string;
}

class ApiError extends Error {
  constructor(message: string, public status?: number) {
    super(message);
    this.name = 'ApiError';
  }
}

async function fetchApi<T>(endpoint: string, options?: RequestInit): Promise<T> {
  try {
    const isFormData = options && options.body instanceof FormData;
    const baseHeaders: Record<string, string> = isFormData ? {} : { 'Content-Type': 'application/json' };
    const headers = { ...baseHeaders, ...(options?.headers as Record<string, string> | undefined) };

    const response = await fetch(`${API_BASE}${endpoint}`, {
      headers,
      ...options,
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new ApiError(
        errorText || `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    return await response.json();
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }
    
    // Network or parsing error
    throw new ApiError(
      error instanceof Error ? error.message : 'Unknown error occurred'
    );
  }
}

export async function encrypt(request: EncryptRequest): Promise<ApiResponse> {
  return fetchApi<ApiResponse>('/encrypt', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

export async function decrypt(request: DecryptRequest): Promise<ApiResponse> {
  return fetchApi<ApiResponse>('/decrypt', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

export async function batchCopy(request: BatchCopyRequest): Promise<ApiResponse> {
  return fetchApi<ApiResponse>('/batch-copy', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

export async function addText(request: AddTextRequest): Promise<ApiResponse> {
  return fetchApi<ApiResponse>('/add-text', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

export async function removeWatermarks(request: RemoveWatermarksRequest): Promise<ApiResponse> {
  return fetchApi<ApiResponse>('/remove-watermarks', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

export async function uploadFiles(files: FileList, folderName?: string): Promise<ApiResponse> {
  const formData = new FormData();
  Array.from(files).forEach((file, index) => {
    formData.append(`file_${index}`, file);
  });
  
  // Add folder name if provided
  if (folderName) {
    formData.append('folderName', folderName);
  }

  return fetchApi<ApiResponse>('/upload', {
    method: 'POST',
    body: formData,
  });
}

export async function getProcessingStatus(jobId: string): Promise<{
  status: string;
  progress: number;
  message?: string;
  result?: any;
  error?: string;
}> {
  return fetchApi(`/processing/${jobId}`);
}

export async function downloadResult(token: string): Promise<Blob> {
  const response = await fetch(`${API_BASE}/download/${token}`);
  if (!response.ok) {
    throw new ApiError(`Download failed: ${response.statusText}`, response.status);
  }
  return response.blob();
}

// Admin API
export interface JobSummary {
  id: string;
  status: string;
  progress: number;
  startTime: string;
}

export async function adminListJobs(): Promise<{ jobs: JobSummary[] }> {
  return fetchApi('/admin/jobs');
}

export async function adminJobDetails(id: string): Promise<any> {
  return fetchApi(`/admin/jobs/${id}`);
}

export async function adminApprove(id: string): Promise<{ success: boolean; token: string }> {
  return fetchApi(`/admin/jobs/${id}/approve`, { method: 'POST' });
}

export interface AdminImagePreview { 
  name: string; 
  previewURL: string; 
}

export interface AdminArchive {
  name: string;
  path: string;
  type: 'zip' | 'folder';
  images: AdminImagePreview[];
}

export async function adminImages(id: string): Promise<{ archives: AdminArchive[] }> {
  return fetchApi(`/admin/jobs/${id}/images`);
}

export function adminPreviewUrl(id: string, qs: string): string {
  return `${API_BASE}/admin/jobs/${id}/preview?${qs}`;
}

export async function adminStats(id: string): Promise<{ stats: { images: number; videos: number; texts: number; zips: number; totalBytes: number } }> {
  return fetchApi(`/admin/jobs/${id}/stats`);
}

export interface AdminLog { message: string; timestamp: string; level: string }
export async function adminLogs(id: string): Promise<{ logs: AdminLog[] }> {
  return fetchApi(`/admin/jobs/${id}/logs`);
}