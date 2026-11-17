import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    hmr: process.env.WEB_HMR_CLIENT_PORT
      ? { clientPort: Number(process.env.WEB_HMR_CLIENT_PORT) }
      : undefined,
    proxy: {
      '/api/menu': { target: 'http://menu-api:8000' },
      '/api/orders': { target: 'http://orders-api:8080' },
      '/api/stats': { target: 'http://stats-api:8080' },
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: './src/test-setup.js',
    clearMocks: true,
  },
});
