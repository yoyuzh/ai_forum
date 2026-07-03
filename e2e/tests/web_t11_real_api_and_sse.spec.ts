import { expect, test } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

const API_BASE = 'http://127.0.0.1:19091';

test.describe('P11 web real API and SSE', () => {
  test.beforeEach(async ({ page }) => {
    await page.route(`${API_BASE}/api/notifications/unread-count`, async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify({ count: 0 }) });
    });
    await page.route(`${API_BASE}/api/notifications`, async (route) => {
      await route.fulfill({ contentType: 'application/json', body: '[]' });
    });
  });

  test('real mode uses live API, surfaces 429, and redirects on 401', async ({ page }) => {
    let sawPosts = false;
    let sawLike = false;
    let sawFavorite = false;
    await page.route(`${API_BASE}/api/posts`, async (route) => {
      sawPosts = true;
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify([
          {
            id: 101,
            title: 'Real backend post',
            content: 'Loaded from intercepted API',
            status: 'NORMAL',
          },
        ]),
      });
    });
    await page.route(`${API_BASE}/api/posts/101`, async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ id: 101, title: 'Real backend post', content: 'Loaded from intercepted API', status: 'NORMAL' }),
      });
    });
    await page.route(`${API_BASE}/api/posts/101/comments`, async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({ status: 429, body: 'rate limited' });
        return;
      }
      await route.fulfill({ contentType: 'application/json', body: '[]' });
    });
    await page.route(`${API_BASE}/api/posts/101/like`, async (route) => {
      sawLike = route.request().method() === 'POST';
      await route.fulfill({ status: 204 });
    });
    await page.route(`${API_BASE}/api/posts/101/favorite`, async (route) => {
      sawFavorite = route.request().method() === 'POST';
      await route.fulfill({ status: 204 });
    });
    await page.route(`${API_BASE}/api/me`, async (route) => {
      await route.fulfill({ status: 401, body: 'unauthorized' });
    });

    await page.goto('/posts');
    await expect(page.getByText('Real backend post')).toBeVisible();
    expect(sawPosts).toBe(true);

    await page.getByText('Real backend post').click();
    await page.getByRole('button', { name: /点赞/ }).click();
    await page.getByRole('button', { name: /收藏/ }).click();
    expect(sawLike).toBe(true);
    expect(sawFavorite).toBe(true);

    await page.getByLabel('评论内容').fill('@ArchTechLead please reply');
    await page.getByRole('button', { name: '发布评论' }).click();
    await expect(page.getByText(/rate limited|请求过快|频率/i)).toBeVisible();

    await page.goto('/profile');
    await expect(page).toHaveURL(/\/login/);
  });

  test('notification badge updates after mark-read without reload', async ({ page }) => {
    await page.unroute(`${API_BASE}/api/notifications/unread-count`);
    await page.unroute(`${API_BASE}/api/notifications`);
    let unread = 2;
    await page.route(`${API_BASE}/api/notifications/unread-count`, async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify({ count: unread }) });
    });
    await page.route(`${API_BASE}/api/notifications`, async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify([
          { id: 1, type: 'ai.reply.completed', payload: { title: 'AI replied' }, read_at: null, created_at: new Date().toISOString() },
          { id: 2, type: 'comment.created', payload: { title: 'New comment' }, read_at: null, created_at: new Date().toISOString() },
        ]),
      });
    });
    await page.route(`${API_BASE}/api/notifications/1/read`, async (route) => {
      unread = 1;
      await route.fulfill({ status: 204 });
    });

    await page.goto('/');
    await expect(page.getByLabel(/通知/)).toContainText('2');
    await page.getByLabel(/通知/).click();
    await page.getByText('AI replied').click();
    await expect(page.getByLabel(/通知/)).toContainText('1');
  });

  test('SSE failure falls back to ai-status polling and dedupes completed comments', async ({ page }) => {
    let comments = [
      {
        id: 1,
        postId: 201,
        parentId: null,
        content: 'Human comment',
        author: { username: 'human', avatar: '', isAi: false },
        likeCount: 0,
        createdAt: new Date().toISOString(),
      },
    ];
    await page.route(`${API_BASE}/api/posts/201`, async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ id: 201, title: 'SSE post', content: 'Watch AI status', status: 'NORMAL' }),
      });
    });
    await page.route(`${API_BASE}/api/posts/201/comments`, async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify(comments) });
    });
    await page.route(`${API_BASE}/api/posts/201/events`, async (route) => {
      await route.fulfill({ status: 503, body: 'sse unavailable' });
    });
    await page.route(`${API_BASE}/api/posts/201/ai-status`, async (route) => {
      comments = [
        comments[0],
        {
          id: 2,
          postId: 201,
          parentId: null,
          content: 'AI completed once',
          author: { username: 'ArchTechLead', avatar: '', isAi: true, aiAgentId: 1 },
          likeCount: 0,
          createdAt: new Date().toISOString(),
        },
      ];
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ completedCount: 1, runningCount: 0, overallStatus: 'COMPLETED' }),
      });
    });

    await page.goto('/posts/201');
    await expect(page.getByText('AI completed once')).toHaveCount(1);
  });

  test('sanitizes injected markdown and has no serious axe issues on feed/detail', async ({ page }) => {
    await page.route(`${API_BASE}/api/posts`, async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify([{ id: 301, title: 'A11y post', content: '<img src=x onerror="window.__pwned=true">', status: 'NORMAL' }]),
      });
    });
    await page.route(`${API_BASE}/api/posts/301`, async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ id: 301, title: 'A11y post', content: '<script>window.__pwned=true</script>safe', status: 'NORMAL' }),
      });
    });
    await page.route(`${API_BASE}/api/posts/301/comments`, async (route) => {
      await route.fulfill({ contentType: 'application/json', body: '[]' });
    });

    await page.goto('/posts');
    await page.waitForTimeout(700);
    let results = await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa']).analyze();
    expect(results.violations.filter((v) => v.impact === 'serious' || v.impact === 'critical')).toEqual([]);

    await page.getByText('A11y post').click();
    await expect(page.locator('main script')).toHaveCount(0);
    await expect(page.locator('[onerror]')).toHaveCount(0);
    await expect(page.getByText('safe')).toBeVisible();
    await expect(page.evaluate(() => (window as any).__pwned)).resolves.toBeUndefined();
    await page.waitForTimeout(700);
    results = await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa']).analyze();
    expect(results.violations.filter((v) => v.impact === 'serious' || v.impact === 'critical')).toEqual([]);
  });
});
