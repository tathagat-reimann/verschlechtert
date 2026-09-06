import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    // bind to all interfaces so the Docker container is reachable from the host
    host: true,
  },
})
