// Phase indicator component showing facilitation progress

import React from 'react';
import { Check, Circle, ArrowRight } from 'lucide-react';
import type { FacilitationPhase } from '../types';

interface PhaseIndicatorProps {
  phase: FacilitationPhase;
  isReadyForPanel: boolean;
}

const phases = [
  { id: 'context_gathering', label: 'Context', short: 'Ctx' },
  { id: 'problem_articulation', label: 'Problem', short: 'Prb' },
  { id: 'discovery', label: 'Discovery', short: 'Disc' },
  { id: 'ready_for_escalation', label: 'Panel', short: 'Panel' },
];

const phaseOrder: Record<FacilitationPhase, number> = {
  init: 0,
  context_gathering: 1,
  problem_articulation: 2,
  discovery: 3,
  ready_for_escalation: 4,
  escalated: 5,
};

export const PhaseIndicator: React.FC<PhaseIndicatorProps> = ({ phase, isReadyForPanel }) => {
  const currentPhaseIndex = phaseOrder[phase] || 0;

  return (
    <div className="px-6 py-3 bg-gray-50 border-b">
      <div className="flex items-center justify-between max-w-md mx-auto">
        {phases.map((p, index) => {
          const phaseIndex = phaseOrder[p.id as FacilitationPhase];
          const isComplete = currentPhaseIndex > phaseIndex;
          const isCurrent = p.id === phase || (phase === 'init' && index === 0);
          const isEscalated = phase === 'escalated' && index === phases.length - 1;

          return (
            <React.Fragment key={p.id}>
              <div className="flex flex-col items-center">
                <div
                  className={`w-8 h-8 rounded-full flex items-center justify-center transition-colors
                    ${isComplete || isEscalated ? 'bg-green-500 text-white' : ''}
                    ${isCurrent && !isComplete ? 'bg-purple-500 text-white ring-4 ring-purple-100' : ''}
                    ${!isComplete && !isCurrent ? 'bg-gray-200 text-gray-500' : ''}`}
                >
                  {isComplete || isEscalated ? (
                    <Check className="w-4 h-4" />
                  ) : (
                    <Circle className="w-3 h-3" fill={isCurrent ? 'currentColor' : 'none'} />
                  )}
                </div>
                <span
                  className={`text-xs mt-1 font-medium
                    ${isComplete || isEscalated ? 'text-green-600' : ''}
                    ${isCurrent && !isComplete ? 'text-purple-600' : ''}
                    ${!isComplete && !isCurrent ? 'text-gray-400' : ''}`}
                >
                  {p.short}
                </span>
              </div>
              {index < phases.length - 1 && (
                <ArrowRight
                  className={`w-4 h-4 mx-1 ${
                    currentPhaseIndex > phaseIndex + 1 ? 'text-green-400' : 'text-gray-300'
                  }`}
                />
              )}
            </React.Fragment>
          );
        })}
      </div>

      {/* Status text */}
      <div className="text-center mt-2">
        <span className="text-xs text-gray-500">
          {isReadyForPanel ? (
            <span className="text-green-600 font-medium">
              Ready to consult the advisory panel
            </span>
          ) : (
            getPhaseDescription(phase)
          )}
        </span>
      </div>
    </div>
  );
};

function getPhaseDescription(phase: FacilitationPhase): string {
  switch (phase) {
    case 'init':
      return 'Starting discovery session...';
    case 'context_gathering':
      return 'Jordan is learning about your context';
    case 'problem_articulation':
      return 'Help Jordan understand the problem';
    case 'discovery':
      return 'Clarifying details before panel';
    case 'ready_for_escalation':
      return 'Ready to bring in the advisory panel';
    case 'escalated':
      return 'Advisory panel engaged';
    default:
      return '';
  }
}

export default PhaseIndicator;
