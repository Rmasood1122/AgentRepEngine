# HIREINSTEIN — CONTINUATION PROMPT
# Paste this entire document into a new Claude chat to resume exactly where we left off.
# Date: February 23, 2026

---

## PROJECT CONTEXT

I'm building **Hireinstein** — an AI-powered recruitment/visibility tool that tracks how AI models (OpenAI GPT, Anthropic Claude, Google Gemini, xAI) rank and mention domains in real-time. Users enter a domain and query, and get a composite visibility score showing how AI perceives their brand.

**Tech Stack:** Next.js App Router + Express API + Prisma + PostgreSQL + TypeScript
**Repository:** C:\Users\rmaso\Hireinstein
**Local Setup:** API on localhost:3001, Frontend on localhost:3000, PostgreSQL on localhost:5432

---

## WHAT'S BEEN COMPLETED

### Development Environment (DONE)
- Node.js, PostgreSQL installed and configured
- Database "hireinstein" created, Prisma schema pushed
- API server running with auth (signup/login/logout) and ranking endpoints
- Frontend running with basic pages
- OpenAI and Anthropic API keys configured and tested
- Critical dotenv module loading order bug diagnosed and fixed (requires NODE_OPTIONS preload)

### Starting Servers
**Terminal 1 (API):**
```powershell
cd C:\Users\rmaso\Hireinstein\apps\api
$env:NODE_OPTIONS="-r dotenv/config"
$env:DOTENV_CONFIG_PATH="../../.env.local"
npm run dev
```

**Terminal 2 (Frontend):**
```powershell
cd C:\Users\rmaso\Hireinstein\apps\web
$env:NODE_OPTIONS=""
$env:DOTENV_CONFIG_PATH=""
npm run dev
```

---

## UI/UX SPECIFICATION STATUS

We followed a strict 5-step process to build an enterprise-grade UI specification that is 11x better than competitors (RankGPT and UANDI). Each step was audited at FAANG/Stanford++ level before locking.

### Step 1: Interface Philosophy — LOCKED + SAVED
- File: `C:\Users\rmaso\Hireinstein\docs\ui\STEP_1_INTERFACE_PHILOSOPHY.md`
- "Structured Intelligence 2.0" — 6 principles: authority through restraint, intelligence demonstrated early, mathematical hierarchy, progressive commitment, executive tone, controlled signal density

### Step 2: Token System v2 — LOCKED + SAVED
- File: `C:\Users\rmaso\Hireinstein\docs\ui\STEP_2_TOKEN_SYSTEM.md`
- Typography: Satoshi (headings), DM Sans (body), JetBrains Mono (dynamic metrics only)
- Scale: 64/36/24/16/14/12px
- Spacing: 8px base grid (12px exception only for compact control vertical padding)
- Color: dark instrument theme, single accent (#10b981 emerald green), no second accent
- Includes: enforcement/governance rules, accessibility acceptance tests, semantic state tokens (overlay, focus-ring, disabled-surface, disabled-text, success, info), reduced motion summary, border token binding, changelog

### Step 3: Plain Text Wireframe v2 — LOCKED + SAVED
- File: `C:\Users\rmaso\Hireinstein\docs\ui\STEP_3_WIREFRAME.md`
- 4 pages: Homepage, Login, Signup, Dashboard
- Governance rules: icon governance, nav taxonomy (4 links + 1 CTA), JetBrains Mono dynamic-only rule, CTA budget (4 max per page), model differentiation (label only, no color)
- Hero badge: low-emphasis (text-muted, no pulse, no accent)
- Pricing: text links not buttons (preserves CTA budget)
- Mobile behavior specified for all pages
- CTA budget table for homepage

### Step 4: Interaction Design v2 — ENTERPRISE AUDIT PASSED, NEEDS TO BE SAVED
- File should go to: `C:\Users\rmaso\Hireinstein\docs\ui\STEP_4_INTERACTION_DESIGN.md`
- **THIS FILE HAS NOT BEEN SAVED YET — it was the last thing we were working on**
- Contains: 9 page-level states, state precedence/cancellation (6 rules), token alias mapping, sections 4A-4L (page load, nav, hero, inputs, submit flow, result reveal, edge states, finding cards, scroll triggers, dashboard, auth, global rules, reduced motion)
- Edge state matrix: auth expired, timeout, 429, 5xx, offline, partial, invalid domain, empty query
- 14 global interaction rules
- The file was generated and downloaded but user needs to run:
```powershell
copy "$HOME\Downloads\STEP_4_INTERACTION_DESIGN_v2.md" "C:\Users\rmaso\Hireinstein\docs\ui\STEP_4_INTERACTION_DESIGN.md"
```
Then verify with:
```powershell
Get-Content "C:\Users\rmaso\Hireinstein\docs\ui\STEP_4_INTERACTION_DESIGN.md" -Tail 25
```

---

## WHAT'S NEXT

### Immediate Next Action
1. Verify Step 4 is saved (user needs to confirm last 25 lines show changelog + guardrail with 10 points)
2. Once Step 4 is verified and locked, proceed to **Step 5: Implementation**

### Step 5: Implementation Plan
Per the original audit document, the build order for Figma/code is:
1. Create color styles first (CSS variables from token system)
2. Create text styles second (typography scale)
3. Create button components third
4. Create card components fourth
5. Build layout last

Since build target is React/Next.js, we skip Figma and go straight to code using the locked specifications. All 4 spec documents serve as the implementation contract.

### Remaining Project Work (from doctor.js roadmap)
- Stripe billing integration (Step 60)
- Rate limiting (Step 56)
- Redis caching (Step 57)
- Sentry error logging (Step 58)
- ESLint + Prettier + Husky (Step 13)
- SEO strategy (Step 10)
- Analytics schema (Step 9)
- Deploy to production

---

## IMPORTANT RULES WE ESTABLISHED

1. **Never move to next step until current step is saved and verified** — always check last 15-25 lines of saved file
2. **Every step gets audited at FAANG/Stanford++ enterprise grade** before locking
3. **All spec files live in** `C:\Users\rmaso\Hireinstein\docs\ui\`
4. **Token system is the source of truth** — no raw hex, no arbitrary spacing, no unlisted radii
5. **Single accent color** (emerald #10b981) — no cyan, no blue, no second accent family
6. **JetBrains Mono only for dynamic metrics** — static marketing claims use DM Sans
7. **4 CTAs max per page** — pricing uses text links, not buttons
8. **Model differentiation by label only** — GPT/CL/GEM/xAI, no color coding

---

## HOW TO REFERENCE THE SPECS

If you need to read any spec file during implementation, the user can show contents with:
```powershell
Get-Content "C:\Users\rmaso\Hireinstein\docs\ui\STEP_1_INTERFACE_PHILOSOPHY.md"
Get-Content "C:\Users\rmaso\Hireinstein\docs\ui\STEP_2_TOKEN_SYSTEM.md"
Get-Content "C:\Users\rmaso\Hireinstein\docs\ui\STEP_3_WIREFRAME.md"
Get-Content "C:\Users\rmaso\Hireinstein\docs\ui\STEP_4_INTERACTION_DESIGN.md"
```

---

Please confirm you understand the full context above, then help me continue from where we left off. The immediate task is to verify Step 4 was saved, and then proceed to Step 5 (Implementation).
