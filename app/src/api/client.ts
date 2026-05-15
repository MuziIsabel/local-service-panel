import type { APIResponse, HealthzResponse, WindowsServiceDTO, ServiceActionResponse, CustomAppDTO, CustomAppLogs, CreateCustomAppRequest, UpdateCustomAppRequest, EventLogDTO } from '../types/api';

const BASE_URL = 'http://127.0.0.1:17645';
const DEFAULT_TOKEN_PATH = 'C:\\ProgramData\\LocalServicePanel\\config\\token';

// In development, we try to read the token. For production, the UI would
// read the token from a secure local file provided by the Tauri shell.

async function getToken(): Promise<string | undefined> {
  // In Tauri environment, use invoke to read token from filesystem.
  const isTauri = typeof window !== 'undefined' && '__TAURI__' in window;
  if (isTauri) {
    try {
      const { invoke } = await import('@tauri-apps/api/core');
      const token = await invoke<string>('read_token_file', { path: DEFAULT_TOKEN_PATH });
      if (token) return token;
    } catch {
      // Token file not found or not in Tauri - fall through.
    }
  }

  // For Vite dev mode, try environment variable via import.meta.env.
  const devToken = import.meta.env.VITE_DEV_TOKEN as string | undefined;
  if (devToken) return devToken;

  // Final fallback: try localStorage for convenience.
  return undefined;
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  const token = await getToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: { ...headers, ...(options?.headers as Record<string, string>) },
  });

  if (!res.ok) {
    const errorBody = await res.text();
    let message = `HTTP ${res.status}`;
    try {
      const parsed = JSON.parse(errorBody);
      message = parsed.error?.message || message;
    } catch {
      // ignore
    }
    throw new Error(message);
  }

  return res.json() as Promise<T>;
}

export async function checkHealthz(): Promise<APIResponse<HealthzResponse>> {
  return request<APIResponse<HealthzResponse>>('/api/healthz');
}

export async function listServices(params?: {
  keyword?: string;
  status?: string;
  startType?: string;
  includeProtected?: boolean;
}): Promise<APIResponse<WindowsServiceDTO[]>> {
  const searchParams = new URLSearchParams();
  if (params?.keyword) searchParams.set('keyword', params.keyword);
  if (params?.status) searchParams.set('status', params.status);
  if (params?.startType) searchParams.set('startType', params.startType);
  if (params?.includeProtected !== undefined) searchParams.set('includeProtected', String(params.includeProtected));

  const qs = searchParams.toString();
  return request<APIResponse<WindowsServiceDTO[]>>(`/api/windows/services${qs ? `?${qs}` : ''}`);
}

export async function startService(serviceName: string): Promise<APIResponse<ServiceActionResponse>> {
  return request<APIResponse<ServiceActionResponse>>(`/api/windows/services/${encodeURIComponent(serviceName)}/start`, {
    method: 'POST',
  });
}

export async function stopService(serviceName: string): Promise<APIResponse<ServiceActionResponse>> {
  return request<APIResponse<ServiceActionResponse>>(`/api/windows/services/${encodeURIComponent(serviceName)}/stop`, {
    method: 'POST',
  });
}

export async function restartService(serviceName: string): Promise<APIResponse<ServiceActionResponse>> {
  return request<APIResponse<ServiceActionResponse>>(`/api/windows/services/${encodeURIComponent(serviceName)}/restart`, {
    method: 'POST',
  });
}

export async function setStartType(serviceName: string, startType: string): Promise<APIResponse<ServiceActionResponse>> {
  return request<APIResponse<ServiceActionResponse>>(`/api/windows/services/${encodeURIComponent(serviceName)}/start-type`, {
    method: 'PATCH',
    body: JSON.stringify({ startType }),
  });
}

// --- Custom App API ---

export async function listCustomApps(keyword?: string): Promise<APIResponse<CustomAppDTO[]>> {
  const qs = keyword ? `?keyword=${encodeURIComponent(keyword)}` : '';
  return request<APIResponse<CustomAppDTO[]>>(`/api/custom-apps${qs}`);
}

export async function getCustomApp(id: string): Promise<APIResponse<CustomAppDTO>> {
  return request<APIResponse<CustomAppDTO>>(`/api/custom-apps/${encodeURIComponent(id)}`);
}

export async function createCustomApp(body: CreateCustomAppRequest): Promise<APIResponse<CustomAppDTO>> {
  return request<APIResponse<CustomAppDTO>>('/api/custom-apps', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export async function updateCustomApp(id: string, body: UpdateCustomAppRequest): Promise<APIResponse<CustomAppDTO>> {
  return request<APIResponse<CustomAppDTO>>(`/api/custom-apps/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
  });
}

export async function deleteCustomApp(id: string): Promise<APIResponse<{status: string}>> {
  return request<APIResponse<{status: string}>>(`/api/custom-apps/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  });
}

export async function startCustomApp(id: string): Promise<APIResponse<CustomAppDTO>> {
  return request<APIResponse<CustomAppDTO>>(`/api/custom-apps/${encodeURIComponent(id)}/start`, {
    method: 'POST',
  });
}

export async function stopCustomApp(id: string): Promise<APIResponse<CustomAppDTO>> {
  return request<APIResponse<CustomAppDTO>>(`/api/custom-apps/${encodeURIComponent(id)}/stop`, {
    method: 'POST',
  });
}

export async function getCustomAppLogs(id: string, lines?: number): Promise<APIResponse<CustomAppLogs>> {
  const qs = lines ? `?lines=${lines}` : '';
  return request<APIResponse<CustomAppLogs>>(`/api/custom-apps/${encodeURIComponent(id)}/logs${qs}`);
}

export async function setCustomAppAutoStart(id: string, enabled: boolean): Promise<APIResponse<CustomAppDTO>> {
  return request<APIResponse<CustomAppDTO>>(`/api/custom-apps/${encodeURIComponent(id)}/autostart`, {
    method: 'POST',
    body: JSON.stringify({ enabled }),
  });
}

export async function listEvents(params?: {
  limit?: number;
  targetId?: string;
  targetType?: string;
  action?: string;
  status?: string;
}): Promise<APIResponse<EventLogDTO[]>> {
  const searchParams = new URLSearchParams();
  if (params?.limit) searchParams.set('limit', String(params.limit));
  if (params?.targetId) searchParams.set('targetId', params.targetId);
  if (params?.targetType) searchParams.set('targetType', params.targetType);
  if (params?.action) searchParams.set('action', params.action);
  if (params?.status) searchParams.set('status', params.status);

  const qs = searchParams.toString();
  return request<APIResponse<EventLogDTO[]>>(`/api/events${qs ? `?${qs}` : ''}`);
}
