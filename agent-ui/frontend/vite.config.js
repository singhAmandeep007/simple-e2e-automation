import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  resolve: {
    alias: {
      // When running under `wails dev`, Wails auto-generates wailsjs/ at
      // frontend/wailsjs/. For standalone `npm run dev`, we alias it to
      // our mock stubs inside src/wailsjs/ so imports always resolve.
      //
      // Import paths in components:
      //   ../../wailsjs/go/main/App      → resolved here
      //   ../../wailsjs/runtime/runtime  → resolved here
      "../../wailsjs": path.resolve(__dirname, "src/wailsjs"),
      "../wailsjs": path.resolve(__dirname, "src/wailsjs"),
      wailsjs: path.resolve(__dirname, "src/wailsjs"),
    },
  },
});
