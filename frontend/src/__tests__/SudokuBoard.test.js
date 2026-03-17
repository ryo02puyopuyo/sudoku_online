import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import SudokuBoard from '../components/SudokuBoard';

const createEmptyBoard = () => {
  const board = [];
  for (let r = 0; r < 9; r++) {
    const row = [];
    for (let c = 0; c < 9; c++) {
      row.push({ value: 0, status: 'empty', filledByTeam: 0, isHotSpot: false });
    }
    board.push(row);
  }
  return board;
};

const createMixedBoard = () => {
  const board = createEmptyBoard();
  board[0][0] = { value: 5, status: 'fixed', filledByTeam: 0, isHotSpot: false };
  board[1][1] = { value: 3, status: 'correct', filledByTeam: 1, isHotSpot: false };
  board[2][2] = { value: 7, status: 'wrong', filledByTeam: 2, isHotSpot: false };
  board[3][3] = { value: 0, status: 'empty', filledByTeam: 0, isHotSpot: true };
  return board;
};

describe('SudokuBoard', () => {
  test('renders 9x9 grid (81 cells)', () => {
    const board = createEmptyBoard();
    render(<SudokuBoard boardState={board} onCellClick={() => {}} onNewGameClick={() => {}} />);
    const table = screen.getByRole('table');
    const cells = table.querySelectorAll('td');
    expect(cells).toHaveLength(81);
  });

  test('renders number selection buttons 1-9 and clear', () => {
    const board = createEmptyBoard();
    render(<SudokuBoard boardState={board} onCellClick={() => {}} onNewGameClick={() => {}} />);
    for (let i = 1; i <= 9; i++) {
      expect(screen.getByText(String(i))).toBeInTheDocument();
    }
    expect(screen.getByText('消')).toBeInTheDocument();
  });

  test('displays fixed cell values in the board', () => {
    const board = createMixedBoard();
    render(<SudokuBoard boardState={board} onCellClick={() => {}} onNewGameClick={() => {}} />);
    const fives = screen.getAllByText('5');
    expect(fives.length).toBeGreaterThanOrEqual(2);
    const table = screen.getByRole('table');
    expect(table).toHaveTextContent('5');
  });

  test('calls onCellClick when editable cell is clicked with selected number', () => {
    const board = createEmptyBoard();
    const mockClick = jest.fn();
    render(<SudokuBoard boardState={board} onCellClick={mockClick} onNewGameClick={() => {}} />);

    const buttons = screen.getAllByRole('button');
    const button1 = buttons.find(b => b.textContent === '1');
    fireEvent.click(button1);

    const table = screen.getByRole('table');
    const firstCell = table.querySelector('td');
    fireEvent.click(firstCell);

    expect(mockClick).toHaveBeenCalledWith(0, 0, 1);
  });

  test('does not call onCellClick when no number is selected', () => {
    const board = createEmptyBoard();
    const mockClick = jest.fn();
    render(<SudokuBoard boardState={board} onCellClick={mockClick} onNewGameClick={() => {}} />);

    const table = screen.getByRole('table');
    const firstCell = table.querySelector('td');
    fireEvent.click(firstCell);

    expect(mockClick).not.toHaveBeenCalled();
  });

  test('renders new game button', () => {
    const board = createEmptyBoard();
    render(<SudokuBoard boardState={board} onCellClick={() => {}} onNewGameClick={() => {}} />);
    expect(screen.getByText('新しい問題')).toBeInTheDocument();
  });

  test('calls onNewGameClick when new game button is clicked', () => {
    const board = createEmptyBoard();
    const mockNewGame = jest.fn();
    render(<SudokuBoard boardState={board} onCellClick={() => {}} onNewGameClick={mockNewGame} />);
    fireEvent.click(screen.getByText('新しい問題'));
    expect(mockNewGame).toHaveBeenCalledTimes(1);
  });
});
