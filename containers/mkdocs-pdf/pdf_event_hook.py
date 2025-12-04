"""
PDF Event Hook for mkdocs-with-pdf

This hook fixes WeasyPrint's file:// URL handling issues in large documents:
1. Embeds external CSS inline (fixes image embedding bug)
2. Embeds images as base64 data URIs (fixes image loading failures)

WeasyPrint fails to properly handle file:// URLs in documents exceeding
~18000 lines, causing images to not render or embed correctly.
"""

import os
import base64
import mimetypes
import logging
from bs4 import BeautifulSoup, Tag


def pre_pdf_render(soup: BeautifulSoup, logger: logging.Logger) -> BeautifulSoup:
    """
    Called before WeasyPrint renders the PDF.
    Embeds external CSS and images inline to fix WeasyPrint issues.
    """
    logger.info("PDF Event Hook: Preprocessing for WeasyPrint compatibility")

    # Debug: Count images and sample different types
    all_imgs = soup.find_all('img')
    logger.info(f"  DEBUG: Total img tags in soup: {len(all_imgs)}")

    # Sample different src patterns
    src_patterns = {}
    for img in all_imgs[:200]:
        src = img.get('src', '')[:80]
        pattern = src.split('/')[0] if '/' in src else src[:30]
        src_patterns[pattern] = src_patterns.get(pattern, 0) + 1
    logger.info(f"  DEBUG: Src pattern counts: {src_patterns}")

    # First embed images (before CSS might reference them)
    embed_images(soup, logger)

    # Then embed CSS
    logger.info("  Embedding CSS inline...")

    # Find all stylesheet link tags
    link_tags = soup.find_all('link', rel='stylesheet')
    if not link_tags:
        logger.info("  No stylesheet links found")
        return soup

    # Collect CSS content
    css_parts = []
    links_to_remove = []

    for link in link_tags:
        href = link.get('href', '')
        if not href:
            continue

        css_path = resolve_css_path(href)
        if css_path and os.path.exists(css_path):
            try:
                with open(css_path, 'r', encoding='utf-8') as f:
                    css_content = f.read()
                    css_parts.append(f"/* Embedded from: {os.path.basename(css_path)} */")
                    css_parts.append(css_content)
                    css_parts.append("")
                    links_to_remove.append(link)
                    logger.info(f"  Embedded: {css_path}")
            except Exception as e:
                logger.warning(f"  Failed to read {css_path}: {e}")

    if not css_parts:
        logger.info("  No CSS files embedded")
        return soup

    # Remove the link tags
    for link in links_to_remove:
        link.decompose()

    # Create a style tag with all CSS
    style_tag = soup.new_tag('style')
    style_tag.string = '\n'.join(css_parts)

    # Insert the style tag in the head
    head = soup.find('head')
    if head:
        head.append(style_tag)
        logger.info(f"  Embedded {len(links_to_remove)} CSS files inline")
    else:
        logger.warning("  Could not find <head> tag")

    return soup


def resolve_css_path(href: str) -> str:
    """Resolve a CSS href to an absolute file path."""
    if href.startswith('file://'):
        path = href[7:]  # Remove file://
        # Handle Unix absolute paths
        if path.startswith('/') and os.path.exists(path):
            return path
        # Handle Windows paths (file:///C:/...)
        path = path.lstrip('/')
        if os.path.exists(path):
            return path
        # Try with /docs prefix (Docker mount)
        docker_path = '/docs/' + path
        if os.path.exists(docker_path):
            return docker_path
        return None

    # Skip http/https URLs
    if href.startswith(('http://', 'https://')):
        return None

    # Relative path - try from /docs (Docker mount point)
    docker_path = os.path.join('/docs', href)
    if os.path.exists(docker_path):
        return docker_path

    return None


def resolve_file_path(url: str) -> str:
    """Resolve a file:// URL to an absolute file path."""
    if not url.startswith('file://'):
        return None

    path = url[7:]  # Remove file://

    # Handle Unix absolute paths
    if path.startswith('/') and os.path.exists(path):
        return path

    # Handle Windows paths (file:///C:/...)
    path = path.lstrip('/')
    if os.path.exists(path):
        return path

    # Try with /docs prefix (Docker mount)
    docker_path = '/docs/' + path
    if os.path.exists(docker_path):
        return docker_path

    return None


def embed_images(soup: BeautifulSoup, logger: logging.Logger) -> None:
    """
    Embed images as base64 data URIs.

    WeasyPrint fails to load file:// URLs and relative paths for images in
    large documents. This function converts all image references to inline base64.
    """
    logger.info("  Embedding images as base64 data URIs...")

    img_tags = soup.find_all('img')
    logger.info(f"    Found {len(img_tags)} img tags total")
    embedded_count = 0
    skipped_count = 0
    failed_count = 0

    # Debug: Sample the first few image sources
    sample_srcs = [img.get('src', '')[:100] for img in img_tags[:10]]
    logger.info(f"    Sample src values: {sample_srcs}")

    for img in img_tags:
        src = img.get('src', '')
        if not src:
            continue

        # Skip already-embedded data URIs
        if src.startswith('data:'):
            skipped_count += 1
            continue

        # Skip external URLs
        if src.startswith(('http://', 'https://')):
            skipped_count += 1
            continue

        # Resolve the image path
        img_path = None

        if src.startswith('file://'):
            img_path = resolve_file_path(src)
        else:
            # Relative path - resolve from /docs (Docker mount point for site)
            # The print_page.html is in /docs/out/build/docs/site/pdf/
            # Images are relative from there, e.g., ../../../../assets/...
            # So we resolve from the site directory
            site_dir = '/docs/out/build/docs/site'
            pdf_dir = os.path.join(site_dir, 'pdf')
            img_path = os.path.normpath(os.path.join(pdf_dir, src))

            if not os.path.exists(img_path):
                # Also try from site root directly
                img_path = os.path.join(site_dir, src.lstrip('/'))

        if not img_path or not os.path.exists(img_path):
            logger.debug(f"    Image not found: {src}")
            failed_count += 1
            continue

        try:
            # Read the image file
            with open(img_path, 'rb') as f:
                img_data = f.read()

            # Determine MIME type
            mime_type, _ = mimetypes.guess_type(img_path)
            if not mime_type:
                # Default to PNG for unknown types
                mime_type = 'image/png'

            # Convert to base64 data URI
            b64_data = base64.b64encode(img_data).decode('utf-8')
            data_uri = f"data:{mime_type};base64,{b64_data}"

            # Replace the src attribute
            img['src'] = data_uri
            embedded_count += 1

        except Exception as e:
            logger.warning(f"    Failed to embed {img_path}: {e}")
            failed_count += 1

    logger.info(f"    Embedded {embedded_count} images, {skipped_count} skipped, {failed_count} failed")
