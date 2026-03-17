import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import ResultModal from '../components/ResultModal';

describe('ResultModal', () => {
  test('renders nothing when result is null', () => {
    const { container } = render(<ResultModal result={null} onNewGame={() => {}} />);
    expect(container.firstChild).toBeNull();
  });

  test('displays winner team', () => {
    const result = { winnerTeam: 1, finalScores: { team1: 10, team2: 5 } };
    render(<ResultModal result={result} onNewGame={() => {}} />);
    expect(screen.getByText('Team 1 の勝利!')).toBeInTheDocument();
    expect(screen.getByText('ゲーム終了！')).toBeInTheDocument();
  });

  test('displays draw result', () => {
    const result = { winnerTeam: 0, finalScores: { team1: 5, team2: 5 } };
    render(<ResultModal result={result} onNewGame={() => {}} />);
    expect(screen.getByText('引き分け')).toBeInTheDocument();
  });

  test('displays final scores', () => {
    const result = { winnerTeam: 2, finalScores: { team1: 3, team2: 8 } };
    render(<ResultModal result={result} onNewGame={() => {}} />);
    expect(screen.getByText('3')).toBeInTheDocument();
    expect(screen.getByText('8')).toBeInTheDocument();
  });

  test('calls onNewGame when button is clicked', () => {
    const mockNewGame = jest.fn();
    const result = { winnerTeam: 1, finalScores: { team1: 10, team2: 5 } };
    render(<ResultModal result={result} onNewGame={mockNewGame} />);
    fireEvent.click(screen.getByText('新しい問題で遊ぶ'));
    expect(mockNewGame).toHaveBeenCalledTimes(1);
  });
});
