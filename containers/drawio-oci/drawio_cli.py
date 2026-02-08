#!/usr/bin/env python3
"""
Drawio CLI - Extract and manipulate drawio diagram files.

Commands:
  extract  - Extract raw XML from drawio.png files (encoded format)
  decode   - Decode to human-readable XML (for LLM editing)
  encode   - Encode human-readable XML back to drawio format
  embed    - Embed XML into a PNG file (update existing)
  create   - Create new .drawio.png file
  render   - Render diagram to actual PNG image
  info     - Show information about a drawio file
  chunks   - List PNG chunks (debug)
"""

import argparse
import json
import sys
from pathlib import Path

from extract import extract_drawio_xml, read_png_chunks
from codec import (
    decode_mxfile,
    encode_mxfile,
    create_blank_mxfile,
    get_mxfile_info,
)
from embed import (
    embed_xml_in_png,
    create_drawio_png,
)


def cmd_extract(args):
    """Extract raw XML from drawio.png file (encoded format)."""
    xml = extract_drawio_xml(args.input)
    if xml is None:
        print(f"Error: No drawio data found in {args.input}", file=sys.stderr)
        return 1

    if args.pretty:
        try:
            from lxml import etree
            doc = etree.fromstring(xml.encode("utf-8"))
            xml = etree.tostring(doc, pretty_print=True, encoding="unicode")
        except Exception as e:
            print(f"Warning: Could not prettify XML: {e}", file=sys.stderr)

    if args.output:
        Path(args.output).write_text(xml, encoding="utf-8")
        print(f"Extracted to {args.output}", file=sys.stderr)
    else:
        print(xml)

    return 0


def cmd_decode(args):
    """Decode drawio content to human-readable XML."""
    # Read input
    if args.input:
        filepath = args.input
        if filepath.endswith(".png"):
            xml = extract_drawio_xml(filepath)
            if xml is None:
                print(f"Error: No drawio data found in {filepath}", file=sys.stderr)
                return 1
        else:
            xml = Path(filepath).read_text(encoding="utf-8")
    else:
        # Read from stdin
        xml = sys.stdin.read()

    # Decode
    try:
        decoded = decode_mxfile(xml)
    except ValueError as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1

    # Output
    if args.output:
        Path(args.output).write_text(decoded, encoding="utf-8")
        print(f"Decoded to {args.output}", file=sys.stderr)
    else:
        print(decoded)

    return 0


def cmd_encode(args):
    """Encode human-readable XML back to drawio format."""
    # Read input
    if args.input:
        xml = Path(args.input).read_text(encoding="utf-8")
    else:
        xml = sys.stdin.read()

    # Encode
    try:
        encoded = encode_mxfile(xml)
    except ValueError as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1

    # Output
    if args.output:
        Path(args.output).write_text(encoded, encoding="utf-8")
        print(f"Encoded to {args.output}", file=sys.stderr)
    else:
        print(encoded)

    return 0


def cmd_embed(args):
    """Embed XML into a PNG file."""
    # Read XML
    if args.xml:
        xml = Path(args.xml).read_text(encoding="utf-8")
    else:
        xml = sys.stdin.read()

    # Determine output path
    output = args.output if args.output else args.png

    # Check if PNG exists
    png_path = args.png if Path(args.png).exists() else None

    # Embed
    try:
        embed_xml_in_png(png_path, xml, output)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1

    print(f"Embedded XML into {output}", file=sys.stderr)
    return 0


def cmd_create(args):
    """Create a new .drawio.png file."""
    # Get XML content
    if args.xml:
        # Use provided XML file
        xml = Path(args.xml).read_text(encoding="utf-8")
        # If it's decoded format, encode it
        if "<mxGraphModel" in xml and "mxfile" in xml:
            try:
                xml = encode_mxfile(xml)
            except ValueError:
                pass  # Already encoded or will fail at embed
    else:
        # Create blank diagram
        decoded = create_blank_mxfile(args.name)
        xml = encode_mxfile(decoded)

    # Create the file
    try:
        create_drawio_png(xml, args.output)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1

    print(f"Created {args.output}", file=sys.stderr)
    return 0


def cmd_info(args):
    """Show information about a drawio file."""
    filepath = args.input

    # Read XML based on file type
    if filepath.endswith(".png"):
        xml = extract_drawio_xml(filepath)
        if xml is None:
            print(f"Error: No drawio data found in {filepath}", file=sys.stderr)
            return 1
        file_format = "drawio.png"
    else:
        xml = Path(filepath).read_text(encoding="utf-8")
        file_format = "drawio/xml"

    # Get info
    info = get_mxfile_info(xml)
    info["file"] = filepath
    info["format"] = file_format

    if "error" in info:
        print(f"Error: {info['error']}", file=sys.stderr)
        return 1

    if args.json:
        print(json.dumps(info, indent=2))
    else:
        print(f"File: {info['file']}")
        print(f"Format: {info['format']}")
        print(f"Host: {info.get('host', 'unknown')}")
        print(f"Agent: {info.get('agent', 'unknown')}")
        print(f"Version: {info.get('version', 'unknown')}")
        print(f"Diagrams: {len(info.get('diagrams', []))}")
        for diag in info.get("diagrams", []):
            cells = diag.get("cells", "?")
            print(f"  - {diag['name']} ({cells} cells)")

    return 0


def cmd_chunks(args):
    """List PNG chunks in a file (for debugging)."""
    chunks = read_png_chunks(args.input)

    print(f"PNG chunks in {args.input}:")
    print(f"{'Type':<6} {'Length':>10}  Notes")
    print("-" * 40)

    for chunk_type, data in chunks:
        notes = ""
        if chunk_type in (b"tEXt", b"iTXt", b"zTXt"):
            try:
                null_idx = data.index(b"\x00")
                keyword = data[:null_idx].decode("latin-1")
                notes = f"keyword={keyword}"
            except (ValueError, UnicodeDecodeError):
                pass

        chunk_name = chunk_type.decode("ascii") if isinstance(chunk_type, bytes) else chunk_type
        print(f"{chunk_name:<6} {len(data):>10}  {notes}")

    return 0


def cmd_render(args):
    """Render diagram to actual PNG image."""
    from render import render_diagram

    try:
        render_diagram(args.input, args.output, max_width=args.max_width)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1

    return 0


def main():
    parser = argparse.ArgumentParser(
        description="Drawio CLI - Extract and manipulate drawio diagram files",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Extract raw XML (encoded format)
  drawio_cli.py extract diagram.drawio.png

  # Decode to human-readable XML for editing
  drawio_cli.py decode -i diagram.drawio.png -o decoded.xml

  # After editing, encode back
  drawio_cli.py encode -i decoded.xml -o encoded.xml

  # Embed into PNG
  drawio_cli.py embed --png diagram.drawio.png --xml encoded.xml

  # Create new diagram
  drawio_cli.py create -o new.drawio.png --name "My Diagram"

  # Full edit cycle (decode -> edit -> encode -> embed)
  drawio_cli.py decode -i diagram.drawio.png | edit | drawio_cli.py encode | \\
    drawio_cli.py embed --png diagram.drawio.png
"""
    )
    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # Extract command
    extract_parser = subparsers.add_parser(
        "extract",
        help="Extract raw XML from drawio.png (encoded format)"
    )
    extract_parser.add_argument("input", help="Input drawio.png file")
    extract_parser.add_argument("-o", "--output", help="Output file (default: stdout)")
    extract_parser.add_argument("-p", "--pretty", action="store_true",
                                help="Pretty-print the XML")

    # Decode command
    decode_parser = subparsers.add_parser(
        "decode",
        help="Decode to human-readable XML (for LLM editing)"
    )
    decode_parser.add_argument("-i", "--input",
                               help="Input file (.drawio.png or .xml, default: stdin)")
    decode_parser.add_argument("-o", "--output",
                               help="Output file (default: stdout)")

    # Encode command
    encode_parser = subparsers.add_parser(
        "encode",
        help="Encode human-readable XML back to drawio format"
    )
    encode_parser.add_argument("-i", "--input",
                               help="Input decoded XML file (default: stdin)")
    encode_parser.add_argument("-o", "--output",
                               help="Output file (default: stdout)")

    # Embed command
    embed_parser = subparsers.add_parser(
        "embed",
        help="Embed XML into a PNG file"
    )
    embed_parser.add_argument("--png", required=True,
                              help="Target PNG file (created if not exists)")
    embed_parser.add_argument("--xml",
                              help="XML file to embed (default: stdin)")
    embed_parser.add_argument("-o", "--output",
                              help="Output file (default: update --png in place)")

    # Create command
    create_parser = subparsers.add_parser(
        "create",
        help="Create new .drawio.png file"
    )
    create_parser.add_argument("-o", "--output", required=True,
                               help="Output .drawio.png file")
    create_parser.add_argument("--xml",
                               help="XML file to use (default: blank diagram)")
    create_parser.add_argument("--name", default="Page-1",
                               help="Diagram page name (default: Page-1)")

    # Info command
    info_parser = subparsers.add_parser(
        "info",
        help="Show drawio file information"
    )
    info_parser.add_argument("input", help="Input drawio file")
    info_parser.add_argument("--json", action="store_true",
                             help="Output as JSON")

    # Chunks command (debug)
    chunks_parser = subparsers.add_parser(
        "chunks",
        help="List PNG chunks (debug)"
    )
    chunks_parser.add_argument("input", help="Input PNG file")

    # Render command
    render_parser = subparsers.add_parser(
        "render",
        help="Render diagram to actual PNG image"
    )
    render_parser.add_argument("-i", "--input", required=True,
                               help="Input .drawio.png or decoded XML file")
    render_parser.add_argument("-o", "--output", required=True,
                               help="Output PNG file (rendered image)")
    render_parser.add_argument("--max-width", type=int, default=0,
                               help="Maximum output width in pixels (0 = no limit)")

    args = parser.parse_args()

    if args.command is None:
        parser.print_help()
        return 0

    commands = {
        "extract": cmd_extract,
        "decode": cmd_decode,
        "encode": cmd_encode,
        "embed": cmd_embed,
        "create": cmd_create,
        "info": cmd_info,
        "chunks": cmd_chunks,
        "render": cmd_render,
    }

    return commands[args.command](args)


if __name__ == "__main__":
    sys.exit(main())
