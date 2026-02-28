// React hook for chat functionality

import { useState, useCallback, useRef, useEffect } from 'react';
import { api } from '../api/client';
import type { ChatMessage, ChatResponse, FacilitationPhase, Brief } from '../types';

interface UseChatOptions {
  domain?: string;
  onEscalation?: (brief: Brief) => void;
  onError?: (error: Error) => void;
}

interface UseChatReturn {
  messages: ChatMessage[];
  isLoading: boolean;
  error: Error | null;
  sessionId: string | null;
  phase: FacilitationPhase;
  isReadyForPanel: boolean;
  brief: Brief | null;
  sendMessage: (content: string) => Promise<void>;
  startSession: () => Promise<void>;
  clearSession: () => void;
}

export function useChat(options: UseChatOptions = {}): UseChatReturn {
  const { domain = 'cio', onEscalation, onError } = options;

  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [phase, setPhase] = useState<FacilitationPhase>('init');
  const [isReadyForPanel, setIsReadyForPanel] = useState(false);
  const [brief, setBrief] = useState<Brief | null>(null);

  const messageIdCounter = useRef(0);

  const generateMessageId = useCallback(() => {
    messageIdCounter.current += 1;
    return `msg-${Date.now()}-${messageIdCounter.current}`;
  }, []);

  const startSession = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const result = await api.createSession(domain);
      setSessionId(result.session_id);
      setMessages([]);
      setPhase('init');
      setIsReadyForPanel(false);
      setBrief(null);

      // Add initial Jordan greeting
      setMessages([
        {
          id: generateMessageId(),
          role: 'jordan',
          content: "Hi! I'm Jordan, your discovery facilitator. I'll help you articulate your technical decision clearly before consulting the advisory panel. What's on your mind?",
          timestamp: new Date().toISOString(),
          metadata: { phase: 'init' },
        },
      ]);
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      setError(error);
      onError?.(error);
    } finally {
      setIsLoading(false);
    }
  }, [domain, generateMessageId, onError]);

  const sendMessage = useCallback(
    async (content: string) => {
      if (!sessionId || !content.trim()) return;

      // Add user message immediately
      const userMessage: ChatMessage = {
        id: generateMessageId(),
        role: 'user',
        content: content.trim(),
        timestamp: new Date().toISOString(),
      };

      setMessages((prev) => [...prev, userMessage]);
      setIsLoading(true);
      setError(null);

      try {
        const response: ChatResponse = await api.sendMessage(sessionId, content);

        // Update phase and readiness
        setPhase(response.state.phase);
        setIsReadyForPanel(response.state.ready_for_panel);

        // Add Jordan's response
        const jordanMessage: ChatMessage = {
          id: generateMessageId(),
          role: response.speaker,
          content: response.response,
          timestamp: new Date().toISOString(),
          metadata: {
            phase: response.state.phase,
            ready_for_panel: response.state.ready_for_panel,
            escalated: response.escalated,
          },
        };

        setMessages((prev) => [...prev, jordanMessage]);

        // Handle escalation
        if (response.escalated && response.brief) {
          setBrief(response.brief);
          onEscalation?.(response.brief);
        }
      } catch (err) {
        const error = err instanceof Error ? err : new Error(String(err));
        setError(error);
        onError?.(error);

        // Add error message
        setMessages((prev) => [
          ...prev,
          {
            id: generateMessageId(),
            role: 'system',
            content: `Error: ${error.message}`,
            timestamp: new Date().toISOString(),
          },
        ]);
      } finally {
        setIsLoading(false);
      }
    },
    [sessionId, generateMessageId, onEscalation, onError]
  );

  const clearSession = useCallback(() => {
    if (sessionId) {
      api.deleteSession(sessionId).catch(console.error);
    }
    setSessionId(null);
    setMessages([]);
    setPhase('init');
    setIsReadyForPanel(false);
    setBrief(null);
    setError(null);
  }, [sessionId]);

  // Auto-start session on mount
  useEffect(() => {
    if (!sessionId) {
      startSession();
    }
  }, [sessionId, startSession]);

  return {
    messages,
    isLoading,
    error,
    sessionId,
    phase,
    isReadyForPanel,
    brief,
    sendMessage,
    startSession,
    clearSession,
  };
}
