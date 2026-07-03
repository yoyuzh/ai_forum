# AI Forum E2E Pre-Flight

Run from the repository root.

## Local Stack

1. Confirm ports `3306`, `6379`, `5672`, `9200`, `19091`, `5173`, and `5174` are free.
2. Start the live stack:
   ```bash
   rtk ./scripts/dev-up.sh
   ```
3. Run Playwright:
   ```bash
   rtk npm run test
   ```
4. Stop the stack:
   ```bash
   rtk ./scripts/dev-down.sh
   ```

## Environment

- Backend compose defaults set `MYSQL_*`, `REDIS_ADDR`, `RABBITMQ_URL`, `ES_ADDRESSES`, `JWT_SECRET`, `INTERNAL_API_TOKEN`, and `AI_API_KEY`.
- Web and admin run with `VITE_API_MODE=real` and `VITE_API_BASE_URL=http://127.0.0.1:19091`.
- `api-server` is not host-exposed; public API access goes through nginx on `19091`.
- `/internal/**` must return `404` through nginx.

## Data

- `scripts/dev-up.sh` runs `make migrate-up` against the compose MySQL database.
- Dev seed data comes from backend migrations, including the admin user and AI agents.
- Search rebuild smoke uses the P9 rebuild path to repopulate Elasticsearch from MySQL source rows:
  ```bash
  rtk cd backend && env MYSQL_HOST=127.0.0.1 MYSQL_PORT=3306 MYSQL_USERNAME=root MYSQL_PASSWORD=ai_forum_root MYSQL_DATABASE=ai_forum ES_ADDRESSES=http://127.0.0.1:9200 go run ./cmd/search-rebuild
  ```

## Performance

- Lighthouse LCP/CLS runs against the already-started local compose/Vite stack with Chrome for Testing and provided throttling:
  ```bash
  rtk cd e2e && CHROME_PATH="$(node -e "const { chromium } = require('@playwright/test'); console.log(chromium.executablePath())")" npx --yes lighthouse http://127.0.0.1:5173/posts --only-audits=largest-contentful-paint,cumulative-layout-shift --throttling-method=provided --output=json --output-path=/tmp/p13-lh-posts.json --chrome-flags="--headless=new --no-sandbox --disable-gpu --disable-dev-shm-usage --disable-extensions" --quiet
  ```
- Repeat for `http://127.0.0.1:5173/posts/1` and `http://127.0.0.1:5174/login`; require LCP < 2500ms and CLS < 0.1.

## Scope Guards

- Reports and user report workflow are v1 out-of-scope until a later OpenSpec phase owns `/admin/reports`, user reporting routes, and backend moderation behavior.

## Accepted Advisories

- `web`: `npm audit --audit-level=critical` reports the Vite/esbuild dev-server advisory (`GHSA-67mh-4wv8-2f99`) as moderate/high. It is accepted for P13 because the affected path is the local Vite development server used by compose e2e, not production serving, and the available audit fix requires a breaking Vite major upgrade.
- `admin`: `npm audit --audit-level=critical` reports the same Vite/esbuild development-server advisory plus `path-to-regexp` advisories through `@refinedev/antd`/`@ant-design/pro-layout` (`GHSA-j3q9-mxjg-w52f`, `GHSA-27v5-c462-wpq7`). They are accepted for P13 because the fix path requires breaking Refine/Ant Design upgrades and these routes are exercised only in the admin SPA/dev-server e2e surface.
