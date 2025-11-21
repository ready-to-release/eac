EXAMPLE OUTPUT FORMAT (your output must match this structure exactly):

refactor(multi-module): simplify commit message prompts

Auditor-Summary: Removed template embedding for clearer AI instructions.

This commit simplifies the prompts by removing YAML template variables
and using direct format instructions. Updated both top-level and module
prompts to show concrete examples rather than abstract schemas.

Changes: 8 files, +95 insertions, -150 deletions

---

Generate a commit message following the format shown above. Your output must have:

1. FIRST LINE: Conventional commit header format: <type>(<scope>): <summary>
   Valid types: feat, fix, refactor, docs, chore, test, perf, style
   Example: refactor(multi-module): simplify commit message prompts

2. BLANK LINE

3. Auditor-Summary line (one sentence summary)
   Format: Auditor-Summary: <sentence>

4. BLANK LINE

5. Body (2-4 sentences, lines wrapped at 72 characters)

6. BLANK LINE

7. Changes line with statistics
   Format: Changes: N files, +X insertions, -Y deletions

CRITICAL: Your FIRST line must be the commit header (<type>(<scope>): <summary>).
Do NOT start with "Auditor-Summary" or any other text.
Output ONLY the commit message. NO code fences. NO explanatory text.
