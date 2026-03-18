import React from 'react';
import { render, screen } from '@testing-library/react';

jest.mock('axios', () => ({
  defaults: { baseURL: '', withCredentials: false },
  interceptors: {
    request: { use: jest.fn() },
  },
}));

jest.mock('react-router-dom', () => ({
  Routes: ({ children }) => <div>{children}</div>,
  Route: ({ element }) => element,
  BrowserRouter: ({ children }) => <div>{children}</div>,
  useNavigate: () => jest.fn(),
}), { virtual: true });

jest.mock('../pages/Landing', () => {
  return function MockLanding() {
    return <div>ナンプレバトル</div>;
  };
});

jest.mock('../pages/Game', () => {
  return function MockGame() {
    return <div>Game Page</div>;
  };
});

const App = require('../App').default;

describe('App', () => {
  test('renders app container', () => {
    const { container } = render(<App />);
    expect(container.querySelector('.app-container')).toBeInTheDocument();
  });

  test('renders landing page content', () => {
    render(<App />);
    expect(screen.getByText('ナンプレバトル')).toBeInTheDocument();
  });
});
