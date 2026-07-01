# Task Completion

- Before claiming skeleton or docs work complete: verify required files exist with `rtk find ...`, check `rtk git status --short`, and ensure existing docs were not moved/deleted.
- For Go implementation tasks after real code exists: run `rtk gofmt -w` on touched Go files and `rtk go test ./...` from `backend/`.
- For frontend implementation tasks after package manifests exist: run the relevant lint/typecheck/build scripts declared in `web/package.json` or `admin/package.json`.
- After onboarding memory updates, user can sanity-check references with `serena memories check` from the project root.