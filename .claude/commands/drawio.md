# DrawIO Editor

```text
description: "Create and edit DrawIO diagrams (.drawio.png files)"
```

You are editing or creating DrawIO diagrams using the EAC visual language.

## Commands Available

Use the eac-cli CLI to work with DrawIO files:

```bash
# Create a new diagram
drawio create --output <file.drawio.png> [--name "Page Name"]

# View diagram info
drawio info --input <file.drawio.png>

# Decode to readable XML (for editing)
drawio decode --input <file.drawio.png>

# Encode edited XML back to DrawIO format
drawio encode --input <decoded.xml> --output <encoded.xml>

# Embed encoded XML into PNG
drawio embed --png <file.drawio.png> --xml <encoded.xml>
```

## Workflow

### Creating a New Diagram

1. Create blank file: `drawio create --output docs/my-diagram.drawio.png`
2. Decode it: `drawio decode --input docs/my-diagram.drawio.png`
3. Edit the XML following EAC visual language (see skill file)
4. Save modified XML to a temp file
5. Encode: `drawio encode --input temp.xml --output encoded.xml`
6. Embed: `drawio embed --png docs/my-diagram.drawio.png --xml encoded.xml`

### Editing an Existing Diagram

1. Decode: `drawio decode --input <file.drawio.png>`
2. Analyze current content
3. Make modifications to the XML
4. Encode and embed back

## EAC Visual Style

- **Background:** `#CFCFCF` (gray)
- **Shadows:** Enabled on all elements
- **Font:** Lucida Console, bold
- **Grid:** 10px alignment

## Component Library

See `.claude/skills/drawio-editor.md` for the full EAC component library including:

- **Trunk** (cylinder) - Repository
- **Module** (hexagon) - Deployable unit
- **Pipeline** (horizontal cylinder) - CI/CD flow
- **LIVE** (cloud) - Production
- **Environment** (circle) - Dev/staging instances
- **Quality Gate** (diamond + automation icon)
- **Signoff** (diamond + checkmark)

## Example Usage

- `/drawio create a pipeline diagram showing trunk -> build -> test -> deploy`
- `/drawio add a quality gate between build and test in docs/pipeline.drawio.png`
- `/drawio show me what's in docs/architecture.drawio.png`
