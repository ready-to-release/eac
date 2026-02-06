#!/usr/bin/env python3
"""
DrawIO rendering - Export DrawIO diagrams to actual PNG images.

Uses Playwright with Chromium to render the diagram by loading it
in draw.io's web viewer and capturing the canvas.
"""

import base64
import sys
import tempfile
from pathlib import Path

# We'll use a simpler approach: generate an SVG from the mxGraphModel
# and convert it to PNG using cairosvg or pillow

def extract_diagram_xml(mxfile_path: str) -> str:
    """Extract and decode the diagram XML from an mxfile."""
    from codec import decode_mxfile, decode_diagram_content
    from embed import read_png_chunks, extract_mxfile_chunk

    if mxfile_path.endswith('.png'):
        chunks = read_png_chunks(mxfile_path)
        mxfile_xml = extract_mxfile_chunk(chunks)
        if not mxfile_xml:
            raise ValueError("No mxfile data found in PNG")
    else:
        mxfile_xml = Path(mxfile_path).read_text(encoding='utf-8')

    return decode_mxfile(mxfile_xml)


def find_graph_model(decoded_xml: str):
    """Find the mxGraphModel element, handling both decoded and encoded diagram content."""
    import xml.etree.ElementTree as ET
    from codec import decode_diagram_content

    root = ET.fromstring(decoded_xml)
    diagram = root.find('.//diagram')
    if diagram is None:
        raise ValueError("No diagram found in mxfile")

    # Case 1: mxGraphModel is already a child element (decoded by decode_mxfile)
    graph_model = diagram.find('mxGraphModel')
    if graph_model is not None:
        return graph_model

    # Case 2: diagram content is still encoded text (decode_mxfile couldn't parse it)
    if diagram.text and diagram.text.strip():
        decoded = decode_diagram_content(diagram.text.strip())
        try:
            inner = ET.fromstring(decoded)
            if inner.tag == 'mxGraphModel':
                return inner
            gm = inner.find('.//mxGraphModel')
            if gm is not None:
                return gm
        except ET.ParseError:
            pass

    # Case 3: search entire tree as fallback
    graph_model = root.find('.//mxGraphModel')
    if graph_model is not None:
        return graph_model

    raise ValueError("No mxGraphModel found in diagram")


def mxgraph_to_svg(decoded_xml: str) -> str:
    """
    Convert mxGraphModel XML to SVG.

    This is a simplified renderer that handles basic shapes.
    For full fidelity, use draw.io's official export.
    """
    import xml.etree.ElementTree as ET
    import html

    graph_model = find_graph_model(decoded_xml)

    # Get page dimensions
    page_width = int(graph_model.get('pageWidth', '1654'))
    page_height = int(graph_model.get('pageHeight', '1169'))
    background = graph_model.get('background', '#CFCFCF')

    # Start SVG
    svg_parts = [
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{page_width}" height="{page_height}" viewBox="0 0 {page_width} {page_height}">',
        f'<rect width="100%" height="100%" fill="{background}"/>',
        '<defs>',
        '<filter id="shadow" x="-20%" y="-20%" width="140%" height="140%">',
        '<feDropShadow dx="2" dy="2" stdDeviation="3" flood-opacity="0.3"/>',
        '</filter>',
        '</defs>',
    ]

    # Process cells
    cells = graph_model.findall('.//mxCell')
    for cell in cells:
        cell_id = cell.get('id', '')
        if cell_id in ('0', '1'):  # Skip root cells
            continue

        value = cell.get('value', '')
        style_str = cell.get('style', '')

        # Parse geometry
        geom = cell.find('mxGeometry')
        if geom is None:
            continue

        x = float(geom.get('x', '0'))
        y = float(geom.get('y', '0'))
        width = float(geom.get('width', '100'))
        height = float(geom.get('height', '50'))

        # Parse style
        style = {}
        for part in style_str.split(';'):
            if '=' in part:
                k, v = part.split('=', 1)
                style[k] = v
            elif part:
                style[part] = 'true'

        fill = style.get('fillColor', '#ffffff')
        stroke = style.get('strokeColor', '#000000')
        font_size = style.get('fontSize', '12')
        shadow = 'filter="url(#shadow)"' if style.get('shadow') == '1' else ''

        # Handle different shapes
        shape = style.get('shape', '')

        if 'text' in style or shape == '':
            # Text element
            if value:
                # Decode HTML entities
                text = html.unescape(value.replace('&#xa;', '\n'))
                lines = text.split('\n')

                text_anchor = 'middle'
                if style.get('align') == 'left':
                    text_anchor = 'start'
                elif style.get('align') == 'right':
                    text_anchor = 'end'

                tx = x + width/2
                if text_anchor == 'start':
                    tx = x + 5

                for i, line in enumerate(lines):
                    ty = y + height/2 + (i - len(lines)/2 + 0.5) * int(font_size) * 1.2
                    font_weight = 'bold' if style.get('fontStyle') == '1' else 'normal'
                    svg_parts.append(
                        f'<text x="{tx}" y="{ty}" font-family="Lucida Console, monospace" '
                        f'font-size="{font_size}" font-weight="{font_weight}" '
                        f'text-anchor="{text_anchor}" dominant-baseline="middle">{html.escape(line)}</text>'
                    )

        elif 'rounded' in style or 'rectangle' in shape.lower():
            # Rounded rectangle
            rx = '10' if 'rounded' in style else '0'
            if fill and fill != 'none':
                svg_parts.append(
                    f'<rect x="{x}" y="{y}" width="{width}" height="{height}" '
                    f'rx="{rx}" fill="{fill}" stroke="{stroke}" {shadow}/>'
                )
            if value:
                text = html.unescape(value.replace('&#xa;', '\n'))
                for i, line in enumerate(text.split('\n')):
                    ty = y + height/2 + (i - len(text.split('\n'))/2 + 0.5) * int(font_size) * 1.2
                    svg_parts.append(
                        f'<text x="{x + width/2}" y="{ty}" font-family="Lucida Console, monospace" '
                        f'font-size="{font_size}" text-anchor="middle" dominant-baseline="middle">{html.escape(line)}</text>'
                    )

        elif 'rhombus' in style or 'diamond' in shape.lower():
            # Diamond shape
            cx, cy = x + width/2, y + height/2
            points = f"{cx},{y} {x+width},{cy} {cx},{y+height} {x},{cy}"
            svg_parts.append(
                f'<polygon points="{points}" fill="{fill}" stroke="{stroke}" {shadow}/>'
            )
            if value:
                svg_parts.append(
                    f'<text x="{cx}" y="{cy}" font-family="Lucida Console, monospace" '
                    f'font-size="{int(font_size)-2}" text-anchor="middle" dominant-baseline="middle">{html.escape(value)}</text>'
                )

        elif 'ellipse' in style or 'cloud' in shape.lower():
            # Ellipse/cloud
            cx, cy = x + width/2, y + height/2
            rx, ry = width/2, height/2
            svg_parts.append(
                f'<ellipse cx="{cx}" cy="{cy}" rx="{rx}" ry="{ry}" '
                f'fill="{fill}" stroke="{stroke}" {shadow}/>'
            )
            if value:
                svg_parts.append(
                    f'<text x="{cx}" y="{cy}" font-family="Lucida Console, monospace" '
                    f'font-size="{font_size}" text-anchor="middle" dominant-baseline="middle">{html.escape(value)}</text>'
                )

        elif 'hexagon' in style or 'hexagon' in shape.lower():
            # Hexagon
            cx, cy = x + width/2, y + height/2
            hw, hh = width/2, height/2
            points = f"{x+hw*0.25},{y} {x+hw*1.75},{y} {x+width},{cy} {x+hw*1.75},{y+height} {x+hw*0.25},{y+height} {x},{cy}"
            svg_parts.append(
                f'<polygon points="{points}" fill="{fill}" stroke="{stroke}" {shadow}/>'
            )
            if value:
                text = html.unescape(value.replace('&#xa;', '\n'))
                for i, line in enumerate(text.split('\n')):
                    ty = cy + (i - len(text.split('\n'))/2 + 0.5) * int(font_size) * 1.2
                    svg_parts.append(
                        f'<text x="{cx}" y="{ty}" font-family="Lucida Console, monospace" '
                        f'font-size="{font_size}" text-anchor="middle" dominant-baseline="middle">{html.escape(line)}</text>'
                    )

        elif 'cylinder' in style or 'cylinder' in shape.lower():
            # Cylinder (simplified as rounded rect)
            svg_parts.append(
                f'<rect x="{x}" y="{y}" width="{width}" height="{height}" '
                f'rx="10" fill="{fill}" stroke="{stroke}" {shadow}/>'
            )
            if value:
                svg_parts.append(
                    f'<text x="{x + width/2}" y="{y + height/2}" font-family="Lucida Console, monospace" '
                    f'font-size="{font_size}" text-anchor="middle" dominant-baseline="middle">{html.escape(value)}</text>'
                )

        elif 'swimlane' in style:
            # Swimlane (container with header)
            svg_parts.append(
                f'<rect x="{x}" y="{y}" width="{width}" height="{height}" '
                f'fill="{fill}" stroke="{stroke}" {shadow}/>'
            )
            # Header
            svg_parts.append(
                f'<rect x="{x}" y="{y}" width="{width}" height="25" '
                f'fill="{fill}" stroke="{stroke}"/>'
            )
            if value:
                svg_parts.append(
                    f'<text x="{x + 10}" y="{y + 17}" font-family="Lucida Console, monospace" '
                    f'font-size="{font_size}" text-anchor="start">{html.escape(value)}</text>'
                )

        elif 'endArrow' in style:
            # Arrow/edge - need source and target points
            source = geom.find('.//mxPoint[@as="sourcePoint"]')
            target = geom.find('.//mxPoint[@as="targetPoint"]')
            if source is not None and target is not None:
                sx, sy = float(source.get('x', 0)), float(source.get('y', 0))
                tx, ty = float(target.get('x', 0)), float(target.get('y', 0))
                svg_parts.append(
                    f'<line x1="{sx}" y1="{sy}" x2="{tx}" y2="{ty}" '
                    f'stroke="#000000" stroke-width="2" marker-end="url(#arrowhead)"/>'
                )

    # Add arrowhead marker
    svg_parts.insert(2, '''<marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
      <polygon points="0 0, 10 3.5, 0 7" fill="#000000"/>
    </marker>''')

    svg_parts.append('</svg>')
    return '\n'.join(svg_parts)


def svg_to_png(svg_content: str, output_path: str) -> None:
    """Convert SVG to PNG using cairosvg."""
    import cairosvg
    cairosvg.svg2png(bytestring=svg_content.encode('utf-8'), write_to=output_path)


def optimize_png(output_path: str, max_width: int) -> None:
    """Resize PNG to fit within max_width, preserving aspect ratio."""
    from PIL import Image

    with Image.open(output_path) as img:
        w, h = img.size
        if w <= max_width:
            return

        ratio = max_width / w
        new_h = int(h * ratio)
        resized = img.resize((max_width, new_h), Image.LANCZOS)
        resized.save(output_path, "PNG", optimize=True)
        print(f"Optimized {w}x{h} -> {max_width}x{new_h}")


def render_diagram(input_path: str, output_path: str, max_width: int = 0) -> None:
    """Render a DrawIO diagram to PNG, optionally resizing to max_width."""
    decoded_xml = extract_diagram_xml(input_path)
    svg = mxgraph_to_svg(decoded_xml)
    svg_to_png(svg, output_path)

    if max_width > 0:
        optimize_png(output_path, max_width)

    print(f"Rendered to {output_path}")


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: render.py <input.drawio.png> <output.png>")
        print("       render.py <input.xml> <output.png>")
        sys.exit(1)

    input_path = sys.argv[1]
    output_path = sys.argv[2]

    try:
        render_diagram(input_path, output_path)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
