import React from 'react';
import { render, screen } from '@testing-library/react';
import ScoreBoard from '../components/ScoreBoard';

describe('ScoreBoard', () => {
  test('displays team scores', () => {
    render(<ScoreBoard scores={{ team1: 5, team2: 3 }} />);
    expect(screen.getByText('5')).toBeInTheDocument();
    expect(screen.getByText('3')).toBeInTheDocument();
  });

  test('displays team labels', () => {
    render(<ScoreBoard scores={{ team1: 0, team2: 0 }} />);
    expect(screen.getByText(/Team 1/)).toBeInTheDocument();
    expect(screen.getByText(/Team 2/)).toBeInTheDocument();
  });

  test('displays zero scores correctly', () => {
    render(<ScoreBoard scores={{ team1: 0, team2: 0 }} />);
    const zeros = screen.getAllByText('0');
    expect(zeros).toHaveLength(2);
  });
});
