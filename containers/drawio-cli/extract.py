#!/usr/bin/env python3
"""
Extract embedded XML diagram data from drawio.png files.

Drawio embeds diagram data in PNG files using the tEXt chunk with key "mxfile".
The data is URL-encoded and may be deflate-compressed.
"""

import base64
import struct
import urllib.parse
import zlib
from pathlib import Path


def read_png_chunks(filepath: str) -> list[tuple[str, bytes]]:
    """Read all chunks from a PNG file."""
    chunks = []
    with open(filepath, "rb") as f:
        # Verify PNG signature
        sig = f.read(8)
        if sig != b"\x89PNG\r\n\x1a\n":
            raise ValueError(f"Not a valid PNG file: {filepath}")

        while True:
            # Read chunk length (4 bytes, big-endian)
            length_data = f.read(4)
            if len(length_data) < 4:
                break

            length = struct.unpack(">I", length_data)[0]

            # Read chunk type (4 bytes)
            chunk_type = f.read(4).decode("ascii")

            # Read chunk data
            data = f.read(length)

            # Read CRC (4 bytes) - we skip validation
            f.read(4)

            chunks.append((chunk_type, data))

            if chunk_type == "IEND":
                break

    return chunks


def extract_mxfile_from_text(data: bytes) -> str | None:
    """Extract mxfile data from a tEXt chunk."""
    # tEXt format: keyword\x00text
    try:
        null_idx = data.index(b"\x00")
        keyword = data[:null_idx].decode("latin-1")
        text = data[null_idx + 1 :].decode("latin-1")

        if keyword == "mxfile":
            return text
    except (ValueError, UnicodeDecodeError):
        pass
    return None


def extract_mxfile_from_itxt(data: bytes) -> str | None:
    """Extract mxfile data from an iTXt chunk."""
    # iTXt format: keyword\x00compression_flag\x00compression_method\x00language_tag\x00translated_keyword\x00text
    try:
        parts = data.split(b"\x00", 5)
        if len(parts) < 6:
            return None

        keyword = parts[0].decode("latin-1")
        if keyword != "mxfile":
            return None

        compression_flag = parts[1][0] if parts[1] else 0
        text_data = parts[5]

        if compression_flag:
            text_data = zlib.decompress(text_data)

        return text_data.decode("utf-8")
    except (ValueError, UnicodeDecodeError, zlib.error):
        pass
    return None


def extract_mxfile_from_ztxt(data: bytes) -> str | None:
    """Extract mxfile data from a zTXt chunk."""
    # zTXt format: keyword\x00compression_method\x00compressed_text
    try:
        null_idx = data.index(b"\x00")
        keyword = data[:null_idx].decode("latin-1")

        if keyword != "mxfile":
            return None

        # compression_method = data[null_idx + 1]  # Should be 0 (deflate)
        compressed_data = data[null_idx + 2 :]
        text = zlib.decompress(compressed_data).decode("utf-8")
        return text
    except (ValueError, UnicodeDecodeError, zlib.error):
        pass
    return None


def decode_drawio_data(encoded: str) -> str:
    """Decode drawio diagram data (URL-encoded, possibly base64+deflate)."""
    # First, URL-decode
    decoded = urllib.parse.unquote(encoded)

    # Check if it starts with mxfile XML
    if decoded.startswith("<mxfile") or decoded.startswith("<?xml"):
        return decoded

    # Try base64 + deflate decode (used in some versions)
    try:
        raw = base64.b64decode(decoded)
        # Try raw inflate (no zlib header)
        try:
            xml = zlib.decompress(raw, -zlib.MAX_WBITS).decode("utf-8")
            return xml
        except zlib.error:
            pass
        # Try with zlib header
        try:
            xml = zlib.decompress(raw).decode("utf-8")
            return xml
        except zlib.error:
            pass
    except Exception:
        pass

    # Return as-is if nothing worked
    return decoded


def extract_drawio_xml(filepath: str) -> str | None:
    """
    Extract drawio XML from a .drawio.png file.

    Args:
        filepath: Path to the drawio.png file

    Returns:
        The extracted XML string, or None if no drawio data found
    """
    chunks = read_png_chunks(filepath)

    mxfile_data = None

    for chunk_type, data in chunks:
        if chunk_type == "tEXt":
            result = extract_mxfile_from_text(data)
            if result:
                mxfile_data = result
                break
        elif chunk_type == "iTXt":
            result = extract_mxfile_from_itxt(data)
            if result:
                mxfile_data = result
                break
        elif chunk_type == "zTXt":
            result = extract_mxfile_from_ztxt(data)
            if result:
                mxfile_data = result
                break

    if mxfile_data:
        return decode_drawio_data(mxfile_data)

    return None


def main():
    import argparse

    parser = argparse.ArgumentParser(
        description="Extract embedded XML from drawio.png files"
    )
    parser.add_argument("input", help="Input drawio.png file")
    parser.add_argument(
        "-o", "--output", help="Output file (default: stdout)", default=None
    )
    parser.add_argument(
        "-f",
        "--format",
        choices=["xml", "drawio"],
        default="xml",
        help="Output format (default: xml)",
    )

    args = parser.parse_args()

    xml = extract_drawio_xml(args.input)
    if xml is None:
        print(f"Error: No drawio data found in {args.input}", file=__import__("sys").stderr)
        exit(1)

    if args.output:
        Path(args.output).write_text(xml, encoding="utf-8")
        print(f"Extracted to {args.output}")
    else:
        print(xml)


if __name__ == "__main__":
    main()
