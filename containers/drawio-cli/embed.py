#!/usr/bin/env python3
"""
DrawIO PNG embedding - Embed/extract XML metadata in PNG files.

PNG files consist of chunks: signature + IHDR + optional chunks + IDAT(s) + IEND
DrawIO stores diagram XML in a tEXt chunk with keyword "mxfile".

This module handles:
- Reading PNG chunks
- Writing PNG chunks
- Embedding mxfile XML into PNG
- Creating new .drawio.png files with minimal PNG image
"""

import struct
import zlib
from pathlib import Path
from typing import Optional

# Minimal 1x1 transparent PNG (used for new files)
# This is a valid PNG that DrawIO will accept and render from the embedded XML
MINIMAL_PNG_CHUNKS = [
    # IHDR: 1x1, 8-bit RGBA
    (b"IHDR", struct.pack(">IIBBBBB", 1, 1, 8, 6, 0, 0, 0)),
    # IDAT: compressed single transparent pixel
    (b"IDAT", zlib.compress(b"\x00\x00\x00\x00\x00", 9)),
    # IEND: end marker
    (b"IEND", b""),
]

PNG_SIGNATURE = b"\x89PNG\r\n\x1a\n"


def read_png_chunks(filepath: str) -> list[tuple[bytes, bytes]]:
    """
    Read all chunks from a PNG file.

    Args:
        filepath: Path to PNG file

    Returns:
        List of (chunk_type, chunk_data) tuples
    """
    chunks = []
    with open(filepath, "rb") as f:
        # Verify PNG signature
        sig = f.read(8)
        if sig != PNG_SIGNATURE:
            raise ValueError(f"Not a valid PNG file: {filepath}")

        while True:
            # Read chunk length (4 bytes, big-endian)
            length_data = f.read(4)
            if len(length_data) < 4:
                break

            length = struct.unpack(">I", length_data)[0]

            # Read chunk type (4 bytes)
            chunk_type = f.read(4)

            # Read chunk data
            data = f.read(length)

            # Read CRC (4 bytes) - we recalculate on write
            f.read(4)

            chunks.append((chunk_type, data))

            if chunk_type == b"IEND":
                break

    return chunks


def write_png_chunks(filepath: str, chunks: list[tuple[bytes, bytes]]) -> None:
    """
    Write chunks to a PNG file.

    Args:
        filepath: Output path
        chunks: List of (chunk_type, chunk_data) tuples
    """
    with open(filepath, "wb") as f:
        # Write PNG signature
        f.write(PNG_SIGNATURE)

        for chunk_type, data in chunks:
            # Write length
            f.write(struct.pack(">I", len(data)))

            # Write type
            f.write(chunk_type)

            # Write data
            f.write(data)

            # Calculate and write CRC (over type + data)
            crc = zlib.crc32(chunk_type + data) & 0xFFFFFFFF
            f.write(struct.pack(">I", crc))


def create_text_chunk(keyword: str, text: str) -> tuple[bytes, bytes]:
    """
    Create a tEXt chunk for PNG.

    Args:
        keyword: The keyword (e.g., "mxfile")
        text: The text content (will be URL-encoded for DrawIO compatibility)

    Returns:
        (chunk_type, chunk_data) tuple
    """
    import urllib.parse
    # DrawIO expects URL-encoded XML in the mxfile tEXt chunk
    # This is required for DrawIO to parse the content correctly
    encoded_text = urllib.parse.quote(text, safe='')
    # tEXt format: keyword + null + text
    data = keyword.encode("latin-1") + b"\x00" + encoded_text.encode("latin-1")
    return (b"tEXt", data)


def extract_mxfile_chunk(chunks: list[tuple[bytes, bytes]]) -> Optional[str]:
    """
    Extract mxfile content from PNG chunks.

    Args:
        chunks: List of PNG chunks

    Returns:
        The mxfile XML string, or None if not found
    """
    import urllib.parse
    for chunk_type, data in chunks:
        if chunk_type == b"tEXt":
            try:
                null_idx = data.index(b"\x00")
                keyword = data[:null_idx].decode("latin-1")
                if keyword == "mxfile":
                    text = data[null_idx + 1:].decode("latin-1")
                    # URL-decode if needed (DrawIO stores URL-encoded XML)
                    if text.startswith('%3C'):
                        text = urllib.parse.unquote(text)
                    return text
            except (ValueError, UnicodeDecodeError):
                continue
        elif chunk_type == b"zTXt":
            try:
                null_idx = data.index(b"\x00")
                keyword = data[:null_idx].decode("latin-1")
                if keyword == "mxfile":
                    # compression_method = data[null_idx + 1]
                    compressed = data[null_idx + 2:]
                    return zlib.decompress(compressed).decode("utf-8")
            except (ValueError, UnicodeDecodeError, zlib.error):
                continue
        elif chunk_type == b"iTXt":
            try:
                parts = data.split(b"\x00", 5)
                if len(parts) >= 6:
                    keyword = parts[0].decode("latin-1")
                    if keyword == "mxfile":
                        compression_flag = parts[1][0] if parts[1] else 0
                        text_data = parts[5]
                        if compression_flag:
                            text_data = zlib.decompress(text_data)
                        return text_data.decode("utf-8")
            except (ValueError, UnicodeDecodeError, zlib.error):
                continue

    return None


def remove_mxfile_chunks(chunks: list[tuple[bytes, bytes]]) -> list[tuple[bytes, bytes]]:
    """
    Remove all mxfile-related text chunks from PNG chunks.

    Args:
        chunks: List of PNG chunks

    Returns:
        Filtered list without mxfile chunks
    """
    result = []
    for chunk_type, data in chunks:
        is_mxfile = False

        if chunk_type in (b"tEXt", b"zTXt"):
            try:
                null_idx = data.index(b"\x00")
                keyword = data[:null_idx].decode("latin-1")
                if keyword == "mxfile":
                    is_mxfile = True
            except (ValueError, UnicodeDecodeError):
                pass
        elif chunk_type == b"iTXt":
            try:
                null_idx = data.index(b"\x00")
                keyword = data[:null_idx].decode("latin-1")
                if keyword == "mxfile":
                    is_mxfile = True
            except (ValueError, UnicodeDecodeError):
                pass

        if not is_mxfile:
            result.append((chunk_type, data))

    return result


def embed_xml_in_png(
    png_path: Optional[str],
    xml: str,
    output_path: str,
) -> None:
    """
    Embed mxfile XML into a PNG file.

    Args:
        png_path: Source PNG file (or None to create new with minimal PNG)
        xml: The mxfile XML to embed (should be encoded format)
        output_path: Output path for the .drawio.png
    """
    if png_path and Path(png_path).exists():
        # Read existing PNG
        chunks = read_png_chunks(png_path)
    else:
        # Use minimal PNG for new file
        chunks = list(MINIMAL_PNG_CHUNKS)

    # Remove existing mxfile chunks
    chunks = remove_mxfile_chunks(chunks)

    # Create new mxfile chunk
    mxfile_chunk = create_text_chunk("mxfile", xml)

    # Insert after IHDR (position 1, since IHDR is at 0)
    # Find IHDR position
    ihdr_idx = next(
        (i for i, (t, _) in enumerate(chunks) if t == b"IHDR"),
        0
    )

    # Insert mxfile chunk after IHDR
    chunks.insert(ihdr_idx + 1, mxfile_chunk)

    # Ensure output directory exists
    Path(output_path).parent.mkdir(parents=True, exist_ok=True)

    # Write output
    write_png_chunks(output_path, chunks)


def create_drawio_png(xml: str, output_path: str) -> None:
    """
    Create a new .drawio.png file from mxfile XML.

    Args:
        xml: The mxfile XML (encoded format)
        output_path: Output path for the .drawio.png
    """
    embed_xml_in_png(None, xml, output_path)


if __name__ == "__main__":
    import sys

    if len(sys.argv) < 2:
        print("Usage: embed.py <command> [args...]")
        print("Commands:")
        print("  create <output.drawio.png> <input.xml>  - Create new drawio.png")
        print("  update <file.drawio.png> <input.xml>    - Update existing drawio.png")
        print("  extract <file.drawio.png>               - Extract mxfile XML")
        sys.exit(1)

    cmd = sys.argv[1]

    if cmd == "create":
        if len(sys.argv) < 4:
            print("Usage: embed.py create <output.drawio.png> <input.xml>")
            sys.exit(1)
        output_path = sys.argv[2]
        xml_path = sys.argv[3]
        xml = Path(xml_path).read_text(encoding="utf-8")
        create_drawio_png(xml, output_path)
        print(f"Created: {output_path}")

    elif cmd == "update":
        if len(sys.argv) < 4:
            print("Usage: embed.py update <file.drawio.png> <input.xml>")
            sys.exit(1)
        png_path = sys.argv[2]
        xml_path = sys.argv[3]
        xml = Path(xml_path).read_text(encoding="utf-8")
        embed_xml_in_png(png_path, xml, png_path)
        print(f"Updated: {png_path}")

    elif cmd == "extract":
        if len(sys.argv) < 3:
            print("Usage: embed.py extract <file.drawio.png>")
            sys.exit(1)
        png_path = sys.argv[2]
        chunks = read_png_chunks(png_path)
        xml = extract_mxfile_chunk(chunks)
        if xml:
            print(xml)
        else:
            print("No mxfile data found", file=sys.stderr)
            sys.exit(1)

    else:
        print(f"Unknown command: {cmd}", file=sys.stderr)
        sys.exit(1)
