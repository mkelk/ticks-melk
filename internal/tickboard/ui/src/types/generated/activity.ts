/* eslint-disable */
/**
 * AUTO-GENERATED FILE - DO NOT EDIT
 * Generated from: ../../../schemas/activity.schema.json
 * Run 'pnpm codegen' to regenerate.
 */

/**
 * Activity log entry tracking changes to ticks
 */
export interface Activity {
  /**
   * ISO timestamp when the activity occurred
   */
  ts: string;
  /**
   * ID of the tick this activity is about
   */
  tick: string;
  /**
   * Type of action (create, update, close, approve, reject, etc.)
   */
  action: string;
  /**
   * Who performed the action (user name or agent)
   */
  actor: string;
  /**
   * Parent epic ID if the tick belongs to an epic
   */
  epic?: string;
  /**
   * Additional action-specific data
   */
  data?: {
    [k: string]: unknown;
  };
  [k: string]: unknown;
}
