# DrawIO Editor

```text
description: "Create and edit DrawIO diagrams (.drawio.png files)"
```

You are editing or creating DrawIO diagrams using the EAC visual language.

## When to Use

- Creating new architecture diagrams, flowcharts, or technical illustrations
- Modifying existing .drawio.png files based on user requests
- Analyzing diagram content to understand system architecture
- Converting text descriptions into visual diagrams

## Important Notes

- **PNG preview is not updated** - only the embedded XML is modified
- When the user opens the file in DrawIO, it renders correctly from the XML
- Always use `.drawio.png` extension (not `.drawio`)
- Follow the EAC visual style (gray background, shadows, Lucida Console font)

---

## Commands Available

Use the eac CLI to work with DrawIO files:

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

---

## Workflow

### Creating a New Diagram

1. **Create blank file**: `drawio create --output docs/my-diagram.drawio.png`
2. **Decode it**: `drawio decode --input docs/my-diagram.drawio.png`
3. **Edit the XML** following EAC visual language (see below)
4. **Save modified XML** to a temp file
5. **Encode**: `drawio encode --input temp.xml --output encoded.xml`
6. **Embed**: `drawio embed --png docs/my-diagram.drawio.png --xml encoded.xml`

### Editing an Existing Diagram

1. **Decode**: `drawio decode --input <file.drawio.png>`
2. **Analyze current content**
3. **Make modifications** to the XML
4. **Encode and embed** back

### Understanding a Request

Analyze what the user wants:

- **New diagram?** → Use `drawio create` then modify
- **Edit existing?** → Use `drawio decode` to read current state
- **Just viewing?** → Use `drawio info` for quick overview

---

## EAC Visual Language

### Global Settings

All EAC diagrams use these base settings:

```xml
<mxGraphModel grid="1" gridSize="10" guides="1" tooltips="1" connect="1"
              arrows="1" fold="1" page="1" pageScale="1"
              background="#CFCFCF" shadow="1">
```

### Typography

- **Font:** `fontFamily=Lucida Console`
- **Style:** `fontStyle=1` (bold)
- **Size:** `fontSize=14` (default)

---

## EAC Component Library

### 1. Trunk (Repository)

Vertical cylinder representing the codebase.

```xml
<mxCell id="trunk" value="Trunk"
        style="shape=cylinder3;whiteSpace=wrap;html=1;boundedLbl=1;backgroundOutline=1;size=15;fontFamily=Lucida Console;fontStyle=1;shadow=1"
        vertex="1" parent="1">
  <mxGeometry x="100" y="100" width="80" height="90" as="geometry"/>
</mxCell>
```

### 2. Deployment Pipeline

Horizontal cylinder (rotated 90°) for CI/CD flow.

```xml
<mxCell id="pipeline" value=""
        style="shape=cylinder3;whiteSpace=wrap;html=1;boundedLbl=1;backgroundOutline=1;size=15;rotation=90;fontFamily=Lucida Console;fontStyle=1;shadow=1"
        vertex="1" parent="1">
  <mxGeometry x="100" y="50" width="31" height="207" as="geometry"/>
</mxCell>
<mxCell id="pipeline-label" value="Deployment Pipeline"
        style="text;html=1;strokeColor=none;fillColor=none;align=center;verticalAlign=middle;whiteSpace=wrap;rounded=0;fontFamily=Lucida Console;fontStyle=1;fontSize=14;shadow=1"
        vertex="1" parent="1">
  <mxGeometry x="50" y="160" width="180" height="20" as="geometry"/>
</mxCell>
```

### 3. Deployable Module

Hexagon for software modules.

```xml
<mxCell id="module" value="Module Name"
        style="shape=hexagon;perimeter=hexagonPerimeter2;whiteSpace=wrap;html=1;fixedSize=1;fontFamily=Lucida Console;fontStyle=1;fontSize=14;shadow=1"
        vertex="1" parent="1">
  <mxGeometry x="100" y="100" width="116" height="68" as="geometry"/>
</mxCell>
```

### 4. LIVE (Production)

Cloud shape for production environment.

```xml
<mxCell id="live" value="LIVE"
        style="whiteSpace=wrap;html=1;shape=mxgraph.basic.cloud_rect;fontFamily=Lucida Console;fontStyle=1;fontSize=14;shadow=1"
        vertex="1" parent="1">
  <mxGeometry x="100" y="100" width="90" height="73" as="geometry"/>
</mxCell>
```

### 5. Environment Instance

Circle for any environment.

```xml
<mxCell id="env" value="dev"
        style="ellipse;whiteSpace=wrap;html=1;aspect=fixed;strokeColor=default;fontSize=11;fontStyle=1;fontFamily=Lucida Console;shadow=1"
        vertex="1" parent="1">
  <mxGeometry x="100" y="100" width="90" height="90" as="geometry"/>
</mxCell>
```

### 6. Quality Gate

Diamond with automation icon.

```xml
<mxCell id="gate-group" value="" style="group" vertex="1" connectable="0" parent="1">
  <mxGeometry x="100" y="100" width="110" height="90" as="geometry"/>
</mxCell>
<mxCell id="gate-diamond" value=""
        style="rhombus;whiteSpace=wrap;html=1;shadow=0;fontFamily=Lucida Console;fontSize=14;align=center;strokeWidth=1;strokeColor=#050505;fontStyle=1"
        vertex="1" parent="gate-group">
  <mxGeometry width="110" height="90" as="geometry"/>
</mxCell>
<mxCell id="gate-icon" value=""
        style="verticalLabelPosition=bottom;html=1;verticalAlign=top;align=center;strokeColor=#050505;fillColor=#00BEF2;shape=mxgraph.azure.automation;shadow=1"
        vertex="1" parent="gate-group">
  <mxGeometry x="32.5" y="27" width="45" height="36" as="geometry"/>
</mxCell>
```

### 7. Signoff / Approval

Diamond with checkmark.

```xml
<mxCell id="signoff-group" value="" style="group;shadow=1" vertex="1" connectable="0" parent="1">
  <mxGeometry x="100" y="100" width="110" height="90" as="geometry"/>
</mxCell>
<mxCell id="signoff-diamond" value=""
        style="rhombus;whiteSpace=wrap;html=1;shadow=0;fontFamily=Lucida Console;fontSize=14;strokeWidth=1;strokeColor=#050505;fontStyle=1"
        vertex="1" parent="signoff-group">
  <mxGeometry width="110" height="90" as="geometry"/>
</mxCell>
<mxCell id="signoff-tick" value=""
        style="verticalLabelPosition=bottom;verticalAlign=top;html=1;shape=mxgraph.basic.tick;fillColor=#66B2FF;shadow=1"
        vertex="1" parent="signoff-group">
  <mxGeometry x="39" y="27.5" width="32" height="35" as="geometry"/>
</mxCell>
```

### 8. Start/End Terminal

```xml
<mxCell id="start" value="Start"
        style="strokeWidth=2;html=1;shape=mxgraph.flowchart.start_1;whiteSpace=wrap;shadow=1"
        vertex="1" parent="1">
  <mxGeometry x="100" y="100" width="100" height="60" as="geometry"/>
</mxCell>
```

---

## Test Level Colors

| Level     | Purpose     | Fill      | Stroke    | Gradient  |
| --------- | ----------- | --------- | --------- | --------- |
| **L0/L1** | Unit Tests  | `#dae8fc` | `#6c8ebf` | `#7ea6e0` |
| **L2**    | Component   | `#fff2cc` | `#d6b656` | `#ffd966` |
| **L3**    | Integration | `#ffcd28` | `#d79b00` | `#ffa500` |
| **L4**    | System/E2E  | `#FA3232` | `#ae4132` | -         |

### Semantic Color Palette

Use these colors for consistent semantic meaning across diagrams:

| Semantic Meaning | Fill | Stroke | Use For |
|-----------------|------|--------|---------|
| Success/Production | `#d5e8d4` | `#82b366` | LIVE, completed states, positive outcomes |
| Build/CI/Modules | `#dae8fc` | `#6c8ebf` | Build stages, modules, CI processes |
| Trunk/Source | `#fff2cc` | `#d6b656` | Source code, repositories, warnings |
| Deploy/Process | `#e1d5e7` | `#9673a6` | Deployment stages, processes, transforms |
| Gates/Errors | `#f8cecc` | `#b85450` | Quality gates, alerts, errors, blockers |
| Neutral | `#f5f5f5` | `#666666` | Legends, info boxes, backgrounds |

### Test Level Ellipse Example

```xml
<mxCell id="l2" value="&lt;b&gt;&lt;font style='font-size: 30px;'&gt;L2&lt;/font&gt;&lt;/b&gt;"
        style="ellipse;whiteSpace=wrap;html=1;fillColor=#fff2cc;strokeColor=#d6b656;gradientColor=#ffd966;shadow=1"
        vertex="1" parent="1">
  <mxGeometry x="100" y="100" width="200" height="100" as="geometry"/>
</mxCell>
```

---

## Arrows / Connectors

Standard EAC flow arrow:

```xml
<mxCell id="arrow1" value=""
        style="endArrow=classic;html=1;rounded=0;fontFamily=Lucida Console;fontStyle=1;shadow=1"
        edge="1" parent="1">
  <mxGeometry relative="1" as="geometry">
    <mxPoint x="100" y="150" as="sourcePoint"/>
    <mxPoint x="200" y="150" as="targetPoint"/>
  </mxGeometry>
</mxCell>
```

### Curved Arrow (for branches)

```xml
<mxCell id="curved" style="curved=1;endArrow=classic;html=1;strokeWidth=2;strokeColor=#82b366;" edge="1" parent="1">
  <mxGeometry relative="1" as="geometry">
    <mxPoint x="100" y="200" as="sourcePoint"/>
    <mxPoint x="300" y="200" as="targetPoint"/>
    <Array as="points"><mxPoint x="200" y="150"/></Array>
  </mxGeometry>
</mxCell>
```

---

## Advanced Components

These components create professional, human-quality diagrams with proper visual hierarchy.

### 1. Container Background

Large rounded rectangle with low opacity for grouping related elements. Place behind content.

```xml
<mxCell id="container" value=""
        style="rounded=1;whiteSpace=wrap;html=1;fillColor=#dae8fc;strokeColor=#6c8ebf;strokeWidth=2;opacity=30;"
        vertex="1" parent="1">
  <mxGeometry x="30" y="70" width="400" height="200" as="geometry"/>
</mxCell>
```

### 2. Pill Header

Small rounded rectangle with white text for section titles. Place inside containers.

```xml
<mxCell id="header" value="SECTION"
        style="rounded=1;whiteSpace=wrap;html=1;fillColor=#6c8ebf;strokeColor=#6c8ebf;fontColor=#ffffff;fontFamily=Lucida Console;fontStyle=1;fontSize=12;"
        vertex="1" parent="1">
  <mxGeometry x="60" y="80" width="150" height="30" as="geometry"/>
</mxCell>
```

### 3. Legend Box

Gray background explaining symbols used in the diagram.

```xml
<mxCell id="legend-box" value=""
        style="rounded=1;whiteSpace=wrap;html=1;fillColor=#f5f5f5;strokeColor=#666666;strokeWidth=1;"
        vertex="1" parent="1">
  <mxGeometry x="500" y="400" width="200" height="120" as="geometry"/>
</mxCell>
<mxCell id="legend-title" value="Legend"
        style="text;html=1;strokeColor=none;fillColor=none;align=left;verticalAlign=middle;fontFamily=Lucida Console;fontStyle=1;fontSize=12;"
        vertex="1" parent="1">
  <mxGeometry x="510" y="405" width="80" height="20" as="geometry"/>
</mxCell>
```

### 4. Key Insight Callout

Highlighted box for important messages. Use deploy/process color (purple).

```xml
<mxCell id="insight" value="Key Insight: Important message here"
        style="rounded=1;whiteSpace=wrap;html=1;fillColor=#e1d5e7;strokeColor=#9673a6;strokeWidth=2;fontFamily=Lucida Console;fontStyle=1;fontSize=11;align=center;"
        vertex="1" parent="1">
  <mxGeometry x="60" y="500" width="400" height="30" as="geometry"/>
</mxCell>
```

### 5. Content Box

Colored box for bullet lists or grouped text content.

```xml
<mxCell id="content-box" value="• Item 1&lt;br&gt;• Item 2&lt;br&gt;• Item 3"
        style="rounded=1;whiteSpace=wrap;html=1;fillColor=#d5e8d4;strokeColor=#82b366;strokeWidth=1;fontFamily=Lucida Console;fontSize=11;align=left;verticalAlign=top;spacingLeft=10;spacingTop=5;"
        vertex="1" parent="1">
  <mxGeometry x="70" y="120" width="180" height="80" as="geometry"/>
</mxCell>
```

### 6. Quadrant Cell

For 2x2 grid layouts (Cynefin-style). Combine with pill headers.

```xml
<mxCell id="quadrant-tl" value=""
        style="rounded=1;whiteSpace=wrap;html=1;fillColor=#f8cecc;strokeColor=#b85450;strokeWidth=2;opacity=30;"
        vertex="1" parent="1">
  <mxGeometry x="30" y="70" width="300" height="200" as="geometry"/>
</mxCell>
```

---

## Layout Pattern Templates

Use these templates as starting points for common diagram types.

### Quadrant Layout (2x2 Grid)

Best for: Cynefin framework, decision matrices, comparison diagrams.

Structure:
- Canvas divided into 4 equal quadrants
- Optional side panel for legend/summary
- Pill headers in each quadrant
- Content boxes within quadrants

### Column Layout

Best for: DORA metrics, parallel processes, category comparisons.

Structure:
- 3-5 vertical columns with headers
- Content boxes below each header
- Relationship section at bottom
- Key insight callout

### Flow Layout (Pipeline Stages)

Best for: CD model, pipeline stages, sequential processes.

Structure:
- Horizontal flow with DEVELOPMENT and RELEASE containers
- Stage boxes within containers
- Connecting arrows between stages
- Legend box explaining symbols

### Timeline Layout

Best for: Branch visualization, trunk-based development, release trains.

Structure:
- Horizontal trunk line
- Curved arrows for branches
- Environment circles
- Commit markers

---

## Composition Guidelines

### Always Include

1. **Title** - Top of diagram, centered, bold, larger font (18-24px)
2. **Containers** - Group related elements with low-opacity backgrounds
3. **Pill Headers** - Label each container/section
4. **Legend** - Explain symbols, colors, and terminology
5. **Key Insight** - Purple callout summarizing the main takeaway

### Z-Order (Back to Front)

1. Container backgrounds (lowest)
2. Content boxes
3. Shapes (modules, gates, etc.)
4. Text labels
5. Arrows/connectors
6. Legend box (top-right, highest)

### Spacing Guidelines

| Element | Distance |
|---------|----------|
| Title to content | 60px |
| Between sections | 40px |
| Inside containers | 20px |
| Shape to arrow | 10px |
| Legend margin | 30px from edge |

### Title Style

```xml
<mxCell id="title" value="Diagram Title"
        style="text;html=1;strokeColor=none;fillColor=none;align=center;verticalAlign=middle;fontFamily=Lucida Console;fontStyle=1;fontSize=24;"
        vertex="1" parent="1">
  <mxGeometry x="300" y="20" width="400" height="40" as="geometry"/>
</mxCell>
```

---

## Layout Guidelines

- **Grid:** 10px (align everything to grid)
- **Arrow gaps:** 40px minimum between shapes
- **Standard sizes:**
  - Trunk: 80×90
  - Module: 116×68
  - Pipeline: 31×207 (rotated)
  - LIVE cloud: 90×73
  - Environment: 90×90
  - Gate/Signoff: 110×90
  - Terminal: 100×60

---

## Example: Pipeline Flow Diagram

User: "Create a diagram showing trunk → module → pipeline → live"

```xml
<mxfile host="drawio-oci" agent="drawio-oci (Claude)" version="1.0">
  <diagram name="Pipeline" id="pipeline-flow">
    <mxGraphModel dx="1426" dy="758" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="1654" pageHeight="1169" background="#CFCFCF" shadow="1">
      <root>
        <mxCell id="0"/>
        <mxCell id="1" parent="0"/>

        <mxCell id="trunk" value="Trunk"
                style="shape=cylinder3;whiteSpace=wrap;html=1;boundedLbl=1;backgroundOutline=1;size=15;fontFamily=Lucida Console;fontStyle=1;shadow=1"
                vertex="1" parent="1">
          <mxGeometry x="100" y="100" width="80" height="90" as="geometry"/>
        </mxCell>

        <mxCell id="a1" style="endArrow=classic;html=1;rounded=0;shadow=1" edge="1" parent="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="190" y="145" as="sourcePoint"/>
            <mxPoint x="240" y="145" as="targetPoint"/>
          </mxGeometry>
        </mxCell>

        <mxCell id="module" value="Module"
                style="shape=hexagon;perimeter=hexagonPerimeter2;whiteSpace=wrap;html=1;fixedSize=1;fontFamily=Lucida Console;fontStyle=1;fontSize=14;shadow=1"
                vertex="1" parent="1">
          <mxGeometry x="250" y="111" width="116" height="68" as="geometry"/>
        </mxCell>

        <mxCell id="a2" style="endArrow=classic;html=1;rounded=0;shadow=1" edge="1" parent="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="376" y="145" as="sourcePoint"/>
            <mxPoint x="426" y="145" as="targetPoint"/>
          </mxGeometry>
        </mxCell>

        <mxCell id="pipeline" value=""
                style="shape=cylinder3;whiteSpace=wrap;html=1;boundedLbl=1;backgroundOutline=1;size=15;rotation=90;fontFamily=Lucida Console;fontStyle=1;shadow=1"
                vertex="1" parent="1">
          <mxGeometry x="520" y="41" width="31" height="207" as="geometry"/>
        </mxCell>
        <mxCell id="pipeline-label" value="Pipeline"
                style="text;html=1;strokeColor=none;fillColor=none;align=center;verticalAlign=middle;fontFamily=Lucida Console;fontStyle=1;fontSize=14;shadow=1"
                vertex="1" parent="1">
          <mxGeometry x="476" y="155" width="120" height="20" as="geometry"/>
        </mxCell>

        <mxCell id="a3" style="endArrow=classic;html=1;rounded=0;shadow=1" edge="1" parent="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="630" y="145" as="sourcePoint"/>
            <mxPoint x="680" y="145" as="targetPoint"/>
          </mxGeometry>
        </mxCell>

        <mxCell id="live" value="LIVE"
                style="whiteSpace=wrap;html=1;shape=mxgraph.basic.cloud_rect;fontFamily=Lucida Console;fontStyle=1;fontSize=14;shadow=1"
                vertex="1" parent="1">
          <mxGeometry x="690" y="108" width="90" height="73" as="geometry"/>
        </mxCell>

      </root>
    </mxGraphModel>
  </diagram>
</mxfile>
```

---

## Tips

1. **Always include `shadow=1`** on all visible elements
2. **Use `fontFamily=Lucida Console;fontStyle=1`** for text
3. **Gray background `#CFCFCF`** is standard for EAC diagrams
4. **Align to 10px grid** for clean layouts
5. **Use meaningful IDs** like `trunk`, `module-api`, `gate-qa`

---

## Example Usage

- `/drawio` create a pipeline diagram showing trunk → build → test → deploy
- `/drawio` add a quality gate between build and test in docs/pipeline.drawio.png
- `/drawio` show me what's in docs/architecture.drawio.png
