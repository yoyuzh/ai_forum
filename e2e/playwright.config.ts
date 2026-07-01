import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright E2E configuration for AI Forum.
 * Configured with two projects:
 * - 'web': targeting the user-facing web application at http://localhost:5173
 * - 'admin': targeting the React Refine admin dashboard at http://localhost:5174
 */
export default defineConfig({
  testDir: './tests',
  /* Run tests in files in parallel */
  fullyParallel: true,
  /* Fail the build on CI if you accidentally left test.only in the source code. */
  forbidOnly: !!process.env.CI,
  /* Retry on CI only */
  retries: process.env.CI ? 2 : 0,
  /* Opt out of parallel tests on CI. */
  workers: process.env.CI ? 1 : undefined,
  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: [['html', { open: 'never' }], ['list']],
  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  use: {
    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: 'on-first-retry',
    /* Screenshot on failure */
    screenshot: 'only-on-failure',
  },

  /* Configure projects for major browsers */
  projects: [
    {
      name: 'web',
      testMatch: /web_t.*\.spec\.ts|integration\.spec\.ts|sanity\.spec\.ts/,
      use: {
        baseURL: 'http://localhost:5173',
        ...devices['Desktop Chrome'],
      },
    },
    {
      name: 'admin',
      testMatch: /admin_t.*\.spec\.ts|integration\.spec\.ts|sanity\.spec\.ts/,
      use: {
        baseURL: 'http://localhost:5174',
        ...devices['Desktop Chrome'],
      },
    },
  ],
});
