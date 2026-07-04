import { expect, test } from '@playwright/test';

const DB = 'mysql -uroot -pai_forum_root ai_forum';
const TITLE_ONE = 'P13 generated notification one';
const TITLE_TWO = 'P13 generated notification two';

async function mysql(sql: string) {
  const escaped = sql.replace(/"/g, '\\"').replace(/\$/g, '\\$');
  return test.step(`mysql: ${sql.split('\n')[0]}`, async () => {
    const { execFileSync } = await import('node:child_process');
    return execFileSync('docker', ['compose', 'exec', '-T', 'mysql', 'sh', '-lc', `${DB} -e "${escaped}"`], {
      cwd: '..',
      encoding: 'utf8',
    });
  });
}

test.describe('P13 notification contract smoke', () => {
  test('generated notifications appear and unread count updates after mark-one and read-all', async ({ page }) => {
    await mysql(`
      DELETE FROM notifications WHERE recipient_id = 1;
      INSERT INTO notifications (recipient_id, type, payload) VALUES
        (1, 'p13.contract', JSON_OBJECT('title', '${TITLE_ONE}', 'body', 'mark one')),
        (1, 'p13.contract', JSON_OBJECT('title', '${TITLE_TWO}', 'body', 'mark all'));
    `);

    await page.goto('/login');
    await page.getByLabel('用户名').fill('admin');
    await page.getByRole('textbox', { name: '密码' }).fill('admin123');
    await page.getByRole('button', { name: /登\s*录/ }).click();
    await expect(page.getByLabel(/通知/)).toContainText('2');

    await page.getByLabel(/通知/).click();
    await expect(page.getByText(TITLE_ONE)).toBeVisible();
    await page.getByText(TITLE_ONE).click();
    await expect(page.getByLabel(/通知/)).toContainText('1');

    await page.getByText('全部已读').click();
    await expect(page.getByLabel(/通知/)).not.toContainText('1');
  });
});
