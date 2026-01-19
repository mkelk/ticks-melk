import { createContext } from '@lit/context';
import type { BoardTick, Epic, TickColumn } from '../types/tick.js';

// Board state interface - shared across all components via Lit Context
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

// Create the context with a unique symbol key
export const boardContext = createContext<BoardState>(Symbol('board'));

// Default/initial board state
export const initialBoardState: BoardState = {
  ticks: [],
  epics: [],
  selectedEpic: '',
  searchTerm: '',
  activeColumn: 'blocked',
  isMobile: false,
};
