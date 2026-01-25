/**
 * Output stream types for live run output display.
 *
 * These types define the normalized format for run events used by the
 * run-output-pane component. The actual streaming is handled by CommsClient.
 */

// ============================================================================
// Types
// ============================================================================

/** Run event data (normalized format for display) */
export interface RunEvent {
  epicId: string;
  taskId?: string;
  source: 'ralph' | 'swarm-orchestrator' | 'swarm-subagent';
  eventType: 'task-started' | 'task-update' | 'tool-activity' | 'task-completed' | 'epic-started' | 'epic-completed' | 'context-generating' | 'context-generated' | 'context-loaded' | 'context-failed' | 'context-skipped' | 'connected';
  output?: string;
  status?: string;
  numTurns?: number;
  iteration?: number;
  success?: boolean;
  metrics?: {
    inputTokens: number;
    outputTokens: number;
    cacheReadTokens: number;
    cacheCreationTokens: number;
    costUsd: number;
    durationMs: number;
  };
  activeTool?: {
    name: string;
    input?: string;
    duration?: number;
  };
  message?: string;
  timestamp: string;
}
