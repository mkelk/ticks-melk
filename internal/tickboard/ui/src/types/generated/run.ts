/* eslint-disable */
/**
 * AUTO-GENERATED FILE - DO NOT EDIT
 * Generated from: ../../../schemas/run.schema.json
 * Run 'pnpm codegen' to regenerate.
 */

/**
 * Current state of an in-progress agent run
 *
 * This interface was referenced by `RunRecords`'s JSON-Schema
 * via the `definition` "RunStatus".
 */
export type RunStatus = 'starting' | 'thinking' | 'writing' | 'tool_use' | 'complete' | 'error';

/**
 * Types for agent run execution records
 */
export interface RunRecords {
  [k: string]: unknown;
}
/**
 * Token and cost metrics for an agent run
 *
 * This interface was referenced by `RunRecords`'s JSON-Schema
 * via the `definition` "MetricsRecord".
 */
export interface MetricsRecord {
  /**
   * Number of input tokens consumed
   */
  input_tokens: number;
  /**
   * Number of output tokens generated
   */
  output_tokens: number;
  /**
   * Number of tokens read from cache
   */
  cache_read_tokens: number;
  /**
   * Number of tokens written to cache
   */
  cache_creation_tokens: number;
  /**
   * Total cost in USD
   */
  cost_usd: number;
  /**
   * Total duration in milliseconds
   */
  duration_ms: number;
  [k: string]: unknown;
}
/**
 * Record of a single tool invocation
 *
 * This interface was referenced by `RunRecords`'s JSON-Schema
 * via the `definition` "ToolRecord".
 */
export interface ToolRecord {
  /**
   * Name of the tool that was invoked
   */
  name: string;
  /**
   * Tool input (may be truncated)
   */
  input?: string;
  /**
   * Tool output (may be truncated)
   */
  output?: string;
  /**
   * Tool execution duration in milliseconds
   */
  duration_ms: number;
  /**
   * Whether the tool invocation resulted in an error
   */
  is_error?: boolean;
  [k: string]: unknown;
}
/**
 * Result from a single verifier
 *
 * This interface was referenced by `RunRecords`'s JSON-Schema
 * via the `definition` "VerifierResult".
 */
export interface VerifierResult {
  /**
   * Name of the verifier (e.g., git, test)
   */
  verifier: string;
  /**
   * Whether this verifier passed
   */
  passed: boolean;
  /**
   * Verifier output (may be truncated)
   */
  output?: string;
  /**
   * Verifier execution duration in milliseconds
   */
  duration_ms: number;
  /**
   * Error message if verification failed due to an error
   */
  error?: string;
  [k: string]: unknown;
}
/**
 * Aggregated verification results for a run
 *
 * This interface was referenced by `RunRecords`'s JSON-Schema
 * via the `definition` "VerificationRecord".
 */
export interface VerificationRecord {
  /**
   * Whether all verifiers passed
   */
  all_passed: boolean;
  /**
   * Individual verifier results
   */
  results?: VerifierResult[];
  [k: string]: unknown;
}
/**
 * Complete record of a finished agent run
 *
 * This interface was referenced by `RunRecords`'s JSON-Schema
 * via the `definition` "RunRecord".
 */
export interface RunRecord {
  /**
   * Unique session identifier
   */
  session_id: string;
  /**
   * Model used for the run (e.g., claude-sonnet-4-20250514)
   */
  model: string;
  /**
   * ISO timestamp when the run started
   */
  started_at: string;
  /**
   * ISO timestamp when the run ended
   */
  ended_at: string;
  /**
   * Final output text from the agent
   */
  output: string;
  /**
   * Thinking/reasoning content (if extended thinking was used)
   */
  thinking?: string;
  /**
   * List of tool invocations during the run
   */
  tools?: ToolRecord[];
  metrics: MetricsRecord1;
  /**
   * Whether the run completed successfully
   */
  success: boolean;
  /**
   * Number of API round-trips
   */
  num_turns: number;
  /**
   * Error message if the run failed
   */
  error_msg?: string;
  verification?: VerificationRecord1;
  [k: string]: unknown;
}
/**
 * Token and cost metrics
 */
export interface MetricsRecord1 {
  /**
   * Number of input tokens consumed
   */
  input_tokens: number;
  /**
   * Number of output tokens generated
   */
  output_tokens: number;
  /**
   * Number of tokens read from cache
   */
  cache_read_tokens: number;
  /**
   * Number of tokens written to cache
   */
  cache_creation_tokens: number;
  /**
   * Total cost in USD
   */
  cost_usd: number;
  /**
   * Total duration in milliseconds
   */
  duration_ms: number;
  [k: string]: unknown;
}
/**
 * Verification results (if verification was run)
 */
export interface VerificationRecord1 {
  /**
   * Whether all verifiers passed
   */
  all_passed: boolean;
  /**
   * Individual verifier results
   */
  results?: VerifierResult[];
  [k: string]: unknown;
}
