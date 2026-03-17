import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import TeamSelector from '../components/TeamSelector';

describe('TeamSelector', () => {
  const mockPlayer = { id: '1', name: 'TestPlayer', team: 1, role: 'guest' };

  test('renders nothing when myPlayer is null', () => {
    const { container } = render(<TeamSelector myPlayer={null} onChangeTeam={() => {}} />);
    expect(container.firstChild).toBeNull();
  });

  test('renders team buttons', () => {
    render(<TeamSelector myPlayer={mockPlayer} onChangeTeam={() => {}} />);
    expect(screen.getByText('Team 1')).toBeInTheDocument();
    expect(screen.getByText('Team 2')).toBeInTheDocument();
  });

  test('marks current team as active', () => {
    render(<TeamSelector myPlayer={mockPlayer} onChangeTeam={() => {}} />);
    const team1Button = screen.getByText('Team 1');
    expect(team1Button).toHaveClass('active');
    const team2Button = screen.getByText('Team 2');
    expect(team2Button).not.toHaveClass('active');
  });

  test('calls onChangeTeam when button is clicked', () => {
    const mockOnChange = jest.fn();
    render(<TeamSelector myPlayer={mockPlayer} onChangeTeam={mockOnChange} />);
    fireEvent.click(screen.getByText('Team 2'));
    expect(mockOnChange).toHaveBeenCalledWith(2);
  });
});
