import { test, expect } from '@playwright/test';

test.describe('P13 live-stack sanity', () => {
  test('web is served in real API mode', async ({ page, context }) => {
    await context.clearCookies();
    await page.addInitScript(() => localStorage.clear());
    const response = await page.goto('http://127.0.0.1:5173/posts');
    expect(response?.status()).toBe(200);
    await expect(page.getByRole('heading', { name: '帖子广场' })).toBeVisible();
    await expect(page.locator('html')).toHaveAttribute('data-api-mode', 'real');
    await expect(page.getByText(/mock/i)).toHaveCount(0);
  });

  test('admin is served in real API mode', async ({ page }) => {
    const response = await page.goto('http://127.0.0.1:5174/login');
    expect(response?.status()).toBe(200);
    await expect(page.getByRole('heading', { name: 'AI Forum Admin' })).toBeVisible();
    await expect(page.locator('html')).toHaveAttribute('data-api-mode', 'real');
    await expect(page.getByText(/mock/i)).toHaveCount(0);
  });
});
