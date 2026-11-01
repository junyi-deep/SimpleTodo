import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./site/e2e",
  workers: 1,
  timeout: 30000,
  use: {
    baseURL: "http://127.0.0.1:9255/SimpleTodo/",
    viewport: { width: 1440, height: 1000 },
    channel: process.env.CI ? undefined : "chrome",
    trace: "retain-on-failure",
  },
  webServer: {
    command: "npm run dev:site",
    url: "http://127.0.0.1:9255/SimpleTodo/",
    reuseExistingServer: !process.env.CI,
  },
});
