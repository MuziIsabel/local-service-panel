export interface APIResponse<T> {
  data: T;
}

export interface APIError {
  error: {
    code: string;
    message: string;
    details?: string;
  };
}

export interface HealthzResponse {
  status: string;
  version: string;
}

export type AgentStatus = 'connected' | 'disconnected' | 'checking';

export interface WindowsServiceDTO {
  id: string;
  serviceName: string;
  displayName: string;
  description?: string;
  status: string;
  startType: string;
  executablePath?: string;
  canStop: boolean;
  canPauseAndContinue: boolean;
  protected: boolean;
}

export interface ServiceActionResponse {
  serviceName: string;
  status?: string;
  startType?: string;
}

export type ServiceStatus = 'running' | 'stopped' | 'start_pending' | 'stop_pending' | 'pause_pending' | 'paused' | 'continue_pending' | 'unknown';

export type ServiceStartType = 'automatic' | 'automatic_delayed' | 'manual' | 'disabled' | 'unknown';

// --- Custom App Types ---

export interface CustomAppDTO {
  id: string;
  name: string;
  type: string;
  status: string;
  executablePath: string;
  workingDir?: string;
  args?: string[];
  autoStart: boolean;
  pid?: number;
  lastStartedAt?: string;
  lastStoppedAt?: string;
  lastError?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CustomAppLogs {
  stdout: string[];
  stderr: string[];
}

export interface CreateCustomAppRequest {
  name: string;
  executablePath: string;
  workingDir?: string;
  args?: string[];
  autoStart?: boolean;
}

export interface UpdateCustomAppRequest {
  name?: string;
  executablePath?: string;
  workingDir?: string;
  args?: string[];
  autoStart?: boolean;
}

// --- Event Log Types ---

export interface EventLogDTO {
  id: string;
  targetId?: string;
  targetType?: string;
  action: string;
  status: string;
  message?: string;
  details?: string;
  createdAt: string;
}
