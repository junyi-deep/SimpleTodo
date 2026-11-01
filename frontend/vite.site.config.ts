import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  root: "site",
  base: "/SimpleTodo/",
  plugins: [vue(), tailwindcss()],
  server: { host: "127.0.0.1", port: 9255, strictPort: true },
  build: { outDir: "../site-dist", emptyOutDir: true },
});
