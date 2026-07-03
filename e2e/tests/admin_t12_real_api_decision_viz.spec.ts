import { expect, test } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

const API_BASE = 'http://127.0.0.1:19091';

test.describe('P12 admin real API and decision visualization', () => {
  test('logs out and redirects to login on backend 401', async ({ page }) => {
    await page.route(`${API_BASE}/api/login`, async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify({ token: 'admin-token' }) });
    });
    await page.route(`${API_BASE}/api/me`, async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ id: 1, username: 'admin', role: 'ADMIN', permissions: [] }),
      });
    });
    await page.route(`${API_BASE}/api/admin/permissions`, async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify({ permissions: [] }) });
    });
    await page.route(`${API_BASE}/api/admin/users`, async (route) => {
      await route.fulfill({ status: 401, body: 'unauthorized' });
    });

    await page.goto('/login');
    await page.getByLabel('用户名').fill('admin');
    await page.getByLabel('密码').fill('admin123');
    await page.getByText(/登\s*录/).click();
    await expect(page.getByRole('heading', { name: /Dashboard|概览/ })).toBeVisible();

    await page.getByText('退出登录').click();
    await expect(page).toHaveURL(/\/login/);

    await page.evaluate(() => localStorage.setItem('ai_forum_admin_token', 'bad-token'));
    await page.goto('/users');
    await expect(page).toHaveURL(/\/login/);
  });

  test('logs in, gates RBAC visibility, renders decision detail, and passes axe', async ({ page }) => {
    await page.route(`${API_BASE}/api/admin/**`, async (route) => {
      const url = route.request().url();
      if (url.endsWith('/api/admin/permissions')) {
        await route.fulfill({
          contentType: 'application/json',
          body: JSON.stringify({ role: 'ADMIN', permissions: ['post:delete-any', 'ai_agent:update', 'decision_log:read'] }),
        });
        return;
      }
      if (url.endsWith('/api/admin/decision-logs')) {
        await route.fulfill({
          contentType: 'application/json',
          body: JSON.stringify([
            {
              id: 10,
              postId: 42,
              commentId: null,
              aiAgentId: 1001,
              aiAgentName: 'cohere_observer',
              triggerType: 'AUTO',
              willingnessScore: 0.32,
              thresholdValue: 0.6,
              decision: 'FALLBACK',
              reason: 'fallback-invoked',
              fallback: true,
              hitTags: ['topic:general'],
              taskId: 55,
              commentLink: null,
              createdAt: '2026-07-03T10:00:00Z',
            },
          ]),
        });
        return;
      }
      if (url.endsWith('/api/admin/ai-tasks/55/retry')) {
        await route.fulfill({ status: 403, body: 'forbidden' });
        return;
      }
      await route.fulfill({ contentType: 'application/json', body: '[]' });
    });
    await page.route(`${API_BASE}/api/login`, async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify({ token: 'admin-token' }) });
    });
    await page.route(`${API_BASE}/api/me`, async (route) => {
      const auth = route.request().headers().authorization ?? '';
      if (!auth.includes('admin-token')) {
        await route.fulfill({ status: 401, body: 'unauthorized' });
        return;
      }
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({
          id: 1,
          username: 'admin',
          role: 'ADMIN',
          permissions: ['post:delete-any', 'ai_agent:update', 'decision_log:read'],
        }),
      });
    });

    await page.goto('/login');
    await page.getByLabel('用户名').fill('admin');
    await page.getByLabel('密码').fill('admin123');
    await page.getByText(/登\s*录/).click();

    await expect(page.getByRole('heading', { name: /Dashboard|概览/ })).toBeVisible();
    await expect(page.getByRole('button', { name: /Retry|重试/ })).toHaveCount(0);

    const dashboardAxe = await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa']).analyze();
    expect(dashboardAxe.violations.filter((v) => v.impact === 'serious' || v.impact === 'critical')).toEqual([]);

    await page.getByRole('link', { name: /AI 决策/ }).click();
    await expect(page.getByText('cohere_observer').first()).toBeVisible();
    await page.getByRole('button', { name: 'decision detail 10' }).click();
    await expect(page.getByRole('dialog')).toContainText('0.32 / 0.60');
    await expect(page.getByRole('dialog')).toContainText('below threshold');
    await expect(page.getByRole('dialog')).toContainText('topic:general');
    await expect(page.getByRole('dialog')).toContainText('fallback-invoked');
    await expect(page.getByRole('dialog')).toContainText('Task #55');

    const decisionAxe = await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa']).analyze();
    expect(decisionAxe.violations.filter((v) => v.impact === 'serious' || v.impact === 'critical')).toEqual([]);

    const denied = await page.evaluate(async (url) => {
      const response = await fetch(url, { method: 'POST', headers: { Authorization: 'Bearer admin-token' } });
      return response.status;
    }, `${API_BASE}/api/admin/ai-tasks/55/retry`);
    expect(denied).toBe(403);
  });
});
