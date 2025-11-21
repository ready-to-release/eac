Generate a module section using this exact format:

```
src-commands
------------
src-commands: refactor: simplify commit message generation

Removed template variable embedding from generation logic and updated
prompts to use direct format instructions. Simplified assembly code to
use blank line separators instead of dashes between module sections.
```

Your output must:
- Line 1: Module name only (e.g., `src-commands`, `contracts`, `docs`)
- Line 2: Exactly 12 dashes: `------------`
- Line 3: `<module>: <type>: <description>` (max 72 chars, no period)
- Line 4: Empty
- Lines 5+: Body text (2-4 sentences, wrapped at 72 chars)

Output only the module section. No markdown fences, no extra text.
