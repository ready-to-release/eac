# Modes

Default behavior

- Start in the minimal mode that can satisfy the request safely and efficiently.
- Escalate modes only when necessary per the decision tree below.

Decision Tree (summary)

1. [CHAT] — Greetings, small talk, quick factual questions.
2. [ADVANCED_CHAT] — Read‑only repo analysis/advice; may read files, no edits.
3. [RUN_VERIFY] — Run app/tests or short safe commands; no edits.
4. [FAST_CODE] — Truly trivial change (1–3 steps, single file), no extra investigation.
5. [CODE] — Anything non‑trivial: multi‑file, tests, or investigation needed.
6. [SETUP] — Install/build/configure/fix environment; no significant app code.
7. [NICHE] — Specialized forensics/reverse‑engineering tasks only.

Switching Rules

- [CHAT]/[ADVANCED_CHAT]: never switch mid‑step; finish answer.
- [FAST_CODE]: switch to [CODE] if not done in 3 steps.
- [RUN_VERIFY]: switch to [CODE] if not done in 3 steps.
- [SETUP]/[NICHE]: may switch to [CODE] if significant code changes become necessary.
