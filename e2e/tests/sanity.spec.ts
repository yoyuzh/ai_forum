import { test, expect } from '@playwright/test';

test.describe('E2E Sanity Check', () => {
  test('basic assertion should pass', async () => {
    expect(1 + 1).toBe(2);
  });

  test('page object should be defined', async ({ page }) => {
    expect(page).toBeDefined();
  });
});
