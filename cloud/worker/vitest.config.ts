import { defineWorkersConfig } from "@cloudflare/vitest-pool-workers/config";
import { readFileSync } from "fs";

// Read schema from file
const schema = readFileSync("./schema.sql", "utf-8");

export default defineWorkersConfig({
  test: {
    poolOptions: {
      workers: {
        wrangler: { configPath: "./wrangler.toml" },
        miniflare: {
          // Configure test-specific bindings with schema
          d1Databases: {
            DB: schema,
          },
        },
        // Disable isolated storage to avoid issues with DO cleanup
        isolatedStorage: false,
      },
    },
    include: ["test/**/*.test.ts"],
  },
});
