/**
 * Board state context using Lit Context API.
 *
 * The root tick-board component provides this context to all children.
 * Child components consume it with `@consume({ context: boardContext, subscribe: true })`.
 *
 * @module board-context
 */
import { createContext } from '@lit/context';
import type { BoardTick, Epic, TickColumn } from '../types/tick.js';

/**
 * Shared board state interface.
 * Provided by tick-board, consumed by child components.
 */
export interface BoardState {
  // Data
  ticks: BoardTick[];
  epics: Epic[];

  // Filters
  selectedEpic: string;
  searchTerm: string;

  // UI state
  activeColumn: TickColumn;
  isMobile: boolean;
}

/** Board context instance. Use with @provide and @consume decorators. */
export const boardContext = createContext<BoardState>(Symbol('board'));

/** Default/initial board state. */
export const initialBoardState: BoardState = {
  ticks: [],
  epics: [],
  selectedEpic: '',
  searchTerm: '',
  activeColumn: 'blocked',
  isMobile: false,
};
