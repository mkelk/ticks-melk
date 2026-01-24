/**
 * Auth module for user management and token handling
 */

import type { Env } from "./index";

// Retry wrapper for D1 operations to handle transient failures
async function withRetry<T>(
  operation: () => Promise<T>,
  maxRetries: number = 3,
  baseDelayMs: number = 100
): Promise<T> {
  let lastError: Error | null = null;
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await operation();
    } catch (err) {
      lastError = err instanceof Error ? err : new Error(String(err));
      // Only retry on timeout/transient errors
      if (!lastError.message.includes("timeout") &&
          !lastError.message.includes("D1_ERROR") &&
          !lastError.message.includes("reset")) {
        throw lastError;
      }
      // Exponential backoff
      if (attempt < maxRetries - 1) {
        await new Promise(r => setTimeout(r, baseDelayMs * Math.pow(2, attempt)));
      }
    }
  }
  throw lastError;
}

// Simple password hashing using Web Crypto API
async function hashPassword(password: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(password);
  const hashBuffer = await crypto.subtle.digest("SHA-256", data);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map((b) => b.toString(16).padStart(2, "0")).join("");
}

// Generate a random token
function generateToken(): string {
  const bytes = new Uint8Array(32);
  crypto.getRandomValues(bytes);
  return Array.from(bytes)
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

// Generate a short ID
function generateId(): string {
  const bytes = new Uint8Array(8);
  crypto.getRandomValues(bytes);
  return Array.from(bytes)
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

interface User {
  id: string;
  email: string;
  password_hash: string;
  created_at: number;
  updated_at: number;
}

interface Token {
  id: string;
  user_id: string;
  name: string;
  token_hash: string;
  last_used_at: number | null;
  created_at: number;
}

interface Board {
  id: string;
  user_id: string;
  name: string;
  machine_id: string | null;
  last_seen_at: number | null;
  created_at: number;
}

// Signup: create new user
export async function signup(
  env: Env,
  email: string,
  password: string
): Promise<Response> {
  if (!email || !password) {
    return Response.json({ error: "Email and password required" }, { status: 400 });
  }

  if (password.length < 8) {
    return Response.json({ error: "Password must be at least 8 characters" }, { status: 400 });
  }

  const passwordHash = await hashPassword(password);
  const userId = generateId();

  try {
    await withRetry(() =>
      env.DB.prepare(
        "INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)"
      )
        .bind(userId, email.toLowerCase(), passwordHash)
        .run()
    );

    return Response.json({ id: userId, email: email.toLowerCase() }, { status: 201 });
  } catch (err: unknown) {
    if (err instanceof Error && err.message.includes("UNIQUE constraint failed")) {
      return Response.json({ error: "Email already registered" }, { status: 409 });
    }
    throw err;
  }
}

// Timing data structure for Server-Timing header
interface TimingData {
  [key: string]: number;
}

function buildServerTimingHeader(timings: TimingData): string {
  return Object.entries(timings)
    .map(([name, dur]) => `${name};dur=${dur}`)
    .join(", ");
}

// Login: authenticate user and return session token
export async function login(
  env: Env,
  email: string,
  password: string
): Promise<Response> {
  const t0 = Date.now();
  const timings: TimingData = {};

  try {
    if (!email || !password) {
      return Response.json({ error: "Email and password required" }, { status: 400 });
    }

    const tHash1 = Date.now();
    const passwordHash = await hashPassword(password);
    timings["hash-password"] = Date.now() - tHash1;
    console.log(`login: hash took ${timings["hash-password"]}ms`);

    const t1 = Date.now();
    const result = await env.DB.prepare(
      "SELECT id, email FROM users WHERE email = ? AND password_hash = ?"
    )
      .bind(email.toLowerCase(), passwordHash)
      .first<User>();
    timings["d1-select-user"] = Date.now() - t1;
    console.log(`login: SELECT user took ${timings["d1-select-user"]}ms`);

    if (!result) {
      timings["total"] = Date.now() - t0;
      const response = Response.json({ error: "Invalid email or password" }, { status: 401 });
      const headers = new Headers(response.headers);
      headers.set("Server-Timing", buildServerTimingHeader(timings));
      return new Response(response.body, { status: 401, headers });
    }

    // Create a new session token first, then delete old ones
    const tGen = Date.now();
    const token = generateToken();
    timings["gen-token"] = Date.now() - tGen;

    const tHash2 = Date.now();
    const tokenHash = await hashPassword(token);
    timings["hash-token"] = Date.now() - tHash2;

    const tokenId = generateId();
    const sessionName = `session-${Date.now()}`;

    const t3 = Date.now();
    await env.DB.prepare(
      "INSERT INTO tokens (id, user_id, name, token_hash) VALUES (?, ?, ?, ?)"
    )
      .bind(tokenId, result.id, sessionName, tokenHash)
      .run();
    timings["d1-insert-token"] = Date.now() - t3;
    console.log(`login: INSERT token took ${timings["d1-insert-token"]}ms`);

    // Delete old session tokens AFTER inserting new one, excluding the one we just created
    env.DB.prepare(
      "DELETE FROM tokens WHERE user_id = ? AND name LIKE 'session%' AND id != ?"
    )
      .bind(result.id, tokenId)
      .run()
      .catch(() => {});

    timings["total"] = Date.now() - t0;
    console.log(`login: total ${timings["total"]}ms | breakdown: ${JSON.stringify(timings)}`);

    const response = Response.json({
      user: { id: result.id, email: result.email },
      token: token,
      _timing: timings, // Include in response body for client visibility
    });

    // Also set session cookie for browser-based access
    const headers = new Headers(response.headers);
    headers.set("Set-Cookie", createSessionCookie(token));
    headers.set("Server-Timing", buildServerTimingHeader(timings));

    return new Response(response.body, {
      status: response.status,
      headers,
    });
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : String(err);
    timings["total"] = Date.now() - t0;
    console.error(`Login error after ${timings["total"]}ms:`, message);
    return Response.json({ error: "Login failed. Please try again." }, { status: 500 });
  }
}

// Create API token for local agents
export async function createToken(
  env: Env,
  userId: string,
  name: string
): Promise<Response> {
  if (!name) {
    return Response.json({ error: "Token name required" }, { status: 400 });
  }

  const token = generateToken();
  const tokenHash = await hashPassword(token);
  const tokenId = generateId();

  try {
    await withRetry(() =>
      env.DB.prepare(
        "INSERT INTO tokens (id, user_id, name, token_hash) VALUES (?, ?, ?, ?)"
      )
        .bind(tokenId, userId, name, tokenHash)
        .run()
    );

    // Return the full token only once - user must save it
    return Response.json({ id: tokenId, name, token }, { status: 201 });
  } catch (err: unknown) {
    if (err instanceof Error && err.message.includes("UNIQUE constraint failed")) {
      return Response.json({ error: "Token name already exists" }, { status: 409 });
    }
    throw err;
  }
}

// List user's tokens (active only - revoked are deleted)
export async function listTokens(env: Env, userId: string): Promise<Response> {
  try {
    const result = await withRetry(() =>
      env.DB.prepare(
        "SELECT id, name, last_used_at, created_at FROM tokens WHERE user_id = ? ORDER BY created_at DESC"
      )
        .bind(userId)
        .all<Token>()
    );

    return Response.json({
      tokens: result.results.map((t) => ({
        id: t.id,
        name: t.name,
        lastUsedAt: t.last_used_at,
        createdAt: t.created_at,
        revoked: false, // Active tokens only
      })),
    });
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : String(err);
    console.error("List tokens error:", message);
    return Response.json({ tokens: [] }); // Return empty on error
  }
}

// Revoke (delete) a token
export async function revokeToken(
  env: Env,
  userId: string,
  tokenId: string
): Promise<Response> {
  try {
    const result = await withRetry(() =>
      env.DB.prepare(
        "DELETE FROM tokens WHERE id = ? AND user_id = ?"
      )
        .bind(tokenId, userId)
        .run()
    );

    if (result.meta.changes === 0) {
      return Response.json({ error: "Token not found" }, { status: 404 });
    }

    return Response.json({ success: true });
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : String(err);
    console.error("Revoke token error:", message);
    return Response.json({ error: "Failed to revoke token. Please try again." }, { status: 500 });
  }
}

// Token expiry duration (1 hour from validation)
const TOKEN_EXPIRY_MS = 60 * 60 * 1000; // 1 hour

// Validate a token and return user info
export async function validateToken(
  env: Env,
  token: string
): Promise<{ userId: string; tokenId: string; expiresAt: number; _timing?: TimingData } | null> {
  const t0 = Date.now();
  const timings: TimingData = {};
  try {
    const tHash = Date.now();
    const tokenHash = await hashPassword(token);
    timings["hash"] = Date.now() - tHash;

    const tSelect = Date.now();
    const result = await env.DB.prepare(
      "SELECT t.id, t.user_id FROM tokens t WHERE t.token_hash = ?"
    )
      .bind(tokenHash)
      .first<{ id: string; user_id: string }>();
    timings["d1-select"] = Date.now() - tSelect;
    timings["total"] = Date.now() - t0;
    console.log(`validateToken: total ${timings["total"]}ms | hash=${timings["hash"]}ms d1=${timings["d1-select"]}ms`);

    if (!result) {
      return null;
    }

    // Update last used (fire and forget - don't block on this)
    env.DB.prepare("UPDATE tokens SET last_used_at = unixepoch() WHERE id = ?")
      .bind(result.id)
      .run()
      .catch(() => {}); // Ignore errors

    // Calculate token expiry (1 hour from now)
    const expiresAt = Date.now() + TOKEN_EXPIRY_MS;

    return { userId: result.user_id, tokenId: result.id, expiresAt, _timing: timings };
  } catch (err) {
    timings["total"] = Date.now() - t0;
    console.error(`Validate token error after ${timings["total"]}ms:`, err);
    return null;
  }
}

// Register or update a board
export async function registerBoard(
  env: Env,
  userId: string,
  boardName: string,
  machineId: string
): Promise<string> {
  const boardId = generateId();

  try {
    // Upsert board
    await withRetry(() =>
      env.DB.prepare(
        `INSERT INTO boards (id, user_id, name, machine_id, last_seen_at)
         VALUES (?, ?, ?, ?, unixepoch())
         ON CONFLICT(user_id, name) DO UPDATE SET
           machine_id = excluded.machine_id,
           last_seen_at = unixepoch()`
      )
        .bind(boardId, userId, boardName, machineId)
        .run()
    );

    // Get the actual ID (in case of upsert)
    const result = await withRetry(() =>
      env.DB.prepare(
        "SELECT id FROM boards WHERE user_id = ? AND name = ?"
      )
        .bind(userId, boardName)
        .first<{ id: string }>()
    );

    return result?.id || boardId;
  } catch (err) {
    console.error("Register board error:", err);
    return boardId;
  }
}

// List user's boards with online status
export async function listBoards(
  env: Env,
  userId: string,
  onlineBoards: Set<string>
): Promise<Response> {
  try {
    const result = await withRetry(() =>
      env.DB.prepare(
        "SELECT id, name, machine_id, last_seen_at, created_at FROM boards WHERE user_id = ? ORDER BY name"
      )
        .bind(userId)
        .all<Board>()
    );

    return Response.json({
      boards: result.results.map((b) => ({
        id: b.id,
        name: b.name,
        machineId: b.machine_id,
        lastSeenAt: b.last_seen_at,
        createdAt: b.created_at,
        online: onlineBoards.has(b.name),
      })),
    });
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : String(err);
    console.error("List boards error:", message);
    return Response.json({ boards: [] }); // Return empty on error
  }
}

// Get user ID from request (via Authorization header or session cookie)
export async function getUserFromRequest(
  env: Env,
  request: Request
): Promise<{ userId: string; tokenId: string } | null> {
  // Try Authorization header first
  const authHeader = request.headers.get("Authorization");
  if (authHeader?.startsWith("Bearer ")) {
    const token = authHeader.slice(7);
    return validateToken(env, token);
  }

  // Try session cookie
  const cookie = request.headers.get("Cookie");
  if (cookie) {
    const sessionMatch = cookie.match(/session=([^;]+)/);
    if (sessionMatch) {
      return validateToken(env, sessionMatch[1]);
    }
  }

  return null;
}

// Check if user owns a specific board
export async function userOwnsBoard(
  env: Env,
  userId: string,
  boardName: string
): Promise<boolean> {
  try {
    const result = await withRetry(() =>
      env.DB.prepare(
        "SELECT 1 FROM boards WHERE user_id = ? AND name = ?"
      )
        .bind(userId, boardName)
        .first()
    );

    return result !== null;
  } catch (err) {
    console.error("Check board ownership error:", err);
    return false;
  }
}

// Check if user owns a specific project (by project ID / board name)
export async function userOwnsProject(
  env: Env,
  userId: string,
  projectId: string
): Promise<boolean> {
  try {
    const result = await withRetry(() =>
      env.DB.prepare(
        'SELECT 1 FROM boards WHERE user_id = ? AND name = ?'
      )
        .bind(userId, projectId)
        .first()
    );
    return result !== null;
  } catch (err) {
    console.error('Check project ownership error:', err);
    return false; // Fail closed
  }
}

// Delete a board from the user's dashboard
export async function deleteBoard(
  env: Env,
  userId: string,
  boardId: string
): Promise<Response> {
  try {
    const result = await withRetry(() =>
      env.DB.prepare(
        "DELETE FROM boards WHERE id = ? AND user_id = ?"
      )
        .bind(boardId, userId)
        .run()
    );

    if (result.meta.changes === 0) {
      return Response.json({ error: "Board not found" }, { status: 404 });
    }

    return Response.json({ success: true });
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : String(err);
    console.error("Delete board error:", message);
    return Response.json({ error: "Failed to delete board. Please try again." }, { status: 500 });
  }
}

// Create session cookie header
export function createSessionCookie(token: string, maxAge = 30 * 24 * 60 * 60): string {
  return `session=${token}; HttpOnly; Secure; SameSite=Strict; Max-Age=${maxAge}; Path=/`;
}

// Clear session cookie header
export function clearSessionCookie(): string {
  return "session=; HttpOnly; Secure; SameSite=Strict; Max-Age=0; Path=/";
}
