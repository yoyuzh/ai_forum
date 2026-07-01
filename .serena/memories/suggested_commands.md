# Suggested Commands

- Shell commands on this machine should be prefixed with `rtk` per `/Users/mac/.codex/RTK.md`.
- Inspect tree: `rtk find . -maxdepth 4 -print | sort` or `rtk rg --files`.
- Git state: `rtk git status --short`.
- Backend placeholders currently avoid dependencies; once implementation begins, use `rtk go test ./...` from `backend/`.
- Frontend placeholders currently avoid package manifests; once package setup exists, run the project-specific npm/pnpm scripts from `web/` and `admin/`.