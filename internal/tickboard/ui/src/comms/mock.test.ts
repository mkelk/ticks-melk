/**
 * Unit tests for MockCommsClient.
 * Tests all functionality without any external dependencies.
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { MockCommsClient } from './mock.js';
import { ReadOnlyError } from './client.js';
import type {
  TickEvent,
  RunEvent,
  ContextEvent,
  ConnectionEvent,
  CommsEvent,
} from './types.js';

describe('MockCommsClient', () => {
  let client: MockCommsClient;

  beforeEach(() => {
    client = new MockCommsClient();
  });

  // ===========================================================================
  // Lifecycle
  // ===========================================================================

  describe('lifecycle', () => {
    it('connect() sets connected to true', async () => {
      expect(client.isConnected()).toBe(false);
      await client.connect();
      expect(client.isConnected()).toBe(true);
    });

    it('connect() emits connection:connected event', async () => {
      const handler = vi.fn();
      client.onConnection(handler);

      await client.connect();

      expect(handler).toHaveBeenCalledWith({ type: 'connection:connected' });
    });

    it('disconnect() sets connected to false', async () => {
      await client.connect();
      expect(client.isConnected()).toBe(true);

      client.disconnect();
      expect(client.isConnected()).toBe(false);
    });

    it('disconnect() emits connection:disconnected event', async () => {
      await client.connect();
      const handler = vi.fn();
      client.onConnection(handler);

      client.disconnect();

      expect(handler).toHaveBeenCalledWith({ type: 'connection:disconnected' });
    });

    it('disconnect() clears run subscriptions', async () => {
      await client.connect();
      client.subscribeRun('epic-1');
      client.subscribeRun('epic-2');
      expect(client.getRunSubscriptions().size).toBe(2);

      client.disconnect();
      expect(client.getRunSubscriptions().size).toBe(0);
    });
  });

  // ===========================================================================
  // Event Subscriptions
  // ===========================================================================

  describe('event subscriptions', () => {
    it('onTick registers handler and returns unsubscribe', () => {
      const handler = vi.fn();
      const unsubscribe = client.onTick(handler);

      client.emitTick({ type: 'tick:updated', tick: createMockTick() });
      expect(handler).toHaveBeenCalledTimes(1);

      unsubscribe();
      client.emitTick({ type: 'tick:updated', tick: createMockTick() });
      expect(handler).toHaveBeenCalledTimes(1); // Still 1, not called again
    });

    it('onRun registers handler and returns unsubscribe', () => {
      const handler = vi.fn();
      const unsubscribe = client.onRun(handler);

      client.emitRun(createRunEvent());
      expect(handler).toHaveBeenCalledTimes(1);

      unsubscribe();
      client.emitRun(createRunEvent());
      expect(handler).toHaveBeenCalledTimes(1);
    });

    it('onContext registers handler and returns unsubscribe', () => {
      const handler = vi.fn();
      const unsubscribe = client.onContext(handler);

      client.emitContext({ type: 'context:generating', epicId: 'e1', taskCount: 5 });
      expect(handler).toHaveBeenCalledTimes(1);

      unsubscribe();
      client.emitContext({ type: 'context:loaded', epicId: 'e1' });
      expect(handler).toHaveBeenCalledTimes(1);
    });

    it('onConnection registers handler and returns unsubscribe', () => {
      const handler = vi.fn();
      const unsubscribe = client.onConnection(handler);

      client.emitConnection({ type: 'connection:connected' });
      expect(handler).toHaveBeenCalledTimes(1);

      unsubscribe();
      client.emitConnection({ type: 'connection:disconnected' });
      expect(handler).toHaveBeenCalledTimes(1);
    });

    it('multiple handlers can be registered for same event type', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();

      client.onTick(handler1);
      client.onTick(handler2);

      client.emitTick({ type: 'tick:updated', tick: createMockTick() });

      expect(handler1).toHaveBeenCalledTimes(1);
      expect(handler2).toHaveBeenCalledTimes(1);
    });
  });

  // ===========================================================================
  // Run Stream Subscriptions
  // ===========================================================================

  describe('run stream subscriptions', () => {
    it('subscribeRun adds epic to subscriptions', () => {
      expect(client.getRunSubscriptions().size).toBe(0);

      client.subscribeRun('epic-1');
      expect(client.getRunSubscriptions().has('epic-1')).toBe(true);
    });

    it('subscribeRun emits connection:connected with epicId', () => {
      const handler = vi.fn();
      client.onConnection(handler);

      client.subscribeRun('epic-1');

      expect(handler).toHaveBeenCalledWith({
        type: 'connection:connected',
        epicId: 'epic-1',
      });
    });

    it('unsubscribe removes epic from subscriptions', () => {
      const unsubscribe = client.subscribeRun('epic-1');
      expect(client.getRunSubscriptions().has('epic-1')).toBe(true);

      unsubscribe();
      expect(client.getRunSubscriptions().has('epic-1')).toBe(false);
    });

    it('multiple epics can be subscribed simultaneously', () => {
      client.subscribeRun('epic-1');
      client.subscribeRun('epic-2');
      client.subscribeRun('epic-3');

      const subs = client.getRunSubscriptions();
      expect(subs.size).toBe(3);
      expect(subs.has('epic-1')).toBe(true);
      expect(subs.has('epic-2')).toBe(true);
      expect(subs.has('epic-3')).toBe(true);
    });
  });

  // ===========================================================================
  // Event Emission
  // ===========================================================================

  describe('event emission', () => {
    it('emit dispatches tick events to tick handlers', () => {
      const tickHandler = vi.fn();
      const runHandler = vi.fn();
      client.onTick(tickHandler);
      client.onRun(runHandler);

      const event: TickEvent = { type: 'tick:updated', tick: createMockTick() };
      client.emit(event);

      expect(tickHandler).toHaveBeenCalledWith(event);
      expect(runHandler).not.toHaveBeenCalled();
    });

    it('emit dispatches run events to run handlers', () => {
      const tickHandler = vi.fn();
      const runHandler = vi.fn();
      client.onTick(tickHandler);
      client.onRun(runHandler);

      const event = createRunEvent();
      client.emit(event);

      expect(runHandler).toHaveBeenCalledWith(event);
      expect(tickHandler).not.toHaveBeenCalled();
    });

    it('emit dispatches context events to context handlers', () => {
      const contextHandler = vi.fn();
      const tickHandler = vi.fn();
      client.onContext(contextHandler);
      client.onTick(tickHandler);

      const event: ContextEvent = { type: 'context:generating', epicId: 'e1', taskCount: 3 };
      client.emit(event);

      expect(contextHandler).toHaveBeenCalledWith(event);
      expect(tickHandler).not.toHaveBeenCalled();
    });

    it('emit dispatches connection events to connection handlers', () => {
      const connectionHandler = vi.fn();
      const tickHandler = vi.fn();
      client.onConnection(connectionHandler);
      client.onTick(tickHandler);

      const event: ConnectionEvent = { type: 'connection:connected' };
      client.emit(event);

      expect(connectionHandler).toHaveBeenCalledWith(event);
      expect(tickHandler).not.toHaveBeenCalled();
    });

    it('emit dispatches activity:updated to tick handlers', () => {
      const tickHandler = vi.fn();
      client.onTick(tickHandler);

      client.emit({ type: 'activity:updated' });

      expect(tickHandler).toHaveBeenCalledWith({ type: 'activity:updated' });
    });

    it('emitConnection updates read-only state on local-status events', () => {
      expect(client.isReadOnly()).toBe(false);

      client.emitConnection({ type: 'connection:local-status', connected: false });
      expect(client.isReadOnly()).toBe(true);

      client.emitConnection({ type: 'connection:local-status', connected: true });
      expect(client.isReadOnly()).toBe(false);
    });
  });

  // ===========================================================================
  // Event Log
  // ===========================================================================

  describe('event log', () => {
    it('getEventLog returns all emitted events', () => {
      const tick = createMockTick();
      client.emitTick({ type: 'tick:updated', tick });
      client.emitConnection({ type: 'connection:connected' });
      client.emitRun(createRunEvent());

      const log = client.getEventLog();
      expect(log).toHaveLength(3);
      expect(log[0]).toEqual({ type: 'tick:updated', tick });
      expect(log[1]).toEqual({ type: 'connection:connected' });
      expect(log[2].type).toBe('run:task-started');
    });

    it('getEventLog returns a copy', () => {
      client.emitTick({ type: 'tick:updated', tick: createMockTick() });
      const log1 = client.getEventLog();
      const log2 = client.getEventLog();

      expect(log1).not.toBe(log2);
      expect(log1).toEqual(log2);
    });

    it('getEventsByType filters events', () => {
      client.emitTick({ type: 'tick:updated', tick: createMockTick() });
      client.emitTick({ type: 'tick:deleted', tickId: 't1' });
      client.emitConnection({ type: 'connection:connected' });

      const tickUpdates = client.getEventsByType('tick:updated');
      expect(tickUpdates).toHaveLength(1);
      expect(tickUpdates[0].type).toBe('tick:updated');

      const deletions = client.getEventsByType('tick:deleted');
      expect(deletions).toHaveLength(1);
      expect(deletions[0].tickId).toBe('t1');
    });

    it('clearEventLog clears all events', () => {
      client.emitTick({ type: 'tick:updated', tick: createMockTick() });
      client.emitConnection({ type: 'connection:connected' });
      expect(client.getEventLog()).toHaveLength(2);

      client.clearEventLog();
      expect(client.getEventLog()).toHaveLength(0);
    });
  });

  // ===========================================================================
  // Write Operations
  // ===========================================================================

  describe('write operations', () => {
    it('createTick returns mock tick', async () => {
      const result = await client.createTick({ title: 'Test Tick' });

      expect(result).toBeDefined();
      expect(result.title).toBe('Test Tick');
      expect(result.status).toBe('open');
    });

    it('createTick logs operation', async () => {
      await client.createTick({ title: 'Test Tick', priority: 1 });

      const writes = client.getWriteLog();
      expect(writes).toHaveLength(1);
      expect(writes[0].type).toBe('createTick');
      expect(writes[0].args.tick).toEqual({ title: 'Test Tick', priority: 1 });
    });

    it('updateTick returns mock tick with updates', async () => {
      const result = await client.updateTick('t1', { title: 'Updated', status: 'in_progress' });

      expect(result.id).toBe('t1');
      expect(result.title).toBe('Updated');
      expect(result.status).toBe('in_progress');
    });

    it('updateTick logs operation', async () => {
      await client.updateTick('t1', { status: 'closed' });

      const writes = client.getWriteLog();
      expect(writes).toHaveLength(1);
      expect(writes[0].type).toBe('updateTick');
      expect(writes[0].args).toEqual({ id: 't1', updates: { status: 'closed' } });
    });

    it('deleteTick logs operation', async () => {
      await client.deleteTick('t1');

      const writes = client.getWriteLog();
      expect(writes).toHaveLength(1);
      expect(writes[0].type).toBe('deleteTick');
      expect(writes[0].args).toEqual({ id: 't1' });
    });

    it('addNote logs operation', async () => {
      await client.addNote('t1', 'This is a note');

      const writes = client.getWriteLog();
      expect(writes).toHaveLength(1);
      expect(writes[0].type).toBe('addNote');
      expect(writes[0].args).toEqual({ id: 't1', message: 'This is a note' });
    });

    it('approveTick logs operation', async () => {
      await client.approveTick('t1');

      const writes = client.getWritesByType('approveTick');
      expect(writes).toHaveLength(1);
      expect(writes[0].args).toEqual({ id: 't1' });
    });

    it('rejectTick logs operation with reason', async () => {
      await client.rejectTick('t1', 'Not ready yet');

      const writes = client.getWritesByType('rejectTick');
      expect(writes).toHaveLength(1);
      expect(writes[0].args).toEqual({ id: 't1', reason: 'Not ready yet' });
    });

    it('closeTick logs operation with optional reason', async () => {
      await client.closeTick('t1', 'Done');

      const writes = client.getWritesByType('closeTick');
      expect(writes).toHaveLength(1);
      expect(writes[0].args).toEqual({ id: 't1', reason: 'Done' });
    });

    it('closeTick works without reason', async () => {
      await client.closeTick('t1');

      const writes = client.getWritesByType('closeTick');
      expect(writes).toHaveLength(1);
      expect(writes[0].args).toEqual({ id: 't1', reason: undefined });
    });

    it('reopenTick logs operation', async () => {
      await client.reopenTick('t1');

      const writes = client.getWritesByType('reopenTick');
      expect(writes).toHaveLength(1);
      expect(writes[0].args).toEqual({ id: 't1' });
    });
  });

  // ===========================================================================
  // Write Log
  // ===========================================================================

  describe('write log', () => {
    it('getWriteLog returns all operations', async () => {
      await client.createTick({ title: 'T1' });
      await client.updateTick('t1', { status: 'closed' });
      await client.deleteTick('t2');

      const log = client.getWriteLog();
      expect(log).toHaveLength(3);
      expect(log.map((w) => w.type)).toEqual(['createTick', 'updateTick', 'deleteTick']);
    });

    it('getWriteLog returns a copy', async () => {
      await client.createTick({ title: 'T1' });

      const log1 = client.getWriteLog();
      const log2 = client.getWriteLog();

      expect(log1).not.toBe(log2);
      expect(log1).toEqual(log2);
    });

    it('getWritesByType filters operations', async () => {
      await client.createTick({ title: 'T1' });
      await client.createTick({ title: 'T2' });
      await client.updateTick('t1', {});
      await client.deleteTick('t1');

      expect(client.getWritesByType('createTick')).toHaveLength(2);
      expect(client.getWritesByType('updateTick')).toHaveLength(1);
      expect(client.getWritesByType('deleteTick')).toHaveLength(1);
      expect(client.getWritesByType('addNote')).toHaveLength(0);
    });

    it('clearWriteLog clears all operations', async () => {
      await client.createTick({ title: 'T1' });
      await client.updateTick('t1', {});
      expect(client.getWriteLog()).toHaveLength(2);

      client.clearWriteLog();
      expect(client.getWriteLog()).toHaveLength(0);
    });

    it('write operations include timestamp', async () => {
      const before = new Date();
      await client.createTick({ title: 'T1' });
      const after = new Date();

      const log = client.getWriteLog();
      expect(log[0].timestamp).toBeInstanceOf(Date);
      expect(log[0].timestamp.getTime()).toBeGreaterThanOrEqual(before.getTime());
      expect(log[0].timestamp.getTime()).toBeLessThanOrEqual(after.getTime());
    });
  });

  // ===========================================================================
  // Write Responses
  // ===========================================================================

  describe('write responses', () => {
    it('setWriteResponse configures custom result', async () => {
      const customTick = createMockTick({ id: 'custom-1', title: 'Custom' });
      client.setWriteResponse('createTick', { result: customTick });

      const result = await client.createTick({ title: 'Ignored' });
      expect(result).toEqual(customTick);
    });

    it('setWriteResponse configures error to throw', async () => {
      const customError = new Error('Custom error');
      client.setWriteResponse('updateTick', { error: customError });

      await expect(client.updateTick('t1', {})).rejects.toThrow('Custom error');
    });

    it('setWriteResponse delay works', async () => {
      client.setWriteResponse('createTick', { delay: 50 });

      const start = Date.now();
      await client.createTick({ title: 'Delayed' });
      const elapsed = Date.now() - start;

      expect(elapsed).toBeGreaterThanOrEqual(45); // Allow small timing variance
    });

    it('clearWriteResponse removes configuration', async () => {
      const customTick = createMockTick({ id: 'custom-1' });
      client.setWriteResponse('createTick', { result: customTick });

      let result = await client.createTick({ title: 'T1' });
      expect(result.id).toBe('custom-1');

      client.clearWriteResponse('createTick');

      result = await client.createTick({ title: 'T2' });
      expect(result.id).not.toBe('custom-1');
    });

    it('write response error still logs operation', async () => {
      client.setWriteResponse('createTick', { error: new Error('Fail') });

      try {
        await client.createTick({ title: 'T1' });
      } catch {
        // Expected
      }

      expect(client.getWriteLog()).toHaveLength(1);
    });
  });

  // ===========================================================================
  // Failure Modes
  // ===========================================================================

  describe('failure modes', () => {
    it('failNextWrite causes next write to fail', async () => {
      client.failNextWrite(new Error('One-time failure'));

      await expect(client.createTick({ title: 'T1' })).rejects.toThrow('One-time failure');
    });

    it('failNextWrite is one-time only', async () => {
      client.failNextWrite(new Error('One-time failure'));

      await expect(client.createTick({ title: 'T1' })).rejects.toThrow('One-time failure');
      await expect(client.createTick({ title: 'T2' })).resolves.toBeDefined();
    });

    it('failNextWrite still logs operation', async () => {
      client.failNextWrite(new Error('Fail'));

      try {
        await client.createTick({ title: 'T1' });
      } catch {
        // Expected
      }

      expect(client.getWriteLog()).toHaveLength(1);
    });

    it('read-only mode causes ReadOnlyError', async () => {
      client.setReadOnly(true);

      await expect(client.createTick({ title: 'T1' })).rejects.toThrow(ReadOnlyError);
      await expect(client.updateTick('t1', {})).rejects.toThrow(ReadOnlyError);
      await expect(client.deleteTick('t1')).rejects.toThrow(ReadOnlyError);
      await expect(client.addNote('t1', 'note')).rejects.toThrow(ReadOnlyError);
      await expect(client.approveTick('t1')).rejects.toThrow(ReadOnlyError);
      await expect(client.rejectTick('t1', 'reason')).rejects.toThrow(ReadOnlyError);
      await expect(client.closeTick('t1')).rejects.toThrow(ReadOnlyError);
      await expect(client.reopenTick('t1')).rejects.toThrow(ReadOnlyError);
    });

    it('read-only mode does not log failed operations', async () => {
      client.setReadOnly(true);

      try {
        await client.createTick({ title: 'T1' });
      } catch {
        // Expected
      }

      expect(client.getWriteLog()).toHaveLength(0);
    });
  });

  // ===========================================================================
  // State Management
  // ===========================================================================

  describe('state management', () => {
    it('isConnected returns connection state', async () => {
      expect(client.isConnected()).toBe(false);
      await client.connect();
      expect(client.isConnected()).toBe(true);
      client.disconnect();
      expect(client.isConnected()).toBe(false);
    });

    it('isReadOnly returns read-only state', () => {
      expect(client.isReadOnly()).toBe(false);
      client.setReadOnly(true);
      expect(client.isReadOnly()).toBe(true);
      client.setReadOnly(false);
      expect(client.isReadOnly()).toBe(false);
    });

    it('setReadOnly emits connection:local-status event', () => {
      const handler = vi.fn();
      client.onConnection(handler);

      client.setReadOnly(true);
      expect(handler).toHaveBeenCalledWith({
        type: 'connection:local-status',
        connected: false,
      });

      client.setReadOnly(false);
      expect(handler).toHaveBeenCalledWith({
        type: 'connection:local-status',
        connected: true,
      });
    });

    it('setConnected changes state without emitting', () => {
      const handler = vi.fn();
      client.onConnection(handler);

      client.setConnected(true);
      expect(client.isConnected()).toBe(true);
      expect(handler).not.toHaveBeenCalled();
    });

    it('getConnectionInfo returns connection details', async () => {
      const info = client.getConnectionInfo();

      expect(info.mode).toBe('local');
      expect(info.connected).toBe(false);
      expect(info.baseUrl).toBe('mock://localhost');

      await client.connect();
      expect(client.getConnectionInfo().connected).toBe(true);
    });

    it('getRunSubscriptions returns copy of subscriptions', () => {
      client.subscribeRun('epic-1');
      client.subscribeRun('epic-2');

      const subs1 = client.getRunSubscriptions();
      const subs2 = client.getRunSubscriptions();

      expect(subs1).not.toBe(subs2);
      expect(subs1).toEqual(subs2);
    });

    it('reset clears all state', async () => {
      // Set up various state
      await client.connect();
      client.subscribeRun('epic-1');
      client.emitTick({ type: 'tick:updated', tick: createMockTick() });
      await client.createTick({ title: 'T1' });
      client.setWriteResponse('createTick', { result: createMockTick() });
      client.failNextWrite(new Error('Fail'));

      const tickHandler = vi.fn();
      client.onTick(tickHandler);

      // Reset
      client.reset();

      // Verify all state is cleared
      expect(client.isConnected()).toBe(false);
      expect(client.isReadOnly()).toBe(false);
      expect(client.getRunSubscriptions().size).toBe(0);
      expect(client.getEventLog()).toHaveLength(0);
      expect(client.getWriteLog()).toHaveLength(0);

      // Handlers should be cleared
      client.emitTick({ type: 'tick:updated', tick: createMockTick() });
      expect(tickHandler).not.toHaveBeenCalled();

      // Write responses should be cleared (returns default mock)
      const result = await client.createTick({ title: 'After Reset' });
      expect(result.title).toBe('After Reset');

      // failNextWrite should be cleared
      await expect(client.createTick({ title: 'T2' })).resolves.toBeDefined();
    });
  });
});

// =============================================================================
// Test Helpers
// =============================================================================

function createMockTick(overrides: Partial<{
  id: string;
  title: string;
  status: string;
  priority: number;
  type: string;
}> = {}) {
  return {
    id: overrides.id || 'test-1',
    title: overrides.title || 'Test Tick',
    status: overrides.status || 'open',
    priority: overrides.priority || 2,
    type: overrides.type || 'task',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    created_by: 'test@user.com',
    owner: '',
  };
}

function createRunEvent(): RunEvent {
  return {
    type: 'run:task-started',
    taskId: 'task-1',
    epicId: 'epic-1',
    status: 'running',
    numTurns: 0,
    timestamp: new Date().toISOString(),
  };
}
