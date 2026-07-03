import { expect, test } from '@playwright/test';

const TITLE = 'P13 full AI chain post';
const REPLY = 'P13 full AI chain reply from cohere_observer';

async function seedFullChain() {
  const { execFileSync } = await import('node:child_process');
  const sql = `
    DELETE FROM decision_logs WHERE post_id IN (SELECT id FROM posts WHERE title = '${TITLE}');
    DELETE FROM ai_reply_tasks WHERE post_id IN (SELECT id FROM posts WHERE title = '${TITLE}');
    DELETE FROM comments WHERE post_id IN (SELECT id FROM posts WHERE title = '${TITLE}');
    DELETE FROM post_tags WHERE post_id IN (SELECT id FROM posts WHERE title = '${TITLE}');
    DELETE FROM posts WHERE title = '${TITLE}';
    INSERT INTO posts (author_id, title, content, status) VALUES (1, '${TITLE}', 'P13 chain body', 'NORMAL');
    SET @post_id = LAST_INSERT_ID();
    INSERT INTO post_tags (post_id, tag_type, tag_name) VALUES (@post_id, 'topic', 'p13'), (@post_id, 'debate', 'high');
    INSERT INTO decision_logs (post_id, ai_agent_id, trigger_type, willingness_score, threshold_value, decision, reason, hit_tags)
      VALUES (@post_id, 1001, 'AUTO', 0.9200, 0.6000, 'REPLY', 'p13 full-chain accepted', JSON_ARRAY(JSON_OBJECT('type','topic','name','p13')));
    INSERT INTO comments (post_id, user_id, parent_comment_id, comment_type, ai_agent_id, trigger_type, content)
      VALUES (@post_id, NULL, NULL, 'AI', 1001, 'AUTO', '${REPLY}');
    SET @comment_id = LAST_INSERT_ID();
    INSERT INTO ai_reply_tasks (post_id, ai_agent_id, trigger_type, status, comment_id)
      VALUES (@post_id, 1001, 'AUTO', 'SUCCESS', @comment_id);
    SELECT @post_id AS post_id;
  `.replace(/\n/g, ' ');
  const out = execFileSync('docker', ['compose', 'exec', '-T', 'mysql', 'sh', '-lc', `mysql -uroot -pai_forum_root ai_forum -N -e "${sql}"`], {
    cwd: '..',
    encoding: 'utf8',
  });
  return Number(out.trim().split(/\s+/).pop());
}

async function loginWeb(page: import('@playwright/test').Page) {
  await page.goto('/login');
  await page.getByLabel('用户名').fill('admin');
  await page.getByRole('textbox', { name: '密码' }).fill('admin123');
  await page.getByRole('button', { name: /登\s*录/ }).click();
  await expect(page).toHaveURL('http://localhost:5173/');
}

test.describe('P13 live full AI chain integration', () => {
  test('web shows completed AI reply and admin shows decision breakdown', async ({ page }) => {
    const postId = await seedFullChain();

    await loginWeb(page);
    await page.goto(`/posts/${postId}`);
    await expect(page.getByRole('heading', { name: TITLE })).toBeVisible();
    await expect(page.getByText(REPLY)).toBeVisible();
    await expect(page.getByText('check已完成')).toBeVisible();

    await page.goto('http://127.0.0.1:5174/login');
    await page.getByLabel('用户名').fill('admin');
    await page.getByRole('textbox', { name: '密码' }).fill('admin123');
    await page.getByText(/登\s*录/).click();
    await expect(page.getByRole('heading', { name: /Dashboard|概览/ })).toBeVisible();
    const hasDecision = await page.evaluate(async () => {
      const token = localStorage.getItem('ai_forum_admin_token');
      const response = await fetch('http://127.0.0.1:19091/api/admin/decision-logs', {
        headers: { Authorization: `Bearer ${token}` },
      });
      const text = await response.text();
      if (!response.ok) throw new Error(`${response.status}: ${text}`);
      const rows = JSON.parse(text);
      return rows.some((row: any) => row.reason === 'p13 full-chain accepted');
    });
    expect(hasDecision).toBe(true);
    await page.getByRole('link', { name: /AI 决策/ }).click();
    await expect(page.getByText('cohere_observer').first()).toBeVisible();
    await page.getByRole('button', { name: /decision detail/ }).first().click();
    await expect(page.getByRole('dialog')).toContainText('0.92 / 0.60');
    await expect(page.getByRole('dialog')).toContainText('topic:p13');
    await expect(page.getByRole('dialog')).toContainText('p13 full-chain accepted');
  });
});
