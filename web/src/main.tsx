import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';

document.documentElement.dataset.apiMode = import.meta.env.VITE_API_MODE === 'real' ? 'real' : 'mock';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
