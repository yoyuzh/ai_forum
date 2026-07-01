# P12 Tasks — Admin Refine + decision-log visualization

## 1. Data + auth + RBAC
- [ ] 1.1 `admin/src/api/dataProvider.ts`: Refine REST dataProvider against backend
- [ ] 1.2 `admin/src/api/authProvider.ts`: login/JWT wiring; 401→login
- [ ] 1.3 `admin/src/providers/accessControlProvider.ts`: RBAC visibility from backend permissions (frontend-only)
- [ ] 1.4 Register resources in `App.tsx` (Users/Posts/Comments/AI Agents/AI Tasks/Tags/Preferences)

## 2. Operational screens
- [ ] 2.1 Users / Posts / Comments list+show+edit (dense, operator-focused)
- [ ] 2.2 AI Agents screen: `replyThreshold`/`activityLevel`/`allowAutoReply`/`allowMention`/`allowFollowup` inline-edit; cross-link to recent decisions
- [ ] 2.3 AI Tasks screen: retry/terminate/mark-processed
- [ ] 2.4 Tags/Preferences management

## 3. Decision-log explorer (differentiator)
- [ ] 3.1 `DecisionDetailDrawer`: agent, trigger, willingness vs threshold, hitTags, decision, reason, fallback flag, link to task/comment
- [ ] 3.2 `WillingnessGauge`: score-vs-threshold, Cohere blue/coral, always paired text label (color never sole signal)
- [ ] 3.3 `HitTagsViewer`, `SkipReasonBlock` (below-threshold/fallback/rate-limited/config-disabled)
- [ ] 3.4 `AgentDecisionBreakdown`: reply-rate, avg willingness, fallback rate per agent
- [ ] 3.5 Register decision-log resource (read-heavy; edits go through agent config)

## 4. Design system + a11y
- [ ] 4.1 Align admin fonts to Cohere (Unica77/Space Grotesk/CohereText/CohereMono) in tailwind; flag any licensing gap
- [ ] 4.2 axe-core scan on admin dashboard + decision log (front-loaded)
- [ ] 4.3 WCAG-AA contrast scan on Cohere text/background pairings

## 5. Tests / E2E
- [ ] 5.1 Admin login + RBAC-gated CRUD
- [ ] 5.2 Decision-log explorer renders full breakdown from P6 `decision_logs` (requires P6 backend)
- [ ] 5.3 RBAC denial E2E: denied action server-side → 403 (backend authoritative)

## 6. Verification
- [ ] 6.1 `npm run lint` (--max-warnings 0) and `npm run build` green; admin runs on 5174
- [ ] 6.2 blockedBy P6 (decision_logs/ai_agents) and P4 (RBAC/users/posts)
