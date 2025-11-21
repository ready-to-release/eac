# Remaining 66 Undefined Steps - Analysis

## BDD Common/Setup Steps (10 steps)
1. the working directory is clean
2. a command that creates a file
3. a test directory exists
4. a command that creates a directory
5. the directory contains a file "README.md"
6. the directory "test-dir" exists
7. the command is run
8. the commands are executed concurrently
9. the command is executed with a timeout of 5 seconds
10. the cleanup command is executed

## Commit-AI Steps (28 steps)
### Setup/Given:
11. a commit message contract
12. a commit message contract with version "0.1.0"
13. a commit message header ending with a period
14. a commit message with an opening code fence but no closing fence
15. a commit message with multiple consecutive blank lines
16. a body text line longer than 72 characters
17. a code block without blank lines before and after
18. a full git diff
19. a full git diff with multiple files
20. a git diff larger than 10 MB
21. no staged changes in git
22. I have staged files affecting 10 modules:
23. followed by a valid commit header "feat(multi-module): add features"
24. followed by module section "src-commands"
25. an Auditor-Summary field
26. module names with edge cases (single char, max length, special patterns)
27. no .r2r directory exists

### Execution/When:
28. I run commit-ai with race detector enabled
29. git diff command fails
30. the command is executed with a timeout of 5 seconds (duplicate - covered in BDD)

### Verification/Then:
31. the context should list all affected modules
32. the context should list the affected module
33. module context is built
34. missing modules are added
35. no module sections are generated
36. module sections are generated in parallel
37. each module section is generated one after another
38. both messages have the same structure
39. performance metrics are logged to stderr
40. the total generation time is less than sequential execution
41. debug files are created for all 5 modules:

## Specs Steps (10 steps)
### Setup/Given:
42. I run the specs create command
43. the contract file does not exist
44. a contract file with invalid YAML
45. the specs directory is not writable

### Verification/Then:
46. it must contain a "Feature:" declaration
47. the AI generates a feature named "user-authentication"
48. the AI generates a feature named "src-commands_user-authentication"
49. the AI generates a feature that would create the same path
50. the AI provider returns output with initialization messages
51. stdout contains error codes like "MISSING_RULE", "INVALID_FEATURE_NAMING"
52. stdout contains provider selection confirmation
53. "out/debug-raw-output.feature" contains raw AI output
54. the contract must include "top_level_heading" section

## AI Provider Steps (5 steps)
55. the AI agent is invoked with the description
56. the AI provider will fail for module "src-core"
57. the AI provider will fail for modules:
58. the AI returns an empty response for module "src-core"
59. AI output wrapped in triple backticks

## Template Steps (7 steps)
60. a template directory with file "../../etc/passwd.tmpl"
61. a template exists at ".r2r/templates/specs/specification.feature"
62. template values with Path="../../../etc"
63. a security violation is detected
64. the output file should contain "# {{ .ProjectName }}"

## Command Behavior Steps (6 steps)
65. a command that handles secrets
66. a command that processes sensitive data
67. a command that requires authentication
68. the standard output contains a usage message
69. they should only need to create a new file "templates/apply/architecture.go"
70. they should only need to create a new file "templates/install/architecture.go"

Total: 70 steps (some duplicates to resolve)
