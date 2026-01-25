/* eslint-disable */
/**
 * AUTO-GENERATED FILE - DO NOT EDIT
 * Generated from: ../../../schemas/api/requests.schema.json
 * Run 'pnpm codegen' to regenerate.
 */

/**
 * Request body types for ticks API endpoints
 */
export interface APIRequests {
  [k: string]: unknown;
}
/**
 * Request body for POST /api/ticks
 *
 * This interface was referenced by `APIRequests`'s JSON-Schema
 * via the `definition` "CreateTickRequest".
 */
export interface CreateTickRequest {
  /**
   * Title of the new tick
   */
  title: string;
  /**
   * Detailed description
   */
  description?: string;
  /**
   * Type of tick (defaults to task)
   */
  type?: 'bug' | 'feature' | 'task' | 'epic' | 'chore';
  /**
   * Priority level (defaults to 2)
   */
  priority?: number;
  /**
   * Parent epic ID
   */
  parent?: string;
  /**
   * Pre-declared gate for closing
   */
  requires?: 'approval' | 'review' | 'content';
  [k: string]: unknown;
}
/**
 * Request body for PATCH /api/ticks/:id
 *
 * This interface was referenced by `APIRequests`'s JSON-Schema
 * via the `definition` "UpdateTickRequest".
 */
export interface UpdateTickRequest {
  /**
   * New priority level
   */
  priority?: number;
  /**
   * New tick type
   */
  type?: 'bug' | 'feature' | 'task' | 'epic' | 'chore';
  /**
   * New parent epic ID
   */
  parent?: string;
  /**
   * New owner
   */
  owner?: string;
  /**
   * Pre-declared gate for closing
   */
  requires?: 'approval' | 'review' | 'content';
  [k: string]: unknown;
}
/**
 * Request body for POST /api/ticks/:id/reject
 *
 * This interface was referenced by `APIRequests`'s JSON-Schema
 * via the `definition` "RejectTickRequest".
 */
export interface RejectTickRequest {
  /**
   * Feedback explaining why the tick was rejected
   */
  feedback?: string;
  [k: string]: unknown;
}
/**
 * Request body for POST /api/ticks/:id/note
 *
 * This interface was referenced by `APIRequests`'s JSON-Schema
 * via the `definition` "AddNoteRequest".
 */
export interface AddNoteRequest {
  /**
   * Note content to add
   */
  message: string;
  [k: string]: unknown;
}
/**
 * Request body for POST /api/ticks/:id/close
 *
 * This interface was referenced by `APIRequests`'s JSON-Schema
 * via the `definition` "CloseTickRequest".
 */
export interface CloseTickRequest {
  /**
   * Reason for closing
   */
  reason?: string;
  [k: string]: unknown;
}
