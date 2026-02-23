package squashmessageinternal

// MaxPromptDiffSize is the maximum size of the cumulative diff included in the AI prompt
// (4 KB). The squash-message prompt already includes commit history, a file table, and
// diff stats in a single call. Keeping the diff small leaves room for those sections
// without exceeding the claude CLI input limit (empirically around 20 KB total).
const MaxPromptDiffSize = 4 * 1024

// MaxPromptCommits is the maximum number of commits shown in the commit history section.
// Beyond this limit a "... and N more commits" note is appended. Commit bodies can be
// several hundred bytes each; 15 commits keeps that section under ~4 KB.
const MaxPromptCommits = 15

// MaxPromptFiles is the maximum number of file rows shown in the changed-files table.
// Each row is roughly 80-120 bytes; 30 rows keeps the table under ~4 KB.
const MaxPromptFiles = 30

// MaxPromptDiffStatsSize is the maximum size of the diff stats block in bytes (2 KB).
// The stats block has one line per changed file; for large branches this can be several KB.
const MaxPromptDiffStatsSize = 2 * 1024
