import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import UserList from '../components/UserList';

describe('UserList', () => {
  const mockUsers = [
    { id: '1', name: 'Player1', team: 1, role: 'guest' },
    { id: '2', name: 'Player2', team: 2, role: 'user' },
  ];
  const mockMyPlayer = { id: '1', name: 'Player1', team: 1, role: 'guest' };

  test('displays member count', () => {
    render(<UserList users={mockUsers} myPlayer={mockMyPlayer} onChangeName={() => {}} />);
    expect(screen.getByText(/参加中のメンバー \(2人\)/)).toBeInTheDocument();
  });

  test('displays all player names', () => {
    render(<UserList users={mockUsers} myPlayer={mockMyPlayer} onChangeName={() => {}} />);
    expect(screen.getByText('Player1')).toBeInTheDocument();
    expect(screen.getByText('Player2')).toBeInTheDocument();
  });

  test('shows edit icon only for own player', () => {
    render(<UserList users={mockUsers} myPlayer={mockMyPlayer} onChangeName={() => {}} />);
    const editIcons = screen.getAllByText('✏️');
    expect(editIcons).toHaveLength(1);
  });

  test('switches to edit mode when name is clicked', () => {
    render(<UserList users={mockUsers} myPlayer={mockMyPlayer} onChangeName={() => {}} />);
    fireEvent.click(screen.getByText('Player1'));
    const input = screen.getByDisplayValue('Player1');
    expect(input).toBeInTheDocument();
  });

  test('submits new name on form submit', () => {
    const mockChangeName = jest.fn();
    render(<UserList users={mockUsers} myPlayer={mockMyPlayer} onChangeName={mockChangeName} />);
    fireEvent.click(screen.getByText('Player1'));

    const input = screen.getByDisplayValue('Player1');
    fireEvent.change(input, { target: { value: 'NewName' } });
    fireEvent.submit(input.closest('form'));

    expect(mockChangeName).toHaveBeenCalledWith('NewName');
  });
});
