import { expect, test } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const repoRoot = path.resolve(__dirname, '..', '..');

test.describe('P13 reports scope guard', () => {
  test('reports are documented out-of-scope and not half-wired in admin', () => {
    const readme = fs.readFileSync(path.join(repoRoot, 'e2e', 'README.md'), 'utf8');
    expect(readme).toMatch(/reports.*out-of-scope/i);

    const app = fs.readFileSync(path.join(repoRoot, 'admin', 'src', 'App.tsx'), 'utf8');
    const sideNav = fs.readFileSync(path.join(repoRoot, 'admin', 'src', 'components', 'layout', 'SideNav.tsx'), 'utf8');

    expect(app).not.toMatch(/path=["']\/reports["']/);
    expect(sideNav).not.toMatch(/\/reports\b/);
  });
});
