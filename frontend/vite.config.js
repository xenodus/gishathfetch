import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

const signatureDirectoryPath = "/.well-known/http-message-signatures-directory";
const signatureDirectoryMediaType =
  "application/http-message-signatures-directory+json";

const apiProxyTarget =
  process.env.VITE_API_PROXY_TARGET || "https://api.gishathfetch.com";

function configureApiProxy(proxy) {
  proxy.on("proxyReq", (proxyReq) => {
    const verifySecret = process.env.VITE_API_ORIGIN_VERIFY_SECRET;
    if (verifySecret) {
      proxyReq.setHeader("X-Origin-Verify", verifySecret);
    }
  });
}

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
      "/search": {
        target: apiProxyTarget,
        changeOrigin: true,
        configure: configureApiProxy,
      },
      "/session": {
        target: apiProxyTarget,
        changeOrigin: true,
        configure: configureApiProxy,
      },
    },
  },
});
