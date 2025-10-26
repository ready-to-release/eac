# Recording Demo Videos

This guide explains how to create professional GIF demonstrations of the VSCode extension for documentation.

## Overview

We need three demo GIFs to showcase the extension:

| Demo            | Duration | Target File                          | Size | Status             |
| --------------- | -------- | ------------------------------------ | ---- | ------------------ |
| Quick Start     | ~45s     | `/docs/assets/quick-start-guide.gif` | ~3MB | 📝 Ready to record |
| Git Commit      | ~20s     | `/docs/assets/git-commit-demo.gif`   | ~2MB | 📝 Ready to record |
| MCP Server Test | ~25s     | `/docs/assets/mcp-server-test.gif`   | ~2MB | 📝 Ready to record |

## Recording Tools

### Windows (Recommended)

#### ScreenToGif - Free, all-in-one solution

- Download: [https://www.screentogif.com/](https://www.screentogif.com/)
- Features: Record, edit, annotate, optimize all in one tool
- Perfect for documentation GIFs

### macOS (Recommended)

#### Kap - Free, modern interface

- Install: `brew install --cask kap`
- Features: Plugins, easy export, trimming
- Clean and professional output

### Linux (Recommended)

#### Peek - Simple and effective

- Install: `sudo apt install peek`
- Lightweight and fast
- Good for quick recordings

### Terminal Recording (For MCP Server Test)

#### asciinema - Record terminal as text

- Install: `npm install -g asciinema @asciinema/agg`
- Record: `asciinema rec demo.cast`
- Convert: `agg demo.cast output.gif`
- Smaller file sizes, crisp text

## Environment Setup

### VSCode Configuration

Recommended settings for recording:

```json
{
  "workbench.colorTheme": "Default Dark Modern",
  "editor.fontSize": 14,
  "editor.fontFamily": "JetBrains Mono, Fira Code, monospace",
  "editor.minimap.enabled": false,
  "workbench.activityBar.visible": true,
  "window.zoomLevel": 0,
  "editor.renderWhitespace": "none"
}
```

### Desktop Preparation

Before recording:

1. **Close unnecessary windows**
2. **Hide desktop icons** (if visible)
3. **Disable notifications:**
   - Windows: Focus Assist → Alarms Only
   - macOS: Do Not Disturb
   - Linux: Notification settings
4. **Set display resolution:** 1920x1080 or 1280x720
5. **Close system tray popups**

## Demo 1: Quick Start Guide (45 seconds)

**Target:** `docs/assets/quick-start-guide.gif` (~3MB)

### Storyboard

**Scene 1: Terminal - Initialize (0:00-0:08)**

```bash
./automation/sh-vscode/init.sh
```

- Show initialization running
- Highlight success checkmarks
- Pause on "✓ Initialization Complete!"

**Scene 2: VSCode - Open (0:08-0:12)**

```bash
code .
```

- VSCode opens
- Show project structure in sidebar
- Focus on `.vscode/extensions/` folder

**Scene 3: Press F5 (0:12-0:18)**

- Show keyboard overlay pressing `F5` (if possible)
- Debug dropdown appears
- Extension Development Host window opens
- Highlight window title: `[Extension Development Host]`

**Scene 4: Open Source Control (0:18-0:24)**

- Press `Ctrl+Shift+G`
- Source Control view opens
- Circle/highlight the **robot icon** (🤖) in toolbar

**Scene 5: Click Button (0:24-0:32)**

- Mouse moves to robot icon
- Click button
- Quick Pick menu appears:
  - Git Commit
  - Git Push
  - Git Pull
  - Custom Action
- Select "Git Commit"
- Input box: "Enter a message (optional)"
- Type: "Test commit"
- Press Enter

**Scene 6: Show Result (0:32-0:38)**

- Notification appears:
  ```
  MCP Server Response: Executing action: git-commit with message: Test commit
  ```
- Highlight notification with circle/arrow

**Scene 7: End Screen (0:38-0:45)**

- Text overlay:

  ```
  ✓ Setup Complete!

  4 MCP Servers Ready
  VSCode Extension Working
  MCP Integration Active
  ```

### Recording Tips

- Move cursor slowly and deliberately
- Pause 2-3 seconds on important screens
- Use smooth transitions
- Type at natural but slower pace (~40-60 WPM)

## Demo 2: Git Commit Demo (20 seconds)

**Target:** `docs/assets/git-commit-demo.gif` (~2MB)

### Preparation

```bash
# Make test changes
echo "test" > test.txt
git add test.txt
```

### Storyboard

**Scene 1: Show Modified Files (0:00-0:04)**

- Source Control view open
- Changes section visible
- Show 2-3 modified files with "M" indicator

**Scene 2: Click Robot Button (0:04-0:08)**

- Mouse moves to robot icon
- Brief hover (0.5s)
- Click button
- Quick Pick menu opens

**Scene 3: Select Action (0:08-0:12)**

- "Git Commit" highlighted
- Press Enter or click

**Scene 4: Enter Message (0:12-0:16)**

- Input box appears
- Type: "Test commit for MCP demo"
- Press Enter

**Scene 5: Show Result (0:16-0:20)**

- Success notification appears
- Green checkmark or success indicator
- Hold for 3 seconds

### Recording Tips

- Focus on the button and menu
- Keep actions smooth and clear
- Show the notification prominently

## Demo 3: MCP Server Test (25 seconds)

**Target:** `docs/assets/mcp-server-test.gif` (~2MB)

### Using asciinema (Recommended)

This creates cleaner terminal recordings:

```bash
cd src/mcp/vscode

# Start recording
asciinema rec demo.cast

# Type these commands (have them ready to copy-paste):
go run .

{"jsonrpc":"2.0","id":1,"method":"initialize"}

{"jsonrpc":"2.0","id":2,"method":"tools/list"}

{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vscode-action","arguments":{"action":"git-commit","message":"test"}}}

# Press Ctrl+C to exit server
# Press Ctrl+D to stop recording

# Convert to GIF
agg demo.cast ../../docs/assets/mcp-server-test.gif

# Clean up
rm demo.cast
```

### Storyboard

**Scene 1: Start Server (0:00-0:05)**

```bash
cd src/mcp/vscode
go run .
```

- Terminal with blinking cursor

**Scene 2: Initialize (0:05-0:10)**

- Type/paste JSON request
- Press Enter
- Server responds with initialization data

**Scene 3: List Tools (0:10-0:15)**

- Type/paste tools/list request
- Press Enter
- Server responds with available tools

**Scene 4: Call Tool (0:15-0:20)**

- Type/paste tools/call request
- Press Enter
- Server responds with action result

**Scene 5: Exit (0:20-0:25)**

- Press `Ctrl+C`
- Server exits cleanly
- Return to prompt

### JSON Commands (Copy-Paste Ready)

```json
{"jsonrpc":"2.0","id":1,"method":"initialize"}
```

```json
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
```

```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vscode-action","arguments":{"action":"git-commit","message":"test"}}}
```

## Recording Process

### Using the Helper Script

The easiest way:

```bash
# This guides you through recording
./automation/sh-vscode/record-demo.sh
```

The script will:

1. Check prerequisites
2. Run initialization
3. Pause for you to start recorder
4. Guide through each step
5. Tell you when to stop

### Manual Recording Steps

**1. Prepare Environment:**

```bash
# Clean and initialize
./automation/sh-vscode/clean.sh
./automation/sh-vscode/init.sh

# Clear terminal
clear
```

**2. Configure Recorder:**

- Resolution: 1280x720 or 1920x1080
- Frame rate: 10-15 FPS
- Colors: 256 (for smaller file size)

**3. Record:**

- Start recorder
- Follow storyboard step-by-step
- Move slowly
- Pause at key moments
- Stop recorder

**4. Edit (using ScreenToGif/Kap/Peek):**

- Trim unnecessary frames at start/end
- Remove mistakes or long pauses
- Add annotations (circles, arrows)
- Add text overlays (optional)

**5. Optimize:**

```bash
# Install gifsicle
# Windows: scoop install gifsicle
# macOS: brew install gifsicle
# Linux: sudo apt install gifsicle

# Basic optimization
gifsicle -O3 --colors 256 input.gif -o output.gif

# With scaling
gifsicle -O3 --scale 0.8 --colors 256 input.gif -o output.gif

# Reduce frame rate
gifsicle --delay=15 input.gif -o output.gif  # ~7 FPS
```

**6. Save:**

```bash
# Save to docs/assets/
mv output.gif docs/assets/quick-start-guide.gif
```

## Target Specifications

### File Requirements

```yaml
Quick Start Demo:
  Duration: ~45 seconds
  Resolution: 1280x720
  Frame Rate: 10-15 FPS
  Colors: 256
  File Size: < 3MB
  Loop: Infinite

Git Commit Demo:
  Duration: ~20 seconds
  Resolution: 1280x720
  Frame Rate: 10-15 FPS
  Colors: 256
  File Size: < 2MB
  Loop: Infinite

MCP Server Test:
  Duration: ~25 seconds
  Resolution: 1000x600 (terminal only)
  Frame Rate: 10-15 FPS
  Colors: 256
  File Size: < 2MB
  Loop: Infinite
```

## Optimization Tips

### Reduce File Size

**Reduce colors:**

```bash
gifsicle --colors 128 input.gif -o output.gif
```

**Reduce frame rate:**

```bash
# 10 FPS
gifsicle --delay=10 input.gif -o output.gif

# 7 FPS
gifsicle --delay=15 input.gif -o output.gif
```

**Scale down:**

```bash
# 80% of original
gifsicle --scale 0.8 input.gif -o output.gif

# 75% of original
gifsicle --scale 0.75 input.gif -o output.gif
```

**Combine all:**

```bash
gifsicle -O3 --scale 0.75 --colors 128 --delay=15 input.gif -o output.gif
```

### Ensure Looping

```bash
# Set infinite loop
gifsicle --loopcount=0 input.gif -o output.gif
```

### Online Optimization

If gifsicle is not available:

- [ezgif.com/optimize](https://ezgif.com/optimize) - Web-based optimizer
- [gifcompressor.com](https://gifcompressor.com/) - Simple compression

## Best Practices

### Do's

✓ **Slow down** - Move cursor slowly and deliberately
✓ **Pause** - Hold on important screens for 2-3 seconds
✓ **Smooth movements** - Straight lines, no jerky mouse
✓ **Clear screen** - Hide unnecessary UI elements
✓ **High contrast** - Use themes that are easy to read
✓ **Test first** - Do a dry run without recording
✓ **Prepare text** - Have commands ready to copy-paste

### Don'ts

✗ **Too fast** - Viewers can't follow
✗ **Mouse off-screen** - Cursor disappears
✗ **No pauses** - Overwhelming to watch
✗ **Too long** - Keep under 60 seconds
✗ **Shaky cursor** - Distracting
✗ **Personal info** - Hide usernames, paths, etc.
✗ **Low quality** - Use adequate resolution and colors

## Troubleshooting

### GIF Too Large

**Problem:** File size > 5MB

**Solutions:**

```bash
# Reduce everything
gifsicle -O3 --scale 0.75 --colors 128 --delay=15 input.gif -o output.gif

# Or use online compressor
# https://ezgif.com/optimize
```

### Poor Quality

**Problem:** Text is blurry or hard to read

**Solutions:**

- Record at higher resolution (1920x1080)
- Use 256 colors (not less)
- Use better screen recorder
- Increase font size before recording

### Not Looping

**Problem:** GIF plays once and stops

**Solution:**

```bash
gifsicle --loopcount=0 input.gif -o output.gif
```

### Stuttering Playback

**Problem:** GIF plays unevenly

**Solutions:**

- Remove duplicate frames
- Optimize with gifsicle -O3
- Use consistent frame delay
- Re-record with better tool

## Alternatives to GIF

### Option 1: Static Screenshots

Take screenshots at key moments:

```
docs/assets/screenshots/
├── quick-start-01-init.png
├── quick-start-02-vscode.png
├── quick-start-03-button.png
├── quick-start-04-action.png
└── quick-start-05-result.png
```

Reference in documentation with numbered steps.

**Pros:**

- Easy to create
- Small file sizes
- Always readable
- Easy to update

**Cons:**

- Less engaging
- Can't show flow
- More explanation needed

### Option 2: Video

Record video and upload to:

- YouTube
- Vimeo
- GitHub Releases

Link in documentation.

**Pros:**

- Higher quality
- Can be longer
- Can include audio explanation
- No file size concerns

**Cons:**

- Requires video hosting
- Longer to produce
- Harder to update

### Option 3: SVG Animation

Use termtosvg for terminal recordings:

```bash
pip install termtosvg
termtosvg output.svg
```

**Pros:**

- Vector graphics (always crisp)
- Smaller file size
- Text is selectable

**Cons:**

- Only works for terminal
- Not widely supported
- Can't capture GUI

## Embedding in Documentation

### Markdown

```markdown
## Quick Start

![Quick Start Demo](../../assets/quick-start-guide.gif)

Follow these steps to get started...
```

### HTML with Size Control

```html
<p align="center">
  <img src="../../assets/quick-start-guide.gif"
       width="800"
       alt="Quick Start Demo">
</p>
```

### With Caption

```markdown
<figure>
  <img src="../../assets/quick-start-guide.gif" alt="Quick Start">
  <figcaption>Complete setup and first use of the extension</figcaption>
</figure>
```

## Verification

After creating GIFs:

```bash
# Check file sizes
ls -lh docs/assets/*.gif

# Expected output:
# quick-start-guide.gif  2.8M
# git-commit-demo.gif    1.9M
# mcp-server-test.gif    2.1M

# Test in browser
# Open any markdown file with the GIF and preview
```

## Checklist

Before finalizing:

- [ ] Duration is appropriate (< 60 seconds)
- [ ] File size is reasonable (< 5MB)
- [ ] All text is readable
- [ ] Actions are clear and obvious
- [ ] No personal information visible
- [ ] GIF loops smoothly
- [ ] No stuttering or jumps
- [ ] Colors are pleasant
- [ ] Tested in browser
- [ ] Tested in README preview

## Resources

**Tools:**

- [ScreenToGif](https://www.screentogif.com/) - Windows
- [Kap](https://getkap.co/) - macOS
- [Peek](https://github.com/phw/peek) - Linux
- [asciinema](https://asciinema.org/) - Terminal
- [gifsicle](https://www.lcdf.org/gifsicle/) - Optimization

**Guides:**

- [Making Great GIFs for Documentation](https://increment.com/documentation/gifs-documentation/)
- [Animated GIF Best Practices](https://dev.to/lydiahallie/animated-gifs-best-practices-1g7p)

**Online Tools:**

- [ezgif.com](https://ezgif.com/) - GIF editor and optimizer
- [gifcompressor.com](https://gifcompressor.com/) - Simple compression
