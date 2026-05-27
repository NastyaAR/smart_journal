import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const apiTarget = process.env.VITE_BACKEND_URL || "http://localhost:3000";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: apiTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
    },
  },
  optimizeDeps: {
    include: [
      'micromark',
      'micromark-core-commonmark', 
      'micromark-extension-gfm',
      'micromark-util-character',
      'micromark-util-chunked',
      'micromark-util-combine-extensions',
    ],
  },
});
