// Package mkdocs contains extracted functionality from the mkdocs builder.
package mkdocs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// mergePDFs merges multiple PDF files with cover page, TOC, and page numbers.
// This function contains an embedded Python script that uses pypdf and reportlab
// to generate a professional PDF with navigation, bookmarks, and consistent styling.
func MergePDFs(siteDir, outputPath, hostRepoRoot, workspaceRoot, stagingDir, imageName string, bookTitle string, bookDescription string, logWriter io.Writer, isDinD bool) error {
	// Get relative paths for Docker
	relSiteDir, err := filepath.Rel(workspaceRoot, siteDir)
	if err != nil {
		return fmt.Errorf("calculating relative site dir: %w", err)
	}
	relOutputPath, err := filepath.Rel(workspaceRoot, outputPath)
	if err != nil {
		return fmt.Errorf("calculating relative output path: %w", err)
	}
	relStagingDir, err := filepath.Rel(workspaceRoot, stagingDir)
	if err != nil {
		return fmt.Errorf("calculating relative staging dir: %w", err)
	}

	// Convert to Docker paths
	dockerSiteDir := "/docs/" + strings.ReplaceAll(relSiteDir, "\\", "/")
	dockerOutputPath := "/docs/" + strings.ReplaceAll(relOutputPath, "\\", "/")
	dockerStagingDir := "/docs/" + strings.ReplaceAll(relStagingDir, "\\", "/")

	volumeMountPath := hostRepoRoot
	dockerVolume := FormatDockerVolumePath(volumeMountPath)

	args := []string{
		"run", "--rm",
		"-v", dockerVolume + ":/docs",
		"-w", "/docs",
	}

	if isDinD {
		uid := os.Getuid()
		gid := os.Getgid()
		args = append(args, "--user", fmt.Sprintf("%d:%d", uid, gid))
	}

	// Python script to generate cover page, TOC, and merge all PDFs
	// Parses .nav.yml files for proper titles and ordering
	// Escape book title and description for Python (handle quotes and newlines)
	escapedTitle := strings.ReplaceAll(strings.ReplaceAll(bookTitle, "\\", "\\\\"), "'", "\\'")
	escapedDescription := strings.ReplaceAll(strings.ReplaceAll(bookDescription, "\\", "\\\\"), "'", "\\'")

	pythonScript := fmt.Sprintf(`
import os
import io
import yaml
from datetime import datetime
from pathlib import Path
from pypdf import PdfWriter, PdfReader
from reportlab.lib.pagesizes import A4
from reportlab.lib.colors import HexColor
from reportlab.pdfgen import canvas
from reportlab.lib.units import cm

site_dir = '%s'
output_path = '%s'
docs_dir = '%s'  # Staging directory with .nav.yml files
book_title = '''%s'''  # Book-specific title for cover page
book_description = '''%s'''  # Book-specific description for cover page

import re

def get_title_from_markdown(md_path):
    """Extract title from markdown file (frontmatter or first H1)."""
    if not os.path.exists(md_path):
        return None

    try:
        with open(md_path, 'r', encoding='utf-8') as f:
            content = f.read()
    except:
        return None

    # Try frontmatter title
    if content.startswith('---'):
        end_idx = content.find('---', 3)
        if end_idx > 0:
            frontmatter = content[3:end_idx]
            match = re.search(r'^title:\s*["\']?([^"\'\n]+)["\']?', frontmatter, re.MULTILINE)
            if match:
                return match.group(1).strip()

    # Try first H1
    match = re.search(r'^#\s+(.+)$', content, re.MULTILINE)
    if match:
        return match.group(1).strip()

    return None

# Dark theme colors (matching PDF dark theme)
BG_COLOR = HexColor('#0d1117')
TEXT_COLOR = HexColor('#e6edf3')
ACCENT_COLOR = HexColor('#58a6ff')
MUTED_COLOR = HexColor('#8b949e')
FOOTER_COLOR = HexColor('#6e7681')

def create_page_footer_overlay(page_num, total_pages, book_title, section_title=''):
    """Create a PDF page with footer overlay (page number, book title, section)."""
    packet = io.BytesIO()
    c = canvas.Canvas(packet, pagesize=A4)
    width, height = A4

    # Footer Y position (2cm from bottom = ~56 points)
    footer_y = 1.5 * cm

    # Page number - center
    c.setFont('Helvetica', 9)
    c.setFillColor(MUTED_COLOR)
    page_text = str(page_num)
    c.drawCentredString(width / 2, footer_y, page_text)

    # Book title - left (truncate to fit available width)
    c.setFont('Helvetica', 8)
    c.setFillColor(FOOTER_COLOR)
    max_book_title_width = (width / 2) - 3*cm
    display_title = truncate_text_to_width(c, book_title, 'Helvetica', 8, max_book_title_width)
    c.drawString(2 * cm, footer_y, display_title)

    # Section title - right (truncate to fit available width)
    if section_title:
        max_section_width = (width / 2) - 1*cm
        display_section = truncate_text_to_width(c, section_title, 'Helvetica', 8, max_section_width)
        c.drawRightString(width - 2 * cm, footer_y, display_section)

    c.save()
    packet.seek(0)
    return PdfReader(packet)

def draw_wrapped_text(canvas, text, x, y, max_width, font='Helvetica', size=12, line_height=0.5*cm, align='center'):
    """Draw text with word wrapping."""
    canvas.setFont(font, size)
    words = text.split()
    lines = []
    current_line = []

    for word in words:
        test_line = ' '.join(current_line + [word])
        if canvas.stringWidth(test_line, font, size) <= max_width:
            current_line.append(word)
        else:
            if current_line:
                lines.append(' '.join(current_line))
            current_line = [word]

    if current_line:
        lines.append(' '.join(current_line))

    # Draw lines centered or left-aligned
    for i, line in enumerate(lines):
        if align == 'center':
            canvas.drawCentredString(x, y - (i * line_height), line)
        else:
            canvas.drawString(x, y - (i * line_height), line)

    return len(lines)  # Return number of lines drawn

def truncate_text_to_width(canvas, text, font, size, max_width):
    """Truncate text to fit within max_width, preferring word boundaries.

    Args:
        canvas: ReportLab canvas object
        text: Text to truncate
        font: Font name (e.g., 'Helvetica')
        size: Font size in points
        max_width: Maximum width in points

    Returns:
        Truncated text with '...' if needed, or original text if it fits
    """
    # Check if full text fits
    if canvas.stringWidth(text, font, size) <= max_width:
        return text

    # Reserve space for ellipsis
    ellipsis = '...'
    ellipsis_width = canvas.stringWidth(ellipsis, font, size)
    available_width = max_width - ellipsis_width

    # Try word-boundary truncation first
    words = text.split()
    if len(words) > 1:
        truncated = []
        for word in words:
            test_text = ' '.join(truncated + [word])
            if canvas.stringWidth(test_text, font, size) <= available_width:
                truncated.append(word)
            else:
                break

        if truncated:
            return ' '.join(truncated) + ellipsis

    # Fallback: character-by-character truncation for single long word
    for i in range(len(text), 0, -1):
        truncated = text[:i]
        if canvas.stringWidth(truncated, font, size) <= available_width:
            return truncated + ellipsis

    # Edge case: even one character is too wide
    return ellipsis

def create_cover_page():
    """Create a cover page PDF with title, subtitle, and metadata."""
    buffer = io.BytesIO()
    c = canvas.Canvas(buffer, pagesize=A4)
    width, height = A4

    # Dark background
    c.setFillColor(BG_COLOR)
    c.rect(0, 0, width, height, fill=True, stroke=False)

    # Brand Title (top, smaller)
    c.setFillColor(TEXT_COLOR)
    c.setFont('Helvetica-Bold', 32)
    c.drawCentredString(width/2, height - 6*cm, 'Ready-to-Release')

    # Brand Subtitle
    c.setFont('Helvetica', 20)
    c.drawCentredString(width/2, height - 7.5*cm, 'Documentation')

    # Book-specific title (larger, prominent, accent color)
    # Auto-size or wrap if title is too long
    if book_title:
        c.setFillColor(ACCENT_COLOR)
        max_title_width = width - 4*cm  # Leave 2cm margin on each side

        # Try different font sizes to fit title on one line
        title_font_size = 28
        c.setFont('Helvetica-Bold', title_font_size)
        title_width = c.stringWidth(book_title, 'Helvetica-Bold', title_font_size)

        if title_width > max_title_width:
            # Title too long, try smaller font
            title_font_size = 24
            c.setFont('Helvetica-Bold', title_font_size)
            title_width = c.stringWidth(book_title, 'Helvetica-Bold', title_font_size)

            if title_width > max_title_width:
                # Still too long, try even smaller
                title_font_size = 20
                c.setFont('Helvetica-Bold', title_font_size)
                title_width = c.stringWidth(book_title, 'Helvetica-Bold', title_font_size)

                if title_width > max_title_width:
                    # Still too long, wrap it
                    draw_wrapped_text(c, book_title, width/2, height - 10*cm, max_title_width, 'Helvetica-Bold', 18, 0.6*cm, 'center')
                else:
                    c.drawCentredString(width/2, height - 10*cm, book_title)
            else:
                c.drawCentredString(width/2, height - 10*cm, book_title)
        else:
            c.drawCentredString(width/2, height - 10*cm, book_title)

        c.setFillColor(TEXT_COLOR)  # Reset color

    # Horizontal line
    c.setStrokeColor(ACCENT_COLOR)
    c.setLineWidth(2)
    c.line(4*cm, height - 11.5*cm, width - 4*cm, height - 11.5*cm)

    # Brand description
    c.setFillColor(MUTED_COLOR)
    c.setFont('Helvetica', 14)
    c.drawCentredString(width/2, height - 13*cm, 'Everything-as-Code Platform')
    c.drawCentredString(width/2, height - 14*cm, 'for Software Delivery Flows')

    # Book-specific description (wrapped if needed)
    current_y = height - 16*cm
    if book_description:
        c.setFont('Helvetica', 12)
        c.setFillColor(TEXT_COLOR)
        lines_drawn = draw_wrapped_text(c, book_description, width/2, current_y, width - 8*cm, 'Helvetica', 12, 0.5*cm, 'center')
        current_y -= lines_drawn * 0.5*cm

    # Date at bottom
    c.setFillColor(MUTED_COLOR)
    c.setFont('Helvetica', 11)
    date_str = datetime.now().strftime('Generated: %%B %%d, %%Y')
    c.drawCentredString(width/2, 3*cm, date_str)

    c.save()
    buffer.seek(0)
    return PdfReader(buffer)

def create_toc_pages(toc_entries, content_start_page):
    """Create table of contents pages with page numbers and clickable links."""
    buffer = io.BytesIO()
    c = canvas.Canvas(buffer, pagesize=A4)
    width, height = A4

    # Track current y position and page
    y = height - 3*cm
    entries_per_page = 35
    entry_count = 0
    # Store link info: (toc_page, x, y, width, height, target_page)
    links = []
    current_toc_page = 0

    def start_new_page():
        nonlocal y, current_toc_page
        c.showPage()
        current_toc_page += 1
        # Dark background
        c.setFillColor(BG_COLOR)
        c.rect(0, 0, width, height, fill=True, stroke=False)
        y = height - 3*cm

    def draw_toc_header():
        nonlocal y
        # Dark background
        c.setFillColor(BG_COLOR)
        c.rect(0, 0, width, height, fill=True, stroke=False)

        # TOC title
        c.setFillColor(TEXT_COLOR)
        c.setFont('Helvetica-Bold', 24)
        c.drawString(2*cm, height - 2*cm, 'Table of Contents')

        # Line under title
        c.setStrokeColor(ACCENT_COLOR)
        c.setLineWidth(1)
        c.line(2*cm, height - 2.5*cm, width - 2*cm, height - 2.5*cm)
        y = height - 3.5*cm

    draw_toc_header()

    for title, page_num, depth, path in toc_entries:
        if entry_count > 0 and entry_count %% entries_per_page == 0:
            start_new_page()
            # Continue TOC header on new page
            c.setFillColor(TEXT_COLOR)
            c.setFont('Helvetica-Bold', 14)
            c.drawString(2*cm, height - 2*cm, 'Table of Contents (continued)')
            c.setStrokeColor(ACCENT_COLOR)
            c.setLineWidth(0.5)
            c.line(2*cm, height - 2.3*cm, width - 2*cm, height - 2.3*cm)
            y = height - 3*cm

        # Indentation based on depth (0.4cm per level)
        indent = depth * 0.4 * cm
        x = 2*cm + indent

        # Font size and style based on depth
        # Treat depth 0 and 1 as top-level items (bold, larger font)
        if depth <= 1:
            font_name = 'Helvetica-Bold'
            font_size = 11
            c.setFont(font_name, font_size)
            c.setFillColor(TEXT_COLOR)
        elif depth == 2:
            font_name = 'Helvetica'
            font_size = 10
            c.setFont(font_name, font_size)
            c.setFillColor(TEXT_COLOR)
        else:  # depth 3, 4, etc. - all same format
            font_name = 'Helvetica'
            font_size = 9
            c.setFont(font_name, font_size)
            c.setFillColor(TEXT_COLOR)

        # Calculate available width for title (leave space for page number and padding)
        # Page numbers are right-aligned at (width - 2*cm), leave 1.5cm space before them
        page_num_x = width - 2*cm
        max_title_x = page_num_x - 1.5*cm
        max_title_width = max_title_x - x

        # Truncate title to fit available width
        display_title = truncate_text_to_width(c, title, font_name, font_size, max_title_width)

        # Draw title
        c.drawString(x, y, display_title)

        # Draw page number (content pages start at 1)
        display_page = page_num + 1
        c.setFillColor(MUTED_COLOR)
        c.setFont('Helvetica', 9)
        c.drawRightString(width - 2*cm, y, str(display_page))

        # Store link rectangle - target is absolute page in merged PDF
        # page_num is 0-indexed content offset, add content_start_page for absolute position
        absolute_target = page_num + content_start_page
        links.append((current_toc_page, x, y - 0.1*cm, width - 2*cm - x, 0.4*cm, absolute_target))

        # Dotted line between title and page number
        c.setStrokeColor(HexColor('#30363d'))
        c.setLineWidth(0.3)
        c.setDash(1, 2)
        # Use truncated title width for dotted line calculation
        title_width = c.stringWidth(display_title, font_name, font_size)
        line_start = x + title_width + 0.3*cm
        line_end = width - 2*cm - 0.5*cm
        if line_end > line_start:
            c.line(line_start, y + 0.1*cm, line_end, y + 0.1*cm)
        c.setDash()

        y -= 0.5*cm
        entry_count += 1

    c.save()
    buffer.seek(0)
    return PdfReader(buffer), links

# Parse .nav.yml files to build proper navigation structure
def parse_nav_yml(nav_dir, base_path=''):
    """Parse .nav.yml and return ordered list of (title, path, depth) entries."""
    nav_file = os.path.join(nav_dir, '.nav.yml')
    entries = []

    if not os.path.exists(nav_file):
        return entries

    with open(nav_file, 'r') as f:
        nav_data = yaml.safe_load(f)

    if not nav_data:
        return entries

    section_title = nav_data.get('title', '')
    nav_items = nav_data.get('nav', [])

    for item in nav_items:
        if isinstance(item, str):
            # Simple file reference: "index.md" or "subdirectory/"
            if item.endswith('.md'):
                # File reference
                file_path = os.path.join(base_path, item)
                # Convert .md to site path (file.md -> file/index.pdf or index.md -> index.pdf)
                if item == 'index.md':
                    site_path = base_path if base_path else ''
                else:
                    site_path = os.path.join(base_path, item[:-3])  # Remove .md

                # Get title from markdown file, fallback to filename
                md_file_path = os.path.join(nav_dir, item)
                title = get_title_from_markdown(md_file_path)
                if not title:
                    # Fallback to filename-based title
                    title = item[:-3].replace('-', ' ').replace('_', ' ').title()

                entries.append((title, site_path, file_path))
            else:
                # Subdirectory reference
                subdir = item.rstrip('/')
                subdir_path = os.path.join(nav_dir, subdir)
                sub_base = os.path.join(base_path, subdir) if base_path else subdir
                sub_entries = parse_nav_yml(subdir_path, sub_base)
                entries.extend(sub_entries)
        elif isinstance(item, dict):
            # Titled section: {"Title": [...]} or {"Title": "path.md"} or {"Title": "subdirectory"}
            for title, content in item.items():
                if isinstance(content, str):
                    if content.endswith('.md'):
                        # Single file with custom title
                        file_path = os.path.join(base_path, content)
                        site_path = os.path.join(base_path, content[:-3])
                        entries.append((title, site_path, file_path))
                    else:
                        # Subdirectory with custom title - add header and recurse
                        subdir = content.rstrip('/')
                        subdir_path = os.path.join(nav_dir, subdir)
                        sub_base = os.path.join(base_path, subdir) if base_path else subdir

                        # Check if subdirectory exists
                        if os.path.isdir(subdir_path):
                            # Add the directory's index.md as an entry with the custom title
                            index_site_path = sub_base
                            index_file_path = os.path.join(sub_base, 'index.md')
                            entries.append((title, index_site_path, index_file_path))

                            # Recursively parse subdirectory
                            sub_entries = parse_nav_yml(subdir_path, sub_base)
                            # Skip the subdirectory's index.md since we added it with custom title
                            for sub_entry in sub_entries:
                                sub_title, sub_site_path, sub_file_path = sub_entry
                                if sub_file_path != os.path.join(sub_base, 'index.md'):
                                    entries.append(sub_entry)
                        else:
                            # Fallback: treat as file path
                            file_path = os.path.join(base_path, content)
                            site_path = os.path.join(base_path, content)
                            entries.append((title, site_path, file_path))
                elif isinstance(content, list):
                    # Inline section with items - add section header first
                    # Use special marker for section headers (no PDF, just TOC entry)
                    entries.append((title, '__section__', '__section__'))
                    for sub_item in content:
                        if isinstance(sub_item, str) and sub_item.endswith('.md'):
                            file_path = os.path.join(base_path, sub_item)
                            site_path = os.path.join(base_path, sub_item[:-3])
                            # Get title from markdown file, fallback to filename
                            md_file_path = os.path.join(nav_dir, sub_item)
                            sub_title = get_title_from_markdown(md_file_path)
                            if not sub_title:
                                sub_title = sub_item[:-3].split('/')[-1].replace('-', ' ').replace('_', ' ').title()
                            entries.append((sub_title, site_path, file_path))

    return entries

def get_pdf_path(site_dir, site_path):
    """Convert site path to PDF file path."""
    if not site_path:
        return os.path.join(site_dir, 'index.pdf')
    # Try directory/index.pdf first
    pdf_path = os.path.join(site_dir, site_path, 'index.pdf')
    if os.path.exists(pdf_path):
        return pdf_path
    # Try file.pdf
    pdf_path = os.path.join(site_dir, site_path + '.pdf')
    if os.path.exists(pdf_path):
        return pdf_path
    return None

# Parse navigation from .nav.yml files
print('Parsing navigation structure from .nav.yml files...')
nav_entries = parse_nav_yml(docs_dir)

# Build TOC entries with PDF paths and page numbers
toc_entries = []
pdf_files = []
page_counts = []
current_section_depth = 1  # Track depth for section headers

for title, site_path, file_path in nav_entries:
    if site_path == '__section__':
        # Section header (no PDF) - will point to next page
        next_page = sum(page_counts)
        toc_entries.append((title, next_page, current_section_depth, '__section__'))
        current_section_depth = 2  # Items after section header are indented
        continue

    pdf_path = get_pdf_path(site_dir, site_path)
    if pdf_path and os.path.exists(pdf_path):
        # Calculate depth from site_path
        if not site_path:
            depth = 0
        else:
            depth = len(site_path.replace('\\\\', '/').split('/'))

        # Use section depth for items in inline sections
        if current_section_depth == 2 and depth <= 2:
            depth = current_section_depth

        reader = PdfReader(pdf_path)
        page_count = len(reader.pages)
        page_counts.append(page_count)

        current_page = sum(page_counts[:-1])
        toc_entries.append((title, current_page, depth, site_path))
        pdf_files.append(pdf_path)

        # Reset section depth when we hit a new top-level item
        if depth <= 1:
            current_section_depth = 1

print(f'Found {len(pdf_files)} PDF files from navigation')

# Create cover page
print('Creating cover page...')
cover_reader = create_cover_page()
cover_pages = len(cover_reader.pages)

# Estimate TOC pages (roughly 35 entries per page)
toc_page_estimate = max(1, (len(toc_entries) + 34) // 35)

# Content starts after cover + TOC (page numbers start at 1, not 0)
content_start_page = cover_pages + toc_page_estimate

# Create TOC pages
print(f'Creating table of contents ({len(toc_entries)} entries)...')
toc_reader, toc_links = create_toc_pages(toc_entries, content_start_page)
toc_pages = len(toc_reader.pages)

# Recalculate if TOC pages changed
if toc_pages != toc_page_estimate:
    content_start_page = cover_pages + toc_pages
    toc_reader, toc_links = create_toc_pages(toc_entries, content_start_page)

# Final merge
writer = PdfWriter()

# Add cover page
for page in cover_reader.pages:
    writer.add_page(page)

# Add TOC pages
for page in toc_reader.pages:
    writer.add_page(page)

# Add content pages and track section titles for each page
current_page = cover_pages + toc_pages
bookmarks = []
page_sections = {}  # Map page index -> section title

current_section = ''
for i, pdf_path in enumerate(pdf_files):
    title, _, depth, site_path = toc_entries[i]
    bookmarks.append((title, current_page, depth, site_path))

    # Update current section for top-level entries (depth <= 1)
    if depth <= 1:
        current_section = title

    reader = PdfReader(pdf_path)
    for j, page in enumerate(reader.pages):
        writer.add_page(page)
        # Store section title for this page (use document title for first page of section)
        page_sections[current_page + j] = title if j == 0 else current_section
    current_page += len(reader.pages)

# Overlay page footers on TOC and content pages (skip cover only)
total_pages = len(writer.pages)
toc_start = cover_pages
content_start = cover_pages + toc_pages
footer_page_count = total_pages - cover_pages  # All pages after cover get footers

print(f'Adding page footers to {toc_pages} TOC + {total_pages - content_start} content pages...')

# Add footers to TOC pages (roman numerals style: i, ii, iii...)
for toc_idx in range(toc_pages):
    page_idx = toc_start + toc_idx
    page = writer.pages[page_idx]
    # Use roman numerals for TOC pages
    roman_num = ['i', 'ii', 'iii', 'iv', 'v', 'vi', 'vii', 'viii', 'ix', 'x'][toc_idx] if toc_idx < 10 else str(toc_idx + 1)
    footer_reader = create_page_footer_overlay(
        page_num=roman_num,
        total_pages=footer_page_count,
        book_title=book_title,
        section_title='Table of Contents'
    )
    page.merge_page(footer_reader.pages[0])

# Add footers to content pages (arabic numerals: 1, 2, 3...)
for page_idx in range(content_start, total_pages):
    page = writer.pages[page_idx]
    display_page_num = page_idx - content_start + 1  # Content pages start at 1
    section = page_sections.get(page_idx, '')

    footer_reader = create_page_footer_overlay(
        page_num=display_page_num,
        total_pages=total_pages - content_start,
        book_title=book_title,
        section_title=section
    )
    page.merge_page(footer_reader.pages[0])

# Add clickable links to TOC pages
print(f'Adding {len(toc_links)} TOC links...')
from pypdf.generic import ArrayObject, DictionaryObject, FloatObject, NameObject, NumberObject

for toc_page_idx, x, y, link_width, link_height, target_page in toc_links:
    # Get the actual page in the merged PDF (cover + toc_page_idx)
    page_num = cover_pages + toc_page_idx
    if page_num < len(writer.pages) and target_page < len(writer.pages):
        page = writer.pages[page_num]

        # Create link annotation
        link = DictionaryObject()
        link[NameObject('/Type')] = NameObject('/Annot')
        link[NameObject('/Subtype')] = NameObject('/Link')
        link[NameObject('/Rect')] = ArrayObject([
            FloatObject(x),
            FloatObject(y),
            FloatObject(x + link_width),
            FloatObject(y + link_height)
        ])
        link[NameObject('/Border')] = ArrayObject([NumberObject(0), NumberObject(0), NumberObject(0)])
        # Destination: go to target page, fit width
        link[NameObject('/Dest')] = ArrayObject([
            writer.pages[target_page].indirect_reference,
            NameObject('/XYZ'),
            NumberObject(0),
            NumberObject(842),  # A4 height
            NumberObject(0)
        ])

        # Add annotation to page
        if '/Annots' not in page:
            page[NameObject('/Annots')] = ArrayObject()
        page['/Annots'].append(link)

# Add hierarchical bookmarks/outline
print(f'Adding {len(bookmarks)} bookmarks...')

# Add TOC bookmark first
toc_bookmark = writer.add_outline_item('Table of Contents', cover_pages)

parent_stack = {}
for title, page_num, depth, path in bookmarks:
    if depth == 1:
        parent = writer.add_outline_item(title, page_num)
        parent_stack[1] = parent
        parent_stack[2] = None
        parent_stack[3] = None
    elif depth == 2 and parent_stack.get(1):
        parent = writer.add_outline_item(title, page_num, parent=parent_stack[1])
        parent_stack[2] = parent
        parent_stack[3] = None
    elif depth == 3 and parent_stack.get(2):
        parent = writer.add_outline_item(title, page_num, parent=parent_stack[2])
        parent_stack[3] = parent
    elif depth >= 4 and parent_stack.get(3):
        writer.add_outline_item(title, page_num, parent=parent_stack[3])

# Ensure output directory exists
os.makedirs(os.path.dirname(output_path), exist_ok=True)

# Write merged PDF
with open(output_path, 'wb') as f:
    writer.write(f)

print(f'Merged PDF written to {output_path}')
print(f'Total pages: {len(writer.pages)}')
print(f'  - Cover: {cover_pages} page(s)')
print(f'  - TOC: {toc_pages} page(s)')
print(f'  - Content: {current_page - cover_pages - toc_pages} pages')
print(f'TOC entries: {len(bookmarks)}')
`, dockerSiteDir, dockerOutputPath, dockerStagingDir, escapedTitle, escapedDescription)

	args = append(args, imageName, "python3", "-c", pythonScript)

	exitCode := RunCommandWithLog(workspaceRoot, logWriter, "docker", args...)
	if exitCode != 0 {
		return fmt.Errorf("PDF merge exited with code %d", exitCode)
	}

	return nil
}

// FormatDockerVolumePath formats a host path for Docker volume mounting.
// This needs to be imported from the parent package or defined here.
func FormatDockerVolumePath(path string) string {
	// This is a placeholder - the actual implementation should be in the parent package
	// or we need to import it
	return path
}

// RunCommandWithLog runs a command with logging.
// This needs to be imported from the parent package or defined here.
func RunCommandWithLog(workspaceRoot string, logWriter io.Writer, name string, args ...string) int {
	// This is a placeholder - the actual implementation should be in the parent package
	// or we need to import it
	return 0
}
