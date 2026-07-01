# Module Instructions

## Responsibility

Own the user-facing React + TypeScript + Vite application.

## Owns

- User web source under `web/src`.
- User-facing API clients, hooks, routes, stores, styles, and components.

## Must Not

- Do not scatter API calls outside `src/api`.
- Do not scatter SSE logic across pages.
- Do not render user-supplied rich text without DOMPurify or equivalent sanitization.

## Allowed Dependencies

- React, TypeScript, Vite, Ant Design, TanStack Query, Zustand, React Router, React Virtuoso, Tiptap or mentions input, react-markdown, DOMPurify, and small UI utilities chosen by project docs.

## Communication Rules

- API requests live in `src/api`.
- Server state uses TanStack Query.
- Lightweight client state uses Zustand.
- SSE logic lives in dedicated hooks or SSE components.

## Data Rules

- Treat server data as authoritative.
- Do not duplicate durable business state in client stores.
- Sanitize rich-text/Markdown display before rendering.

## Testing Rules

- Add component tests for important user flows and hook tests for API/SSE behavior when implementation starts.

## Notes for Codex

- Long lists such as post feeds and comments should use React Virtuoso.
