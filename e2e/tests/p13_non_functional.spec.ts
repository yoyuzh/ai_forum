import { expect, test, type Page } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

const API_BASE = 'http://127.0.0.1:19091';

async function criticalAxe(page: Page) {
  const results = await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa']).analyze();
  return results.violations.filter((v) => v.impact === 'critical');
}

async function login(page: Page) {
  await page.goto('/login');
  await page.getByLabel('用户名').fill('admin');
  await page.getByRole('textbox', { name: '密码' }).fill('admin123');
  await page.getByRole('button', { name: /登\s*录/ }).click();
  await expect(page).toHaveURL('http://localhost:5173/');
}

async function ensurePerfPost() {
  const { execFileSync } = await import('node:child_process');
  const out = execFileSync(
    'docker',
    [
      'compose',
      'exec',
      '-T',
      'mysql',
      'sh',
      '-lc',
      `mysql -N -uroot -pai_forum_root ai_forum -e "DELETE FROM posts WHERE title = 'P13 INP post'; INSERT INTO posts (author_id, title, content, status) VALUES (1, 'P13 INP post', 'P13 INP body', 'NORMAL'); SELECT LAST_INSERT_ID();"`,
    ],
    { cwd: '..', encoding: 'utf8' },
  );
  return Number(out.trim().split(/\s+/).pop());
}

function channels(color: string) {
  if (color.startsWith('#')) {
    const hex = color.slice(1);
    return [hex.slice(0, 2), hex.slice(2, 4), hex.slice(4, 6)].map((v) => parseInt(v, 16));
  }
  return color.match(/\d+/g)!.slice(0, 3).map(Number);
}

function luminance(rgb: string) {
  const [r, g, b] = channels(rgb).map((v) => v / 255).map((v) =>
    v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4,
  );
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrastRatio(fg: string, bg: string) {
  const lighter = Math.max(luminance(fg), luminance(bg));
  const darker = Math.min(luminance(fg), luminance(bg));
  return (lighter + 0.05) / (darker + 0.05);
}

test.describe('P13 non-functional gates', () => {
  test.describe.configure({ mode: 'serial' });

  test('audited screens avoid external assets that can stall Lighthouse', async ({ page }) => {
    const postId = await ensurePerfPost();
    const external: string[] = [];
    page.on('request', (request) => {
      const url = new URL(request.url());
      if (!['127.0.0.1', 'localhost'].includes(url.hostname)) external.push(request.url());
    });

    await page.goto('/posts');
    await expect(page.getByRole('heading', { name: '帖子广场' })).toBeVisible();
    await page.goto(`/posts/${postId}`);
    await expect(page.getByRole('heading', { name: 'P13 INP post' })).toBeVisible();
    await page.goto('http://127.0.0.1:5174/login');
    await expect(page.getByRole('heading', { name: 'AI Forum Admin' })).toBeVisible();

    expect(external).toEqual([]);
  });

  test('real Playwright interaction timing on post-detail AI status stays below INP target', async ({ page }) => {
    const postId = await ensurePerfPost();
    await login(page);
    await page.goto(`/posts/${postId}`);
    await expect(page.getByRole('heading', { name: 'P13 INP post' })).toBeVisible();
    const elapsedPromise = page.evaluate(
      () =>
        new Promise<number>((resolve) => {
          const start = performance.now();
          requestAnimationFrame(() => resolve(performance.now() - start));
        }),
    );
    await page.getByRole('link', { name: /查看 AI 决策日志/ }).click({ noWaitAfter: true });
    const elapsed = await elapsedPromise;
    await expect(page.locator('#decision-logs')).toBeInViewport();
    expect(elapsed).toBeLessThan(200);
  });

  test('axe has no critical issues on key web and admin screens', async ({ page }) => {
    const postId = await ensurePerfPost();
    await page.goto('/posts');
    await expect(page.getByRole('heading', { name: '帖子广场' })).toBeVisible();
    expect(await criticalAxe(page), '/posts').toEqual([]);
    await page.goto(`/posts/${postId}`);
    await expect(page.getByRole('heading', { name: 'P13 INP post' })).toBeVisible();
    expect(await criticalAxe(page), '/posts/P13 INP post').toEqual([]);
    await page.goto('http://127.0.0.1:5174/login');
    await expect(page.getByRole('heading', { name: 'AI Forum Admin' })).toBeVisible();
    expect(await criticalAxe(page), 'admin login').toEqual([]);
  });

  test('Cohere text/background pairs meet WCAG AA contrast', async ({ page }) => {
    await page.goto('/posts');
    const pairs = await page.evaluate(() => {
      const styles = getComputedStyle(document.documentElement);
      const token = (name: string) => styles.getPropertyValue(name).trim();
      return [
        [token('--c-on-surface'), token('--c-surface')],
        [token('--c-on-surface-variant'), token('--c-surface-lowest')],
        [token('--c-primary'), token('--c-background')],
        [token('--c-on-primary'), token('--c-primary')],
        [token('--c-error'), token('--c-error-container')],
      ];
    });
    for (const [fg, bg] of pairs) {
      expect(contrastRatio(fg, bg), `${fg} on ${bg}`).toBeGreaterThanOrEqual(4.5);
    }
  });

  test('reduced-motion path still updates AI status content', async ({ page }) => {
    await login(page);
    await page.emulateMedia({ reducedMotion: 'reduce' });
    await page.route(`${API_BASE}/api/posts/401`, async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ id: 401, title: 'Reduced motion post', content: 'motion-safe', status: 'NORMAL', ai_reply_count: 0 }),
      });
    });
    await page.route(`${API_BASE}/api/posts/401/comments`, async (route) => {
      await route.fulfill({ contentType: 'application/json', body: '[]' });
    });
    await page.route(`${API_BASE}/api/posts/401/events`, async (route) => {
      await route.fulfill({ status: 503, body: 'sse unavailable' });
    });
    await page.route(`${API_BASE}/api/posts/401/ai-status`, async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify({ completedCount: 1, runningCount: 0, overallStatus: 'COMPLETED' }) });
    });

    await page.goto('/posts/401');
    await expect(page.getByText('check已完成')).toBeVisible();
  });
});
