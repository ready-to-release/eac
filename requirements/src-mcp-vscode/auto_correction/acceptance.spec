# Auto Correction

> **Feature ID**: src_mcp_vscode_auto_correction
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: src-mcp-vscode
> **Tags**: mcp, vscode, correction, formatting

## User Story

* As a commit generation system
* I want to automatically fix common formatting issues in generated commits
* So that all commits pass validation without manual intervention

## Acceptance Criteria

* Adds blank lines before headings (MD022)
* Ensures file ends with single newline (MD047)
* Removes trailing whitespace from lines
* Fixes multiple consecutive blank lines
* Preserves code block content without modification
* Preserves table formatting
* Returns corrected commit message
* Does not alter semantic content

## Acceptance Tests

### AC1: Adds blank lines before headings (MD022)
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, correction, markdown

* Parse commit message for heading markers
* Detect headings without preceding blank line
* Insert blank line before each affected heading
* Skip blank line before first heading
* Preserve all other content
* Return corrected message

### AC2: Ensures file ends with single newline (MD047)
**Validated by**: behavior.feature -> @ac2 scenarios

Tags: critical, correction, markdown

* Check last character of commit message
* If missing newline, add single newline
* If multiple newlines, reduce to single newline
* Ensure exactly one newline at end
* Return corrected message

### AC3: Removes trailing whitespace from lines
**Validated by**: behavior.feature -> @ac3 scenarios

Tags: correction, formatting

* Parse commit message line by line
* Detect trailing spaces or tabs
* Remove all trailing whitespace
* Preserve leading indentation
* Preserve line content
* Return corrected message

### AC4: Fixes multiple consecutive blank lines
**Validated by**: behavior.feature -> @ac4 scenarios

Tags: correction, formatting

* Scan for consecutive blank lines
* Reduce multiple blank lines to single blank line
* Preserve intentional single blank lines
* Do not affect code blocks
* Return corrected message

### AC5: Preserves code block content
**Validated by**: behavior.feature -> @ac5 scenarios

Tags: critical, correction, preservation

* Identify code block boundaries (``` markers)
* Skip correction inside code blocks
* Allow trailing whitespace in code blocks
* Allow multiple blank lines in code blocks
* Preserve exact formatting in code blocks
* Apply corrections outside code blocks

### AC6: Preserves table formatting
**Validated by**: behavior.feature -> @ac6 scenarios

Tags: correction, preservation

* Identify markdown tables (| delimiters)
* Skip line length corrections for tables
* Preserve table structure
* Allow necessary spacing in tables
* Apply corrections to non-table content

### AC7: Returns corrected commit message
**Validated by**: behavior.feature -> @ac7 scenarios

Tags: critical, correction

* Apply all correction rules
* Build corrected message string
* Return complete corrected message
* Ensure message is valid markdown
* Preserve original semantic meaning

### AC8: Does not alter semantic content
**Validated by**: behavior.feature -> @ac8 scenarios

Tags: critical, correction, preservation

* Verify text content unchanged
* Verify headings unchanged (except blank lines)
* Verify file lists unchanged
* Verify module names unchanged
* Only formatting is modified
* Semantic information preserved
