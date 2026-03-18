import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import LoginRegisterModal from '../components/LoginRegisterModal';

jest.mock('react-router-dom', () => ({
  useNavigate: () => jest.fn(),
}), { virtual: true });

jest.mock('axios', () => ({
  post: jest.fn(),
}));

describe('LoginRegisterModal', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders nothing when isOpen is false', () => {
    const { container } = render(<LoginRegisterModal isOpen={false} onClose={() => {}} />);
    expect(container.firstChild).toBeNull();
  });

  test('renders modal when isOpen is true', () => {
    render(<LoginRegisterModal isOpen={true} onClose={() => {}} />);
    expect(screen.getByText('ログイン / 新規登録')).toBeInTheDocument();
  });

  test('renders username and password inputs', () => {
    render(<LoginRegisterModal isOpen={true} onClose={() => {}} />);
    expect(screen.getByLabelText('ユーザー名')).toBeInTheDocument();
    expect(screen.getByLabelText('パスワード')).toBeInTheDocument();
  });

  test('renders register and login buttons', () => {
    render(<LoginRegisterModal isOpen={true} onClose={() => {}} />);
    expect(screen.getByText('新規登録')).toBeInTheDocument();
    expect(screen.getByText('ログイン')).toBeInTheDocument();
  });

  test('shows error when fields are empty on register', () => {
    render(<LoginRegisterModal isOpen={true} onClose={() => {}} />);
    fireEvent.click(screen.getByText('新規登録'));
    expect(screen.getByText('ユーザー名とパスワードを入力してください。')).toBeInTheDocument();
  });

  test('shows error when fields are empty on login', () => {
    render(<LoginRegisterModal isOpen={true} onClose={() => {}} />);
    fireEvent.click(screen.getByText('ログイン'));
    expect(screen.getByText('ユーザー名とパスワードを入力してください。')).toBeInTheDocument();
  });

  test('calls onClose when close button is clicked', () => {
    const mockClose = jest.fn();
    render(<LoginRegisterModal isOpen={true} onClose={mockClose} />);
    fireEvent.click(screen.getByText('×'));
    expect(mockClose).toHaveBeenCalledTimes(1);
  });

  test('calls onClose when overlay is clicked', () => {
    const mockClose = jest.fn();
    render(<LoginRegisterModal isOpen={true} onClose={mockClose} />);
    const overlay = screen.getByText('ログイン / 新規登録').closest('.modal-overlay');
    fireEvent.click(overlay);
    expect(mockClose).toHaveBeenCalledTimes(1);
  });
});
