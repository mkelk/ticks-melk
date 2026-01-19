import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { provide } from '@lit/context';
import { boardContext, initialBoardState, type BoardState } from '../contexts/board-context.js';
import type { BoardTick, TickColumn } from '../types/tick.js';

// Column definitions for the kanban board
const COLUMNS = [
  { id: 'blocked' as TickColumn, name: 'Blocked', color: 'var(--red)', icon: '⊘' },
  { id: 'ready' as TickColumn, name: 'Agent Queue', color: 'var(--blue)', icon: '▶' },
  { id: 'agent' as TickColumn, name: 'In Progress', color: 'var(--peach)', icon: '●' },
  { id: 'human' as TickColumn, name: 'Needs Human', color: 'var(--yellow)', icon: '👤' },
  { id: 'done' as TickColumn, name: 'Done', color: 'var(--green)', icon: '✓' },
] as const;

@customElement('tick-board')
export class TickBoard extends LitElement {
  static styles = css`
    :host {
      display: flex;
      flex-direction: column;
      min-height: 100vh;
    }

    /* Header */
    header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 1rem;
      padding: 1rem 1.5rem;
      background-color: var(--surface0);
      border-bottom: 1px solid var(--surface1);
    }

    .header-left {
      display: flex;
      align-items: center;
      gap: 1rem;
    }

    .header-left h1 {
      font-size: 1.25rem;
      font-weight: 600;
      color: var(--rosewater);
      margin: 0;
    }

    .repo-badge {
      font-size: 0.75rem;
      padding: 0.25rem 0.5rem;
      background: var(--surface1);
      border-radius: 4px;
      font-family: monospace;
      color: var(--subtext0);
    }

    .header-center {
      flex: 1;
      display: flex;
      justify-content: center;
      gap: 0.75rem;
      max-width: 600px;
    }

    .header-center sl-input {
      flex: 1;
      max-width: 250px;
    }

    .header-center sl-select {
      min-width: 180px;
    }

    .header-right {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    /* Mobile menu button */
    .menu-toggle {
      display: none;
      background: none;
      border: none;
      color: var(--text);
      font-size: 1.5rem;
      cursor: pointer;
      padding: 0.5rem;
      border-radius: 6px;
    }

    .menu-toggle:hover {
      background: var(--surface1);
    }

    /* Kanban board */
    main {
      flex: 1;
      padding: 1rem;
      overflow: hidden;
    }

    .kanban-board {
      display: flex;
      gap: 1rem;
      height: calc(100vh - 80px);
      overflow-x: auto;
    }

    /* Column placeholder styling */
    .column-placeholder {
      flex: 1;
      min-width: 220px;
      max-width: 320px;
      background: var(--surface0);
      border-radius: 8px;
      display: flex;
      flex-direction: column;
    }

    .column-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0.75rem 1rem;
      border-bottom: 1px solid var(--surface1);
    }

    .column-title {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-weight: 600;
      font-size: 0.875rem;
    }

    .column-icon {
      font-size: 0.75rem;
    }

    .column-count {
      font-size: 0.75rem;
      padding: 0.125rem 0.5rem;
      background: var(--surface1);
      border-radius: 999px;
      color: var(--subtext0);
    }

    .column-content {
      flex: 1;
      padding: 0.5rem;
      overflow-y: auto;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      color: var(--subtext0);
      font-size: 0.875rem;
    }

    /* Mobile column selector */
    .mobile-column-select {
      display: none;
      padding: 0.75rem 1rem;
      background: var(--surface0);
      border-bottom: 1px solid var(--surface1);
    }

    .mobile-column-select sl-select {
      width: 100%;
    }

    /* Responsive */
    @media (max-width: 768px) {
      .header-center {
        display: none;
      }

      .menu-toggle {
        display: block;
      }

      .kanban-board {
        gap: 0.75rem;
      }

      .column-placeholder {
        min-width: 260px;
        flex: 0 0 260px;
      }
    }

    @media (max-width: 480px) {
      header {
        padding: 0.75rem 1rem;
      }

      .repo-badge {
        display: none;
      }

      .header-left h1 {
        font-size: 1.125rem;
      }

      main {
        padding: 0;
      }

      .mobile-column-select {
        display: block;
      }

      .kanban-board {
        display: block;
        height: calc(100vh - 140px);
        overflow-y: auto;
      }

      .column-placeholder {
        display: none;
        width: 100%;
        max-width: none;
        height: 100%;
      }

      .column-placeholder.mobile-active {
        display: flex;
      }
    }
  `;

  // Provide board context to all child components
  @provide({ context: boardContext })
  @state()
  boardState: BoardState = { ...initialBoardState };

  // Local state
  @state() private ticks: BoardTick[] = [];
  @state() private selectedEpic = '';
  @state() private searchTerm = '';
  @state() private activeColumn: TickColumn = 'blocked';
  @state() private isMobile = window.matchMedia('(max-width: 480px)').matches;

  private mediaQuery = window.matchMedia('(max-width: 480px)');

  connectedCallback() {
    super.connectedCallback();
    this.mediaQuery.addEventListener('change', this.handleMediaChange);
    this.updateBoardState();
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    this.mediaQuery.removeEventListener('change', this.handleMediaChange);
  }

  private handleMediaChange = (e: MediaQueryListEvent) => {
    this.isMobile = e.matches;
    this.updateBoardState();
  };

  // Update the shared board state when local state changes
  private updateBoardState() {
    this.boardState = {
      ticks: this.ticks,
      epics: [],
      selectedEpic: this.selectedEpic,
      searchTerm: this.searchTerm,
      activeColumn: this.activeColumn,
      isMobile: this.isMobile,
    };
  }

  private handleSearchInput(e: Event) {
    const input = e.target as HTMLInputElement;
    this.searchTerm = input.value;
    this.updateBoardState();
  }

  private handleMobileColumnChange(e: Event) {
    const select = e.target as HTMLSelectElement;
    this.activeColumn = select.value as TickColumn;
    this.updateBoardState();
  }

  private getColumnTicks(columnId: TickColumn): BoardTick[] {
    return this.ticks.filter(tick => tick.column === columnId);
  }

  render() {
    return html`
      <header>
        <div class="header-left">
          <button class="menu-toggle" aria-label="Menu">☰</button>
          <h1>Tick Board</h1>
          <span class="repo-badge">ticks</span>
        </div>

        <div class="header-center">
          <sl-input
            placeholder="Search by ID or title..."
            size="small"
            clearable
            .value=${this.searchTerm}
            @sl-input=${this.handleSearchInput}
          >
            <sl-icon name="search" slot="prefix"></sl-icon>
          </sl-input>

          <sl-select
            placeholder="All Ticks"
            size="small"
            clearable
            .value=${this.selectedEpic}
          >
            <!-- Epic options will be populated from API -->
          </sl-select>
        </div>

        <div class="header-right">
          <sl-tooltip content="Create new tick">
            <sl-button variant="primary" size="small">
              <sl-icon name="plus-lg"></sl-icon>
            </sl-button>
          </sl-tooltip>
        </div>
      </header>

      <!-- Mobile column selector -->
      <div class="mobile-column-select">
        <sl-select .value=${this.activeColumn} @sl-change=${this.handleMobileColumnChange}>
          ${COLUMNS.map(col => html`
            <sl-option value=${col.id}>
              ${col.icon} ${col.name} (${this.getColumnTicks(col.id).length})
            </sl-option>
          `)}
        </sl-select>
      </div>

      <main>
        <div class="kanban-board">
          ${COLUMNS.map(col => html`
            <div
              class="column-placeholder ${this.activeColumn === col.id ? 'mobile-active' : ''}"
            >
              <div class="column-header">
                <span class="column-title">
                  <span class="column-icon" style="color: ${col.color}">${col.icon}</span>
                  ${col.name}
                </span>
                <span class="column-count">${this.getColumnTicks(col.id).length}</span>
              </div>
              <div class="column-content">
                <!-- tick-column components will render cards here in future tasks -->
                No ticks
              </div>
            </div>
          `)}
        </div>
      </main>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'tick-board': TickBoard;
  }
}
