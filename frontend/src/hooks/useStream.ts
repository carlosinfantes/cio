// React hook for Server-Sent Events (SSE) streaming

import { useState, useEffect, useCallback, useRef } from 'react';

interface StreamEvent {
  event: string;
  data: unknown;
}

interface UseStreamOptions {
  onThinking?: () => void;
  onChunk?: (content: string, complete: boolean) => void;
  onComplete?: (data: unknown) => void;
  onEscalation?: (brief: unknown) => void;
  onError?: (error: Error) => void;
  onHeartbeat?: () => void;
}

interface UseStreamReturn {
  isConnected: boolean;
  isStreaming: boolean;
  connect: (sessionId: string) => void;
  disconnect: () => void;
  currentContent: string;
}

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8765';

export function useStream(options: UseStreamOptions = {}): UseStreamReturn {
  const {
    onThinking,
    onChunk,
    onComplete,
    onEscalation,
    onError,
    onHeartbeat,
  } = options;

  const [isConnected, setIsConnected] = useState(false);
  const [isStreaming, setIsStreaming] = useState(false);
  const [currentContent, setCurrentContent] = useState('');

  const eventSourceRef = useRef<EventSource | null>(null);
  const sessionIdRef = useRef<string | null>(null);

  const connect = useCallback((sessionId: string) => {
    // Disconnect existing connection
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    sessionIdRef.current = sessionId;
    const url = `${API_BASE}/api/v1/stream/${sessionId}`;
    const eventSource = new EventSource(url);

    eventSource.onopen = () => {
      setIsConnected(true);
    };

    eventSource.onerror = (event) => {
      console.error('SSE error:', event);
      setIsConnected(false);
      setIsStreaming(false);
      onError?.(new Error('Stream connection error'));
    };

    // Handle specific event types
    eventSource.addEventListener('connected', (event) => {
      try {
        const data = JSON.parse(event.data);
        console.log('Stream connected:', data);
        setIsConnected(true);
      } catch {
        // Ignore parse errors
      }
    });

    eventSource.addEventListener('thinking', () => {
      setIsStreaming(true);
      setCurrentContent('');
      onThinking?.();
    });

    eventSource.addEventListener('chunk', (event) => {
      try {
        const data = JSON.parse(event.data) as { content: string; complete: boolean };
        setCurrentContent(data.content);
        onChunk?.(data.content, data.complete);
      } catch {
        // Ignore parse errors
      }
    });

    eventSource.addEventListener('complete', (event) => {
      try {
        const data = JSON.parse(event.data);
        setIsStreaming(false);
        onComplete?.(data);
      } catch {
        // Ignore parse errors
      }
    });

    eventSource.addEventListener('escalation', (event) => {
      try {
        const data = JSON.parse(event.data);
        onEscalation?.(data);
      } catch {
        // Ignore parse errors
      }
    });

    eventSource.addEventListener('error', (event) => {
      try {
        const data = JSON.parse((event as MessageEvent).data) as { message: string };
        setIsStreaming(false);
        onError?.(new Error(data.message));
      } catch {
        onError?.(new Error('Unknown streaming error'));
      }
    });

    eventSource.addEventListener('heartbeat', () => {
      onHeartbeat?.();
    });

    eventSourceRef.current = eventSource;
  }, [onThinking, onChunk, onComplete, onEscalation, onError, onHeartbeat]);

  const disconnect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    sessionIdRef.current = null;
    setIsConnected(false);
    setIsStreaming(false);
    setCurrentContent('');
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return {
    isConnected,
    isStreaming,
    connect,
    disconnect,
    currentContent,
  };
}

export default useStream;
