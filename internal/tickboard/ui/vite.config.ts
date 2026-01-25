import { defineConfig, type Plugin } from 'vite';
import { viteStaticCopy } from 'vite-plugin-static-copy';
import { resolve } from 'path';

// Plugin to serve index.html for /p/* routes (cloud mode testing)
function cloudModeRoutes(): Plugin {
  return {
    name: 'cloud-mode-routes',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        // Serve index.html for /p/{projectId}/* paths
        if (req.url?.startsWith('/p/')) {
          req.url = '/index.html';
        }
        next();
      });
    },
  };
}

export default defineConfig({
  base: './',  // Use relative paths for cloud proxy compatibility
  server: {
    proxy: {
      // Proxy API requests to testrig or local Go server
      '/api': {
        target: process.env.VITE_API_URL || 'http://localhost:18787',
        changeOrigin: true,
        ws: true,  // Enable WebSocket proxy for cloud mode
      },
    },
  },
  build: {
    outDir: '../server/static',
    emptyOutDir: true,
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html'),
        app: resolve(__dirname, 'app.html'),
      },
      output: {
        entryFileNames: 'assets/[name]-[hash].js',
        chunkFileNames: 'assets/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash][extname]',
      },
    },
  },
  plugins: [
    cloudModeRoutes(),
    viteStaticCopy({
      targets: [
        {
          src: 'node_modules/@shoelace-style/shoelace/dist/assets/icons/*',
          dest: 'shoelace/assets/icons',
        },
      ],
    }),
  ],
});
