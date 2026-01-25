/**
 * Test setup file for comms tests.
 * Sets up EventSource polyfill for Node.js environment.
 */

import { EventSource } from 'eventsource';
import { vi } from 'vitest';

// Polyfill EventSource for Node.js/happy-dom environment
// This must run before any test files are loaded
// Use vi.stubGlobal to ensure it's available in all contexts
vi.stubGlobal('EventSource', EventSource);

// Also set on globalThis directly as backup
// eslint-disable-next-line @typescript-eslint/no-explicit-any
(globalThis as any).EventSource = EventSource;
