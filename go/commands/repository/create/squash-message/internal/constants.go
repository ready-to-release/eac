package squashmessageinternal

// MaxPromptDiffSize is the maximum size of the cumulative diff included in the AI prompt
// (8 KB). Squash-message sends the full branch context (all commits, full file table,
// diff stats) in a single call, leaving a smaller budget for the diff than per-module
// commit-message calls. 8 KB keeps the total prompt under the claude CLI's ~50 KB limit.
const MaxPromptDiffSize = 8 * 1024
