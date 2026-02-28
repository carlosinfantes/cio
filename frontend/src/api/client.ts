// API Client for CIO - Chief Intelligence Officer

import type {
  Session,
  ChatResponse,
  CRFEntity,
  DRFDocument,
} from '../types';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8765';

class APIClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // === Session Management ===

  async createSession(domain: string = 'cio'): Promise<{ session_id: string }> {
    return this.request('/api/v1/session', {
      method: 'POST',
      body: JSON.stringify({ domain }),
    });
  }

  async getSession(sessionId: string): Promise<Session> {
    return this.request(`/api/v1/session/${sessionId}`);
  }

  async listSessions(): Promise<{ sessions: Session[] }> {
    return this.request('/api/v1/session');
  }

  async deleteSession(sessionId: string): Promise<void> {
    await this.request(`/api/v1/session/${sessionId}`, {
      method: 'DELETE',
    });
  }

  // === Chat ===

  async sendMessage(sessionId: string, content: string): Promise<ChatResponse> {
    return this.request(`/api/v1/chat/${sessionId}/message`, {
      method: 'POST',
      body: JSON.stringify({ content }),
    });
  }

  // === Context (CRF) ===

  async getContext(): Promise<{ entities: CRFEntity[] }> {
    return this.request('/api/v1/context');
  }

  async createEntity(entity: Partial<CRFEntity>): Promise<{ id: string }> {
    return this.request('/api/v1/context', {
      method: 'POST',
      body: JSON.stringify(entity),
    });
  }

  async getEntity(id: string): Promise<CRFEntity> {
    return this.request(`/api/v1/context/${id}`);
  }

  // === Decisions (DRF) ===

  async listDecisions(filters?: {
    status?: string;
    tag?: string;
    domain?: string;
  }): Promise<{ decisions: DRFDocument[] }> {
    const params = new URLSearchParams();
    if (filters?.status) params.set('status', filters.status);
    if (filters?.tag) params.set('tag', filters.tag);
    if (filters?.domain) params.set('domain', filters.domain);

    const query = params.toString();
    return this.request(`/api/v1/decisions${query ? `?${query}` : ''}`);
  }

  async getDecision(id: string): Promise<DRFDocument> {
    return this.request(`/api/v1/decisions/${id}`);
  }

  async updateDecisionStatus(id: string, status: string): Promise<void> {
    await this.request(`/api/v1/decisions/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    });
  }

  // === Panel Direct Access ===

  async askPanel(
    question: string,
    options?: {
      advisors?: string[];
      mode?: string;
    }
  ): Promise<unknown> {
    return this.request('/api/v1/panel/ask', {
      method: 'POST',
      body: JSON.stringify({
        question,
        ...options,
      }),
    });
  }

  // === Health Check ===

  async healthCheck(): Promise<{ status: string }> {
    return this.request('/api/health');
  }
}

// Export singleton instance
export const api = new APIClient();

// Export class for custom instances
export { APIClient };
