// Types for CTO Advisory Board Frontend

export interface Session {
  id: string;
  domain: string;
  created_at: string;
  updated_at: string;
  messages: ChatMessage[];
  state: FacilitationPhase;
}

export interface ChatMessage {
  id: string;
  role: 'user' | 'jordan' | 'panel' | string;
  content: string;
  timestamp: string;
  metadata?: {
    phase?: FacilitationPhase;
    ready_for_panel?: boolean;
    escalated?: boolean;
    advisor_id?: string;
    advisor_name?: string;
  };
}

export type FacilitationPhase =
  | 'init'
  | 'context_gathering'
  | 'problem_articulation'
  | 'discovery'
  | 'ready_for_escalation'
  | 'escalated';

export interface Brief {
  problem_statement: string;
  context: string;
  constraints: string[];
  goals: string[];
  key_questions: string[];
  suggested_advisors: string[];
}

export interface CRFEntity {
  id: string;
  type: string;
  name: string;
  description?: string;
  attributes?: Record<string, unknown>;
  tags?: string[];
}

export interface DRFDocument {
  drf_version: string;
  decision: {
    id: string;
    title: string;
    domain: string;
    intent: string;
  };
  context: {
    constraints: Array<{ description: string; negotiable: boolean }>;
    objectives: Array<{ description: string; priority: string }>;
  };
  cognitive_state: {
    phase: string;
    confidence: number;
  };
  synthesis: {
    decision: string;
    rationale: string;
  };
  meta: {
    created_at: string;
    status: string;
    tags: string[];
  };
}

export interface AdvisorResponse {
  advisor_id: string;
  name: string;
  role: string;
  response: string;
  emoji?: string;
  color?: string;
}

export interface ParsedResponse {
  advisors: AdvisorResponse[];
  synthesis: string;
}

export interface ChatResponse {
  response: string;
  speaker: string;
  state: {
    phase: FacilitationPhase;
    ready_for_panel: boolean;
  };
  escalated?: boolean;
  brief?: Brief;
  suggested_mode?: string;
}

export interface APIError {
  message: string;
  code?: string;
}
