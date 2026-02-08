#!/usr/bin/env python3
"""
DrawIO codec - Encode and decode DrawIO diagram content.

DrawIO uses multiple encoding layers for diagram content:
1. mxGraphModel XML (the actual diagram)
2. deflate compressed
3. base64 encoded
4. URL encoded
5. Wrapped in <diagram> element
6. Wrapped in <mxfile> element

This module handles encoding/decoding between these layers.
"""

import base64
import re
import urllib.parse
import zlib
import xml.etree.ElementTree as ET
from typing import Optional


def decode_diagram_content(encoded: str) -> str:
    """
    Decode a single diagram's content from DrawIO format to mxGraphModel XML.

    Args:
        encoded: The encoded content from inside a <diagram> element

    Returns:
        Decoded mxGraphModel XML string
    """
    # Step 1: URL decode
    decoded = urllib.parse.unquote(encoded)

    # Check if already XML (some older formats store uncompressed)
    if decoded.strip().startswith("<"):
        return decoded

    # Step 2: Base64 decode
    try:
        raw = base64.b64decode(decoded)
    except Exception:
        # Not base64, return as-is
        return decoded

    # Step 3: Inflate (decompress)
    try:
        # Try raw deflate (no zlib header) - most common
        xml = zlib.decompress(raw, -zlib.MAX_WBITS).decode("utf-8")
        return xml
    except zlib.error:
        pass

    try:
        # Try with zlib header
        xml = zlib.decompress(raw).decode("utf-8")
        return xml
    except zlib.error:
        pass

    # Return raw bytes as string if decompression fails
    return raw.decode("utf-8", errors="replace")


def encode_diagram_content(xml: str) -> str:
    """
    Encode mxGraphModel XML to DrawIO format for embedding in PNG files.

    Args:
        xml: The mxGraphModel XML string

    Returns:
        Encoded string suitable for <diagram> element content

    Note:
        DrawIO format for files: deflate (raw) → base64
        No URL encoding for file storage - DrawIO handles this internally.
    """
    # Step 1: Deflate compress (raw deflate, no zlib header)
    compress_obj = zlib.compressobj(9, zlib.DEFLATED, -zlib.MAX_WBITS)
    compressed = compress_obj.compress(xml.encode("utf-8"))
    compressed += compress_obj.flush()

    # Step 2: Base64 encode only (no URL encoding for file format)
    b64 = base64.b64encode(compressed).decode("ascii")

    return b64


def decode_mxfile(mxfile_xml: str) -> str:
    """
    Fully decode an mxfile XML, expanding all diagram contents.

    Args:
        mxfile_xml: The raw mxfile XML (as extracted from PNG)

    Returns:
        Decoded mxfile XML with all diagrams expanded to readable mxGraphModel
    """
    try:
        root = ET.fromstring(mxfile_xml)
    except ET.ParseError as e:
        raise ValueError(f"Invalid mxfile XML: {e}")

    if root.tag != "mxfile":
        raise ValueError(f"Expected mxfile root element, got: {root.tag}")

    # Process each diagram
    for diagram in root.findall(".//diagram"):
        # Get the encoded content (text content of diagram element)
        encoded_content = diagram.text
        if encoded_content and encoded_content.strip():
            # Decode the content
            decoded_xml = decode_diagram_content(encoded_content.strip())

            # Parse the decoded XML and insert as child element
            try:
                graph_model = ET.fromstring(decoded_xml)
                diagram.text = None  # Clear the encoded text
                diagram.append(graph_model)
            except ET.ParseError:
                # If parsing fails, store as text with marker
                diagram.text = decoded_xml
                diagram.set("_decode_failed", "true")

    # Convert back to string with proper formatting
    return _prettify_xml(root)


def encode_mxfile(decoded_xml: str) -> str:
    """
    Encode a decoded mxfile XML back to DrawIO format.

    Args:
        decoded_xml: The decoded mxfile XML with expanded diagram contents

    Returns:
        Encoded mxfile XML ready for embedding in PNG
    """
    try:
        root = ET.fromstring(decoded_xml)
    except ET.ParseError as e:
        raise ValueError(f"Invalid mxfile XML: {e}")

    if root.tag != "mxfile":
        raise ValueError(f"Expected mxfile root element, got: {root.tag}")

    # Remove all attributes from mxfile for maximum DrawIO compatibility
    # DrawIO seems to have issues with URL-encoded attributes in some cases
    root.attrib.clear()

    # Process each diagram - keep mxGraphModel as uncompressed XML
    # DrawIO works better with uncompressed content when URL-encoded in PNG
    for diagram in root.findall(".//diagram"):
        graph_model = diagram.find("mxGraphModel")
        if graph_model is not None:
            # Keep mxGraphModel as child element (uncompressed format)
            # This matches how DrawIO saves files after editing
            pass  # Leave as-is
        elif diagram.text and diagram.text.strip():
            # Text content - check if it's encoded or raw XML
            content = diagram.text.strip()
            if not content.startswith("<"):
                # Encoded content - decode it and add as child element
                try:
                    decoded_xml = decode_diagram_content(content)
                    graph_model = ET.fromstring(decoded_xml)
                    diagram.text = None
                    diagram.append(graph_model)
                except Exception:
                    pass  # Leave as-is if decode fails

    # Convert back to string with proper formatting
    return _prettify_xml(root)


def _escape_attr(value: str) -> str:
    """Escape special characters in XML attribute values."""
    value = value.replace("&", "&amp;")
    value = value.replace("<", "&lt;")
    value = value.replace(">", "&gt;")
    value = value.replace('"', "&quot;")
    value = value.replace("\n", "&#xa;")
    return value


def _prettify_xml(elem: ET.Element, level: int = 0) -> str:
    """
    Pretty-print an XML element with proper indentation.
    """
    indent = "  "
    result = []

    # Opening tag - escape attribute values properly
    attribs = " ".join(f'{k}="{_escape_attr(v)}"' for k, v in elem.attrib.items())
    if attribs:
        result.append(f"{indent * level}<{elem.tag} {attribs}")
    else:
        result.append(f"{indent * level}<{elem.tag}")

    # Check if has children or text
    has_children = len(elem) > 0
    has_text = elem.text and elem.text.strip()

    if has_children:
        result.append(">\n")
        for child in elem:
            result.append(_prettify_xml(child, level + 1))
        result.append(f"{indent * level}</{elem.tag}>\n")
    elif has_text:
        result.append(f">{elem.text}</{elem.tag}>\n")
    else:
        result.append("/>\n")

    return "".join(result)


def create_blank_mxfile(name: str = "Page-1") -> str:
    """
    Create a blank mxfile XML structure with EAC visual defaults.

    Args:
        name: Name for the first diagram page

    Returns:
        Decoded mxfile XML with empty diagram using EAC styling
    """
    import uuid
    diagram_id = str(uuid.uuid4())[:8]

    # EAC visual defaults:
    # - background="#CFCFCF" (gray background)
    # - shadow="1" (drop shadows enabled)
    # - Standard page dimensions for technical diagrams
    return f'''<mxfile host="drawio-oci" agent="drawio-oci (Claude)" version="1.0">
  <diagram name="{name}" id="{diagram_id}">
    <mxGraphModel dx="1426" dy="758" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="1654" pageHeight="1169" background="#CFCFCF" shadow="1">
      <root>
        <mxCell id="0"/>
        <mxCell id="1" parent="0"/>
      </root>
    </mxGraphModel>
  </diagram>
</mxfile>'''


def get_mxfile_info(mxfile_xml: str) -> dict:
    """
    Extract metadata from an mxfile XML.

    Args:
        mxfile_xml: The mxfile XML (encoded or decoded)

    Returns:
        Dictionary with metadata
    """
    try:
        root = ET.fromstring(mxfile_xml)
    except ET.ParseError as e:
        return {"error": str(e)}

    if root.tag != "mxfile":
        return {"error": f"Expected mxfile root, got {root.tag}"}

    info = {
        "host": root.get("host", "unknown"),
        "agent": root.get("agent", "unknown"),
        "version": root.get("version", "unknown"),
        "modified": root.get("modified", "unknown"),
        "diagrams": [],
    }

    for diagram in root.findall(".//diagram"):
        diag_info = {
            "name": diagram.get("name", "unnamed"),
            "id": diagram.get("id", ""),
        }

        # Try to count cells
        graph_model = diagram.find("mxGraphModel")
        if graph_model is not None:
            cells = graph_model.findall(".//mxCell")
            diag_info["cells"] = len(cells)
        else:
            # Content is still encoded, try to decode and count
            if diagram.text and diagram.text.strip():
                try:
                    decoded = decode_diagram_content(diagram.text.strip())
                    temp_root = ET.fromstring(decoded)
                    cells = temp_root.findall(".//mxCell")
                    diag_info["cells"] = len(cells)
                except Exception:
                    diag_info["cells"] = -1  # Unknown

        info["diagrams"].append(diag_info)

    return info


if __name__ == "__main__":
    # Simple test
    blank = create_blank_mxfile("Test")
    print("Created blank mxfile:")
    print(blank)
    print()

    # Encode it
    encoded = encode_mxfile(blank)
    print("Encoded:")
    print(encoded[:200] + "...")
    print()

    # Decode it back
    decoded = decode_mxfile(encoded)
    print("Decoded back:")
    print(decoded)
