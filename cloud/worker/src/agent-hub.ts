/**
 * AgentHub Durable Object
 *
 * Registry of online sync-mode boards (ProjectRoom connections).
 * Used to track which boards are currently connected and online.
 */

import type { Env } from "./index";

export class AgentHub {
  private state: DurableObjectState;
  private env: Env;
  private syncBoards: Set<string> = new Set(); // boardName (sync mode - ProjectRoom connections)

  constructor(state: DurableObjectState, env: Env) {
    this.state = state;
    this.env = env;

    // Restore state from storage (blocking to ensure it's ready before fetch)
    this.state.blockConcurrencyWhile(async () => {
      const storedSyncBoards = await this.state.storage.get<string[]>("syncBoards");
      if (storedSyncBoards) {
        this.syncBoards = new Set(storedSyncBoards);
        console.log(`[AgentHub] Restored ${this.syncBoards.size} sync boards from storage`);
      }
    });
  }

  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);

    // List connected boards (for debugging/admin)
    if (url.pathname === "/boards") {
      const boards = Array.from(this.syncBoards);
      return Response.json({ boards, count: boards.length });
    }

    // Register a sync-mode board as online (called by ProjectRoom)
    if (url.pathname.startsWith("/sync-register/")) {
      const boardName = decodeURIComponent(url.pathname.slice("/sync-register/".length));
      this.syncBoards.add(boardName);
      // Persist to storage so it survives hibernation/restart
      await this.state.storage.put("syncBoards", Array.from(this.syncBoards));
      console.log(`[AgentHub] Sync board registered: ${boardName} (total: ${this.syncBoards.size})`);
      return Response.json({ ok: true, board: boardName });
    }

    // Unregister a sync-mode board (called by ProjectRoom)
    if (url.pathname.startsWith("/sync-unregister/")) {
      const boardName = decodeURIComponent(url.pathname.slice("/sync-unregister/".length));
      this.syncBoards.delete(boardName);
      // Persist to storage so it survives hibernation/restart
      await this.state.storage.put("syncBoards", Array.from(this.syncBoards));
      console.log(`[AgentHub] Sync board unregistered: ${boardName} (total: ${this.syncBoards.size})`);
      return Response.json({ ok: true, board: boardName });
    }

    // Clear all sync boards (admin cleanup)
    if (url.pathname === "/sync-clear") {
      const cleared = this.syncBoards.size;
      this.syncBoards.clear();
      await this.state.storage.put("syncBoards", []);
      console.log(`[AgentHub] Cleared all sync boards (${cleared} total)`);
      return Response.json({ ok: true, cleared });
    }

    return new Response("Not found", { status: 404 });
  }
}
