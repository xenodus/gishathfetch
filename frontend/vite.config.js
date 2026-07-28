import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

const signatureDirectoryPath = "/.well-known/http-message-signatures-directory";
const signatureDirectoryMediaType =
  "application/http-message-signatures-directory+json";

// https://vite.dev/config/
export default defineConfig({
  envPrefix: ["VITE_"],
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
      "/api": {
        target:
          process.env.VITE_API_PROXY_TARGET || "https://api.gishathfetch.com",
        changeOrigin: true,
        rewrite: (path) => {
          const stripped = path.replace(/^\/api/, "");
          return stripped === "" ? "/" : stripped;
        },
        configure: (proxy) => {
          proxy.on("proxyReq", (proxyReq) => {
            const verifySecret = process.env.VITE_API_ORIGIN_VERIFY_SECRET;
            if (verifySecret) {
              proxyReq.setHeader("X-Origin-Verify", verifySecret);
            }
          });
        },
      },
    },
  },
});
