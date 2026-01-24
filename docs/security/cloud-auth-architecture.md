# Cloud Authentication Architecture

This document describes the security architecture for the Ticks cloud service, specifically the authentication and authorization patterns used to ensure user data isolation in Durable Objects.

## Overview

The cloud architecture consists of three layers:

```
┌─────────────────────────────────────────────────────────────┐
│                     Cloudflare Edge                          │
├─────────────────────────────────────────────────────────────┤
│  Main Worker (index.ts)                                      │
│  - Authentication (token validation)                         │
│  - Authorization (project ownership)                         │
│  - Routes requests to appropriate Durable Objects            │
├─────────────────────────────────────────────────────────────┤
│  Durable Objects                                             │
│  ┌─────────────────┐  ┌─────────────────────────────────┐   │
│  │   AgentHub      │  │   ProjectRoom (per-project)     │   │
│  │   (singleton)   │  │   - Tick state storage          │   │
│  │   - Agent relay │  │   - Real-time sync              │   │
│  └─────────────────┘  └─────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│  D1 Database                                                 │
│  - users, tokens, boards tables                              │
└─────────────────────────────────────────────────────────────┘
```

## Authentication Flow

### 1. Token Validation

All protected endpoints require authentication via Bearer token:

```
Client Request
     │
     ▼
┌─────────────────────────────────────┐
│  Main Worker                         │
│  1. Extract token from:              │
│     - Authorization: Bearer <token>  │
│     - Sec-WebSocket-Protocol header  │
│     - Cookie (session)               │
│  2. Hash token (SHA-256)             │
│  3. Query D1: tokens WHERE hash = ?  │
│  4. Return userId if valid           │
└─────────────────────────────────────┘
     │
     ▼
  Proceed or 401 Unauthorized
```

### 2. Project Authorization

After authentication, project ownership is verified:

```typescript
// auth.ts
export async function userOwnsProject(
  env: Env,
  userId: string,
  projectId: string
): Promise<boolean> {
  try {
    const result = await env.DB.prepare(
      'SELECT 1 FROM boards WHERE user_id = ? AND name = ?'
    ).bind(userId, projectId).first();
    return result !== null;
  } catch (err) {
    return false; // Fail closed
  }
}
```

**Key principle:** Returns `false` on any error (fail closed).

### 3. Validated Claims Passing

Once validated, user identity is passed to DOs via trusted headers:

```typescript
// Worker sets these headers - never trust from external requests
headers.set("X-Validated-User-Id", tokenInfo.userId);
headers.set("X-Validated-Project-Id", projectId);
```

## Security Headers

| Header | Purpose | Set By |
|--------|---------|--------|
| `X-Validated-User-Id` | Authenticated user ID | Main Worker only |
| `X-Validated-Project-Id` | Verified project the user owns | Main Worker only |

**Important:** These headers are set by the worker after validation. External requests attempting to set these headers have no effect - the worker overwrites them.

## Defense in Depth

The ProjectRoom DO performs secondary validation:

```typescript
// project-room.ts
async fetch(request: Request): Promise<Response> {
  const validatedUserId = request.headers.get("X-Validated-User-Id");

  if (request.headers.get("Upgrade") === "websocket") {
    if (!validatedUserId) {
      console.error("WebSocket request missing X-Validated-User-Id header");
      return new Response("Unauthorized - missing validation", { status: 401 });
    }
  }
  // ...
}
```

This protects against:
- Bugs in worker routing that bypass auth
- Future code changes that accidentally remove auth
- Direct DO access (if somehow exposed)

## Token Transmission

### WebSocket Connections

Tokens are transmitted via the `Sec-WebSocket-Protocol` header, not URL query parameters:

```typescript
// Client
const ws = new WebSocket(url, ['ticks-v1', `token-${token}`]);

// Server extracts from header
const protocols = request.headers.get("Sec-WebSocket-Protocol") || "";
const tokenMatch = protocols.match(/token-([^,\s]+)/);
```

**Why not query params?**
- Query params are logged in browser history
- Captured by proxy/CDN logs
- Visible in referrer headers

### HTTP Requests

Standard `Authorization: Bearer <token>` header or session cookie.

## Authorization Model

```
┌──────────┐      owns       ┌─────────┐
│   User   │ ──────────────▶ │  Board  │
└──────────┘                 └─────────┘
     │                            │
     │ has                        │ = projectId
     ▼                            ▼
┌──────────┐               ┌─────────────┐
│  Tokens  │               │ ProjectRoom │
└──────────┘               │   (DO)      │
                           └─────────────┘
```

- Users own boards (projects)
- Board `name` = Project ID (format: `owner/repo`)
- Tokens belong to users
- All project access requires ownership verification

## Protected Endpoints

### Require Authentication + Ownership

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/projects/:project/sync` | WebSocket | Real-time tick sync |
| `/api/projects/:project/ticks/:id/note` | POST | Add note to tick |
| `/api/projects/:project/ticks/:id/approve` | POST | Approve tick |
| `/api/projects/:project/ticks/:id/reject` | POST | Reject tick |
| `/api/projects/:project/ticks/:id/close` | POST | Close tick |
| `/api/projects/:project/ticks/:id/reopen` | POST | Reopen tick |

### Require Authentication Only

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/tokens` | GET/POST | List/create tokens |
| `/api/tokens/:id` | DELETE | Revoke token |
| `/api/boards` | GET | List user's boards |
| `/events/:board` | GET (SSE) | Board events |

### Public Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/health` | GET | Health check |
| `/api/auth/signup` | POST | User registration |
| `/api/auth/login` | POST | User login |

## Threat Model

| Threat | Mitigation |
|--------|------------|
| Cross-user data access | Ownership verification on every request |
| Token theft | HTTPS only, no URL exposure, session expiry |
| Token brute force | SHA-256 hashing, rate limiting (Cloudflare) |
| Header injection | Worker overwrites untrusted headers |
| DO direct access | CF routing prevents; header validation as backup |
| Timing attacks | Constant-time token comparison (via hash) |
| Session fixation | New token on each login, old sessions deleted |

## Session Management

- Session tokens expire after 30 days
- Login creates new token, deletes old sessions
- API tokens persist until manually revoked
- Long-lived WebSocket connections check token expiry periodically

## Code References

| File | Purpose |
|------|---------|
| `cloud/worker/src/auth.ts` | Authentication functions, token validation |
| `cloud/worker/src/index.ts` | Route handlers, authorization checks |
| `cloud/worker/src/project-room.ts` | ProjectRoom DO, defense-in-depth |

## Incident Response

If an auth bypass is discovered:

1. **Immediate:** Deploy hotfix via `wrangler deploy`
2. **Assess:** Review logs for exploitation (`wrangler tail`)
3. **Notify:** Contact affected users if data was accessed
4. **Post-mortem:** Document and improve

## Changelog

| Date | Change |
|------|--------|
| 2026-01-24 | Initial security hardening: added auth to all ProjectRoom routes, removed debug endpoints, implemented defense-in-depth |
