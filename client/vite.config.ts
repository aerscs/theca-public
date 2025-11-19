import { defineConfig } from "vite";
import { resolve } from "node:path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react-swc";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: [{ find: "@", replacement: resolve(__dirname, "./src") }],
  },
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8080/",
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
    },
  },
});
