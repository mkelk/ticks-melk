import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'happy-dom',
    include: ['src/**/*.test.ts'],
    setupFiles: ['./src/comms/test-setup.ts'],
    coverage: {
      provider: 'v8',
      include: ['src/comms/**/*.ts'],
      exclude: ['src/comms/**/*.test.ts', 'src/comms/index.ts', 'src/comms/test-setup.ts'],
      reporter: ['text', 'html', 'lcov'],
      thresholds: {
        lines: 100,
        functions: 100,
        branches: 100,
        statements: 100,
      },
    },
  },
});
