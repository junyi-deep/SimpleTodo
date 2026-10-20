import { defineConfig } from "@playwright/test";
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  workers: 1,
  timeout: 30000,
  use: {
    baseURL: "http://127.0.0.1:9245",
    viewport: { width: 460, height: 760 },
    channel: "chrome",
    trace: "retain-on-failure",
  },
  webServer: [
    {
      command: "go run ./cmd/preview",
      cwd: "..",
      url: "http://127.0.0.1:9246/health",
      reuseExistingServer: !process.env.CI,
    },
    {
      command: "npm run dev -- --host 127.0.0.1",
      url: "http://127.0.0.1:9245",
      reuseExistingServer: !process.env.CI,
    },
  ],
});
