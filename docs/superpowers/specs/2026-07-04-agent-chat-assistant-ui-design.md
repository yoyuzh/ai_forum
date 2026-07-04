# AI Agent Chat With assistant-ui

## Goal

Clicking an AI role in the AI role plaza opens a real one-to-one chat with that agent. The chat UI uses `assistant-ui`, while API calls remain inside `web/src/api`.

## Scope

- Add a route from `/agents` to `/agents/:agentId/chat`.
- Add backend persistence for AI chat sessions and messages.
- Generate replies with the existing AI persona/model-client path.
- Support mock and real web API modes with the same client shape.

Out of scope for this slice: multi-session history, history search, streaming output, message editing, attachments, and admin chat management.

## Backend Design

Add `backend/internal/ai/chat` as the owner of one-to-one AI chat. This keeps AI chat out of forum post/comment services and preserves the modular monolith boundary.

Add two migrations:

- `ai_chat_sessions`: id, user_id, ai_agent_id, title, created_at, updated_at, with a unique active session key for `(user_id, ai_agent_id)` in this v1 flow.
- `ai_chat_messages`: id, session_id, role (`user` or `assistant`), content, error_message nullable, created_at.

Add public authenticated endpoints:

- `GET /api/agents/{agentId}/chat`: find or create the current user's session for the agent and return session metadata plus ordered messages.
- `POST /api/agents/{agentId}/chat/messages`: persist the user message, call the model using the selected agent persona, persist the assistant reply, and return the created assistant message plus the saved user message.

The message endpoint is synchronous for v1 because the user is actively waiting in a chat UI. It must not publish RabbitMQ messages inside a transaction. If model generation fails, the user message remains saved and the API returns a structured error; it must not fabricate an assistant success message.

## Frontend Design

Update `AIAgentCard` so the card links to `/agents/:agentId/chat`.

Add `AgentChatPage.tsx` under the existing app shell. It fetches the agent and chat session, then renders an `assistant-ui` thread styled with existing Cohere/Synthetica tokens and the supplied Stitch chat direction.

Extend the web API contract in `web/src/api/types.ts`:

- `chat.get(agentId)`
- `chat.sendMessage(agentId, content)`

Implement both in `mockClient.ts` and `realClient.ts`. Page code calls hooks/client methods only; it does not issue raw fetches.

## Error Handling

- Unknown agent: show a normal not-found state.
- Empty message: blocked client-side and validated server-side.
- Model failure: keep the user's message visible, show a failed assistant state; retry is explicitly outside this slice.
- Auth failure in real mode: reuse existing HTTP client behavior.

## Testing

Follow red-green TDD for production behavior:

- Backend route test for the two chat endpoints.
- Backend chat repository/service test proving get-or-create session and user+assistant message persistence.
- Frontend test or build-time check proving agent cards link to chat and API client shape compiles.

Final verification should run the narrow Go tests touched by the chat package/router and `npm run build` in `web`.
