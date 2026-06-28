import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

const signatureDirectoryPath = "/.well-known/http-message-signatures-directory";
const signatureDirectoryMediaType =
  "application/http-message-signatures-directory+json";

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    {
      name: "well-known-signature-directory-content-type",
      configureServer(server) {
        server.middlewares.use((req, res, next) => {
          if (req.url === signatureDirectoryPath) {
            res.setHeader("Content-Type", signatureDirectoryMediaType);
          }
          next();
        });
      },
      configurePreviewServer(server) {
        server.middlewares.use((req, res, next) => {
          if (req.url === signatureDirectoryPath) {
            res.setHeader("Content-Type", signatureDirectoryMediaType);
          }
          next();
        });
      },
    },
  ],
  base: "",
  server: {
    proxy: {
      "/analytics": {
        target: "https://gishathfetch.com",
        changeOrigin: true,
      },
    },
  },
});
