import { 
  EncryptRequest, 
  DecryptRequest, 
  BatchCopyRequest, 
  AddTextRequest, 
  RemoveWatermarksRequest, 
  ProcessingJob,
  User,
  UserStats,
  Subscription,
  SubscriptionPlan,
  Payment,
  UserUsage
} from '../types';

const API_BASE = `${window.location.origin}/api`;

interface ApiResponse {
  success: boolean;
  message?: string;
  jobId?: string;
  error?: string;
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
      credentials: 'include',
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
  const response = await fetch(`${API_BASE}/download/${token}`, { credentials: 'include' });
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

// Auth API
export async function register(email: string, password: string): Promise<{ id: string; email: string }> {
  return fetchApi('/auth/register', { method: 'POST', body: JSON.stringify({ email, password }) });
}

export async function login(email: string, password: string): Promise<{ id: string; email: string }> {
  return fetchApi('/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) });
}

export async function logout(): Promise<{ success: boolean }> {
  return fetchApi('/auth/logout', { method: 'POST' });
}

export async function me(): Promise<User | { error: string }> {
  return fetchApi('/auth/me');
}

// Super Admin API
export async function adminUsers(): Promise<{ users: User[] }> {
  return fetchApi('/admin/users');
}

export async function adminUserStats(): Promise<UserStats> {
  return fetchApi('/admin/users/stats');
}

// Subscription API
export async function getMySubscription(): Promise<{ subscription: Subscription; plan: SubscriptionPlan }> {
  return fetchApi('/subscription/my');
}

export async function getSubscriptionPlans(): Promise<{ 
  plans: SubscriptionPlan[]; 
  currencies: { code: string; name: string; symbol: string }[] 
}> {
  return fetchApi('/subscription/plans');
}

export async function getMyUsage(): Promise<{ 
  usage: UserUsage; 
  limits: { processing_jobs: number; max_file_size: number } 
}> {
  return fetchApi('/subscription/usage');
}

export async function createCryptoPayment(planType: string, currency: string): Promise<{
  payment_id: string;
  payment_url: string;
  crypto_address: string;
  crypto_amount: string;
  currency: string;
  expires_at: string;
  status: string;
}> {
  return fetchApi('/subscription/payment/crypto', {
    method: 'POST',
    body: JSON.stringify({ plan_type: planType, currency })
  });
}

export async function getPaymentStatus(paymentId: string): Promise<{
  payment_id: string;
  status: string;
  amount: number;
  currency: string;
  crypto_address: string;
  crypto_amount: string;
  created_at: string;
  expires_at?: string;
  paid_at?: string;
}> {
  return fetchApi(`/subscription/payment/${paymentId}`);
}

export async function completeMockPayment(paymentId: string): Promise<{ success: boolean; message: string }> {
  return fetchApi(`/subscription/payment/mock/${paymentId}/complete`, { method: 'POST' });
}

// Admin Subscription API
export async function adminGetAllSubscriptions(): Promise<{ subscriptions: Subscription[] }> {
  return fetchApi('/admin/subscription/all');
}

export async function adminGetSubscriptionStats(): Promise<{ stats: any }> {
  return fetchApi('/admin/subscription/stats');
}

export async function adminExtendSubscription(userId: string, planType: string, days: number): Promise<{
  success: boolean;
  message: string;
  subscription: Subscription;
}> {
  return fetchApi('/admin/subscription/extend', {
    method: 'POST',
    body: JSON.stringify({ user_id: userId, plan_type: planType, days })
  });
}