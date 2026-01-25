---
name: drawio-editor
description: Edit DrawIO diagrams using natural language
---

# DrawIO Diagram Editor Skill

This skill enables Claude to create, view, and edit DrawIO diagrams stored as .drawio.png files, following the EAC visual language.

## When to Use This Skill

- Creating new architecture diagrams, flowcharts, or technical illustrations
- Modifying existing .drawio.png files based on user requests
- Analyzing diagram content to understand system architecture
- Converting text descriptions into visual diagrams

## Important Notes

- **PNG preview is not updated** - only the embedded XML is modified
- When the user opens the file in DrawIO, it renders correctly from the XML
- Always use `.drawio.png` extension (not `.drawio`)
- Follow the EAC visual style (gray background, shadows, Lucida Console font)

## Workflow

### Step 1: Understand the Request

Analyze what the user wants:

- **New diagram?** → Use `drawio create` then modify
- **Edit existing?** → Use `drawio decode` to read current state
- **Just viewing?** → Use `drawio info` for quick overview

### Step 2: Read Current State (for edits)

```bash
mcp__commands__drawio-decode --input <file.drawio.png>
```

### Step 3: Modify the XML

Edit the decoded XML using the EAC component library below.

### Step 4: Write Changes Back

```bash
# 1. Save modified XML to temp file
# 2. Encode
mcp__commands__drawio-encode --input /path/to/modified.xml --output /path/to/encoded.xml
# 3. Embed
mcp__commands__drawio-embed --png <file.drawio.png> --xml /path/to/encoded.xml
```

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
<mxfile host="drawio-cli" agent="drawio-cli (Claude)" version="1.0">
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
