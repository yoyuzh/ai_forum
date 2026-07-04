# AI Agent Chat assistant-ui Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Clicking an AI role in `/agents` opens `/agents/:agentId/chat`, a persisted one-to-one chat powered by the selected AI persona and rendered with `assistant-ui`.

**Architecture:** Add a focused `backend/internal/ai/chat` package with MySQL-backed sessions/messages and synchronous model generation. Expose two authenticated REST routes through existing router/bootstrap wiring. Extend the existing web API client and add one React route/page; no raw page-level fetches.

**Tech Stack:** Go, sqlx, MySQL migrations, existing `auth` and `modelclient`, React 18, Vite, React Router, TanStack Query, `@assistant-ui/react`.

---

### Task 1: Backend Chat Package

**Files:**
- Create: `backend/internal/ai/chat/chat.go`
- Create: `backend/internal/ai/chat/chat_test.go`
- Create: `backend/migrations/000020_ai_chat_sessions.up.sql`
- Create: `backend/migrations/000020_ai_chat_sessions.down.sql`

- [x] Write failing tests for get-or-create session, ordered messages, empty input rejection, and user+assistant persistence.
- [x] Run `rtk go test ./backend/internal/ai/chat` and confirm it fails because the package does not exist or methods are missing.
- [x] Implement `Store`, `Service`, request/response types, and a small HTTP `Handler`.
- [x] Reuse `modelclient.Client`; build the prompt from agent `system_prompt` plus chat history and latest user message.
- [x] Run `rtk go test ./backend/internal/ai/chat` and confirm it passes.

### Task 2: Backend Route Wiring

**Files:**
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_test.go`
- Modify: `backend/internal/bootstrap/bootstrap.go`

- [x] Write failing router tests for `GET /api/agents/{agentId}/chat` and `POST /api/agents/{agentId}/chat/messages`.
- [x] Run `rtk go test ./backend/internal/router` and confirm the chat routes fail before registration.
- [x] Add `GetAgentChat` and `SendAgentChatMessage` fields to `router.BusinessRoutes` and register the two routes.
- [x] Wire chat handler in `bootstrap.NewAPIServer`, wrapping both routes with existing JWT middleware.
- [x] Run `rtk go test ./backend/internal/router ./backend/internal/bootstrap`.

### Task 3: Web API Contract

**Files:**
- Modify: `web/package.json`
- Modify: `web/package-lock.json`
- Modify: `web/src/api/types.ts`
- Modify: `web/src/api/mockClient.ts`
- Modify: `web/src/api/realClient.ts`
- Create: `web/src/hooks/useAgentChat.ts`

- [x] Install `@assistant-ui/react` in `web`.
- [x] Add `AIChatSession`, `AIChatMessage`, and `ApiClient.chat`.
- [x] Implement mock chat persistence in the mock client.
- [x] Implement real client calls to the two backend routes.
- [x] Add a TanStack Query hook for loading and sending messages.
- [x] Run `rtk npm run build` in `web` and fix type errors.

### Task 4: Web Route And Page

**Files:**
- Modify: `web/src/App.tsx`
- Modify: `web/src/components/cards/AIAgentCard.tsx`
- Create: `web/src/pages/AgentChatPage.tsx`

- [x] Make each agent card link to `/agents/:agentId/chat` without changing its visual structure.
- [x] Add the route under the existing app shell.
- [x] Build `AgentChatPage` with `assistant-ui` components/runtime and the existing Cohere/Synthetica tokens.
- [x] Keep unknown-agent, loading, empty, sending, and failed-generation states visible.
- [x] Run `rtk npm run build` in `web`.

### Task 5: Final Verification

**Files:**
- All touched backend and web files.

- [x] Run targeted Go tests: `rtk go test ./backend/internal/ai/chat ./backend/internal/router ./backend/internal/bootstrap`.
- [x] Run frontend build: `cd web && rtk npm run build`.
- [x] If feasible, start the web dev server and check `/agents` and `/agents/1001/chat` render.
- [x] Report skipped scope: streaming, multi-session history, retry UI, attachments.
