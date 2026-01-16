// Chat message component

import React from 'react';
import ReactMarkdown from 'react-markdown';
import type { ChatMessage as ChatMessageType } from '../types';
import { User, Bot, AlertCircle } from 'lucide-react';

interface ChatMessageProps {
  message: ChatMessageType;
}

const roleConfig = {
  user: {
    icon: User,
    bgColor: 'bg-blue-50',
    borderColor: 'border-blue-200',
    iconColor: 'text-blue-600',
    label: 'You',
  },
  jordan: {
    icon: Bot,
    bgColor: 'bg-purple-50',
    borderColor: 'border-purple-200',
    iconColor: 'text-purple-600',
    label: 'Jordan',
  },
  system: {
    icon: AlertCircle,
    bgColor: 'bg-red-50',
    borderColor: 'border-red-200',
    iconColor: 'text-red-600',
    label: 'System',
  },
  panel: {
    icon: Bot,
    bgColor: 'bg-green-50',
    borderColor: 'border-green-200',
    iconColor: 'text-green-600',
    label: 'Panel',
  },
};

export const ChatMessage: React.FC<ChatMessageProps> = ({ message }) => {
  const config = roleConfig[message.role as keyof typeof roleConfig] || roleConfig.jordan;
  const Icon = config.icon;

  const isEscalated = message.metadata?.escalated;

  return (
    <div
      className={`flex gap-3 p-4 rounded-lg border ${config.bgColor} ${config.borderColor} ${
        isEscalated ? 'ring-2 ring-green-400' : ''
      }`}
    >
      <div className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center ${config.bgColor}`}>
        <Icon className={`w-5 h-5 ${config.iconColor}`} />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <span className={`text-sm font-medium ${config.iconColor}`}>{config.label}</span>
          {message.metadata?.phase && (
            <span className="text-xs text-gray-500 bg-gray-100 px-2 py-0.5 rounded">
              {message.metadata.phase.replace('_', ' ')}
            </span>
          )}
          {isEscalated && (
            <span className="text-xs text-green-600 bg-green-100 px-2 py-0.5 rounded font-medium">
              Ready for Panel
            </span>
          )}
        </div>
        <div className="prose prose-sm max-w-none text-gray-700">
          <ReactMarkdown>{message.content}</ReactMarkdown>
        </div>
        <div className="text-xs text-gray-400 mt-2">
          {new Date(message.timestamp).toLocaleTimeString()}
        </div>
      </div>
    </div>
  );
};

export default ChatMessage;
