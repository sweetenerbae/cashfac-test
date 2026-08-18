import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    host: "0.0.0.0",
    port: 5173,
    proxy: {
      "/api": {
        target: process.env.VITE_API_TARGET || "http://localhost:8081",
        changeOrigin: true
      },
      "/health": {
        target: process.env.VITE_API_TARGET || "http://localhost:8081",
        changeOrigin: true
      },
      "/docs": {
        target: process.env.VITE_API_TARGET || "http://localhost:8081",
        changeOrigin: true
      },
      "/openapi.yaml": {
        target: process.env.VITE_API_TARGET || "http://localhost:8081",
        changeOrigin: true
      }
    }
  }
});
