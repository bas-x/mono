import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { fileURLToPath, URL } from 'node:url';
import tailwindcss from '@tailwindcss/vite';

const workspaceRoot = fileURLToPath(new URL('..', import.meta.url));

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@backend-assets': fileURLToPath(new URL('../backend/assets', import.meta.url)),
    },
  },
  server: {
    fs: {
      allow: [workspaceRoot],
    },
  },
});
