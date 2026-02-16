'use client';

import { useState, useRef, useId, useCallback } from 'react';

import { cn } from '@/lib/utils/cn';

const TOOLTIP_POSITIONS = {
  top: 'bottom-full left-1/2 -translate-x-1/2 mb-2',
  bottom: 'top-full left-1/2 -translate-x-1/2 mt-2',
  left: 'right-full top-1/2 -translate-y-1/2 mr-2',
  right: 'left-full top-1/2 -translate-y-1/2 ml-2',
} as const;

type TooltipPosition = keyof typeof TOOLTIP_POSITIONS;

const ARROW_POSITIONS = {
  top: 'top-full left-1/2 -translate-x-1/2 border-t-tooltip border-x-transparent border-b-transparent',
  bottom: 'bottom-full left-1/2 -translate-x-1/2 border-b-tooltip border-x-transparent border-t-transparent',
  left: 'left-full top-1/2 -translate-y-1/2 border-l-tooltip border-y-transparent border-r-transparent',
  right: 'right-full top-1/2 -translate-y-1/2 border-r-tooltip border-y-transparent border-l-transparent',
} as const;

interface TooltipProps {
  content: string;
  position?: TooltipPosition;
  delayMs?: number;
  children: React.ReactNode;
}

export function Tooltip({
  content,
  position = 'top',
  delayMs = 200,
  children,
}: TooltipProps) {
  const [isVisible, setIsVisible] = useState(false);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const tooltipId = useId();

  const show = useCallback(() => {
    timeoutRef.current = setTimeout(() => setIsVisible(true), delayMs);
  }, [delayMs]);

  const hide = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    setIsVisible(false);
  }, []);

  return (
    <div
      className="relative inline-flex"
      onMouseEnter={show}
      onMouseLeave={hide}
      onFocus={show}
      onBlur={hide}
      aria-describedby={isVisible ? tooltipId : undefined}
    >
      {children}

      {isVisible && (
        <div
          id={tooltipId}
          role="tooltip"
          className={cn(
            'absolute z-50 whitespace-nowrap rounded px-2 py-1 text-xs font-medium shadow-md',
            'bg-tooltip text-tooltip-foreground',
            'pointer-events-none',
            TOOLTIP_POSITIONS[position]
          )}
        >
          {content}
          <span
            className={cn(
              'absolute block h-0 w-0 border-4',
              ARROW_POSITIONS[position]
            )}
            aria-hidden="true"
          />
        </div>
      )}
    </div>
  );
}
