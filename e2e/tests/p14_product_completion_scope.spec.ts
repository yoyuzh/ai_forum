import { expect, test } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const repoRoot = path.resolve(__dirname, '..', '..');

function read(...parts: string[]): string {
  return fs.readFileSync(path.join(repoRoot, ...parts), 'utf8');
}

test.describe('P14 product completion source guards', () => {
  test('reports and fake SSO are not visible routes or login actions', () => {
    const adminApp = read('admin', 'src', 'App.tsx');
    const sideNav = read('admin', 'src', 'components', 'layout', 'SideNav.tsx');
    const loginPage = read('web', 'src', 'pages', 'LoginPage.tsx');
    const webSrc = fs.readdirSync(path.join(repoRoot, 'web', 'src'), { recursive: true })
      .filter((p) => typeof p === 'string' && /\.[tj]sx?$/.test(p))
      .map((p) => read('web', 'src', p as string))
      .join('\n');

    expect(adminApp).not.toMatch(/path=["']\/reports["']/);
    expect(sideNav).not.toMatch(/\/reports\b/i);
    expect(webSrc).not.toMatch(/\breport\b/i);
    expect(loginPage).not.toMatch(/SSO|OR CONTINUE WITH|Google|GitHub|社交登录/);
  });

  test('real mode clients do not fall back to mock or hard-coded empty AI data', () => {
    const webRealClient = read('web', 'src', 'api', 'realClient.ts');
    const adminClient = read('admin', 'src', 'api', 'client.ts');
    const adminDataProvider = read('admin', 'src', 'api', 'dataProvider.ts');

    expect(webRealClient).not.toMatch(/list:\s*async\s*\([^)]*\)\s*:\s*Promise<AI[A-Za-z]+>\s*=>\s*\[\]/);
    expect(webRealClient).not.toContain('updates as UserProfile');
    expect(adminClient).not.toContain('dashboard: mockApi.dashboard');
    expect(adminDataProvider).not.toMatch(/not implemented/i);
  });
});
