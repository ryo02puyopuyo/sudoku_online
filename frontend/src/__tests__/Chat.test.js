import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import Chat from '../components/Chat';

// scrollIntoView はJSDOMに存在しないためモック
beforeAll(() => {
  Element.prototype.scrollIntoView = jest.fn();
});

describe('Chat', () => {
  const sampleMessages = [
    { senderName: 'Alice', senderTeam: 1, message: 'こんにちは', timestamp: '12:00' },
    { senderName: 'Bob', senderTeam: 2, message: 'よろしく', timestamp: '12:01' },
  ];

  test('displays chat messages', () => {
    render(<Chat messages={sampleMessages} onSendMessage={() => {}} />);
    expect(screen.getByText('こんにちは')).toBeInTheDocument();
    expect(screen.getByText('よろしく')).toBeInTheDocument();
  });

  test('displays sender names', () => {
    render(<Chat messages={sampleMessages} onSendMessage={() => {}} />);
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('Bob')).toBeInTheDocument();
  });

  test('displays timestamps', () => {
    render(<Chat messages={sampleMessages} onSendMessage={() => {}} />);
    expect(screen.getByText('12:00')).toBeInTheDocument();
    expect(screen.getByText('12:01')).toBeInTheDocument();
  });

  test('renders empty message list', () => {
    render(<Chat messages={[]} onSendMessage={() => {}} />);
    expect(screen.getByText('チャット')).toBeInTheDocument();
  });

  test('does not send empty messages', () => {
    const mockSend = jest.fn();
    render(<Chat messages={[]} onSendMessage={mockSend} />);
    const form = screen.getByText('送信').closest('form');
    fireEvent.submit(form);
    expect(mockSend).not.toHaveBeenCalled();
  });

  test('sends message and clears input', () => {
    const mockSend = jest.fn();
    render(<Chat messages={[]} onSendMessage={mockSend} />);
    const input = screen.getByPlaceholderText('メッセージを入力...');
    fireEvent.change(input, { target: { value: 'テストメッセージ' } });
    const form = screen.getByText('送信').closest('form');
    fireEvent.submit(form);
    expect(mockSend).toHaveBeenCalledWith('テストメッセージ');
    expect(input.value).toBe('');
  });
});
