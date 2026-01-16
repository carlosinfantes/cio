// Main chat panel component

import React, { useRef, useEffect } from 'react';
import { ChatMessage } from './ChatMessage';
import { ChatInput } from './ChatInput';
import { PhaseIndicator } from './PhaseIndicator';
import { useChat } from '../hooks/useChat';
import { RefreshCw, MessageSquare } from 'lucide-react';

interface ChatPanelProps {
  domain?: string;
}

export const ChatPanel: React.FC<ChatPanelProps> = ({ domain = 'cto-advisory' }) => {
  const {
    messages,
    isLoading,
    error,
    phase,
    isReadyForPanel,
    brief,
    sendMessage,
    startSession,
    clearSession,
  } = useChat({ domain });

  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  return (
    <div className="flex flex-col h-full bg-white rounded-xl shadow-lg overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-6 py-4 border-b bg-gradient-to-r from-purple-600 to-purple-800">
        <div className="flex items-center gap-3">
          <MessageSquare className="w-6 h-6 text-white" />
          <div>
            <h1 className="text-lg font-semibold text-white">CTO Advisory Board</h1>
            <p className="text-sm text-purple-200">AI-powered technical decisions</p>
          </div>
        </div>
        <button
          onClick={() => {
            clearSession();
            startSession();
          }}
          className="p-2 rounded-lg hover:bg-white/10 transition-colors"
          title="New Session"
        >
          <RefreshCw className="w-5 h-5 text-white" />
        </button>
      </div>

      {/* Phase Indicator */}
      <PhaseIndicator phase={phase} isReadyForPanel={isReadyForPanel} />

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((message) => (
          <ChatMessage key={message.id} message={message} />
        ))}

        {/* Brief Display */}
        {brief && (
          <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
            <h3 className="font-semibold text-green-800 mb-2">Discovery Brief</h3>
            <div className="space-y-2 text-sm">
              <div>
                <span className="font-medium text-green-700">Problem: </span>
                <span className="text-green-900">{brief.problem_statement}</span>
              </div>
              {brief.constraints.length > 0 && (
                <div>
                  <span className="font-medium text-green-700">Constraints: </span>
                  <span className="text-green-900">{brief.constraints.join(', ')}</span>
                </div>
              )}
              {brief.goals.length > 0 && (
                <div>
                  <span className="font-medium text-green-700">Goals: </span>
                  <span className="text-green-900">{brief.goals.join(', ')}</span>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Loading indicator */}
        {isLoading && (
          <div className="flex items-center gap-2 text-gray-500">
            <div className="w-2 h-2 bg-purple-500 rounded-full animate-bounce" />
            <div className="w-2 h-2 bg-purple-500 rounded-full animate-bounce [animation-delay:0.1s]" />
            <div className="w-2 h-2 bg-purple-500 rounded-full animate-bounce [animation-delay:0.2s]" />
            <span className="text-sm ml-2">Jordan is thinking...</span>
          </div>
        )}

        {/* Error display */}
        {error && (
          <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
            {error.message}
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="p-4 border-t bg-gray-50">
        <ChatInput
          onSend={sendMessage}
          isLoading={isLoading}
          placeholder={
            isReadyForPanel
              ? 'Ask the panel a question or continue discovery...'
              : 'Describe your technical decision or challenge...'
          }
        />
      </div>
    </div>
  );
};

export default ChatPanel;
