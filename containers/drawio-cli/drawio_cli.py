#!/usr/bin/env python3
"""
Drawio CLI - Extract and manipulate drawio diagram files.

Commands:
  extract  - Extract XML from drawio.png files
  info     - Show information about a drawio file
  embed    - Embed XML into a PNG file (create drawio.png)
"""

import argparse
import json
import sys
import xml.etree.ElementTree as ET
from pathlib import Path

from extract import extract_drawio_xml, read_png_chunks


def cmd_extract(args):
    """Extract XML from drawio.png file."""
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
        print(f"Extracted to {args.output}")
    else:
        print(xml)

    return 0


def cmd_info(args):
    """Show information about a drawio file."""
    filepath = args.input

    # Check file type
    if filepath.endswith(".drawio") or filepath.endswith(".xml"):
        xml = Path(filepath).read_text(encoding="utf-8")
    elif filepath.endswith(".png"):
        xml = extract_drawio_xml(filepath)
        if xml is None:
            print(f"Error: No drawio data found in {filepath}", file=sys.stderr)
            return 1
    else:
        print(f"Error: Unknown file type: {filepath}", file=sys.stderr)
        return 1

    # Parse XML
    try:
        root = ET.fromstring(xml)
    except ET.ParseError as e:
        print(f"Error: Invalid XML: {e}", file=sys.stderr)
        return 1

    info = {
        "file": filepath,
        "format": "drawio.png" if filepath.endswith(".png") else "drawio",
    }

    # Extract mxfile attributes
    if root.tag == "mxfile":
        info["host"] = root.get("host", "unknown")
        info["modified"] = root.get("modified", "unknown")
        info["agent"] = root.get("agent", "unknown")
        info["version"] = root.get("version", "unknown")

        # Count diagrams (pages)
        diagrams = root.findall(".//diagram")
        info["diagrams"] = len(diagrams)

        # Get diagram names
        info["diagram_names"] = [d.get("name", f"Page {i+1}") for i, d in enumerate(diagrams)]

        # Count cells
        cells = root.findall(".//mxCell")
        info["cells"] = len(cells)

    if args.json:
        print(json.dumps(info, indent=2))
    else:
        print(f"File: {info['file']}")
        print(f"Format: {info.get('format', 'unknown')}")
        if "host" in info:
            print(f"Host: {info['host']}")
        if "version" in info:
            print(f"Version: {info['version']}")
        if "diagrams" in info:
            print(f"Diagrams: {info['diagrams']}")
            for name in info.get("diagram_names", []):
                print(f"  - {name}")
        if "cells" in info:
            print(f"Cells: {info['cells']}")

    return 0


def cmd_chunks(args):
    """List PNG chunks in a file (for debugging)."""
    chunks = read_png_chunks(args.input)

    print(f"PNG chunks in {args.input}:")
    print(f"{'Type':<6} {'Length':>10}  Notes")
    print("-" * 40)

    for chunk_type, data in chunks:
        notes = ""
        if chunk_type in ("tEXt", "iTXt", "zTXt"):
            # Try to get keyword
            try:
                null_idx = data.index(b"\x00")
                keyword = data[:null_idx].decode("latin-1")
                notes = f"keyword={keyword}"
            except (ValueError, UnicodeDecodeError):
                pass

        print(f"{chunk_type:<6} {len(data):>10}  {notes}")

    return 0


def main():
    parser = argparse.ArgumentParser(
        description="Drawio CLI - Extract and manipulate drawio diagram files"
    )
    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # Extract command
    extract_parser = subparsers.add_parser("extract", help="Extract XML from drawio.png")
    extract_parser.add_argument("input", help="Input drawio.png file")
    extract_parser.add_argument("-o", "--output", help="Output file (default: stdout)")
    extract_parser.add_argument(
        "-p", "--pretty", action="store_true", help="Pretty-print the XML"
    )

    # Info command
    info_parser = subparsers.add_parser("info", help="Show drawio file information")
    info_parser.add_argument("input", help="Input drawio file (.drawio, .xml, or .png)")
    info_parser.add_argument(
        "--json", action="store_true", help="Output as JSON"
    )

    # Chunks command (for debugging)
    chunks_parser = subparsers.add_parser("chunks", help="List PNG chunks (debug)")
    chunks_parser.add_argument("input", help="Input PNG file")

    args = parser.parse_args()

    if args.command is None:
        parser.print_help()
        return 0

    if args.command == "extract":
        return cmd_extract(args)
    elif args.command == "info":
        return cmd_info(args)
    elif args.command == "chunks":
        return cmd_chunks(args)

    return 0


if __name__ == "__main__":
    sys.exit(main())
