from __future__ import annotations

import io
import os
import re
from concurrent.futures import ProcessPoolExecutor, ThreadPoolExecutor
from datetime import datetime
from io import BytesIO
from typing import List, Tuple, Union

from pypdf import PdfWriter, PdfReader
from pypdf.generic import ArrayObject, DictionaryObject, FloatObject, NameObject, NumberObject
from reportlab.lib.colors import HexColor
from reportlab.lib.pagesizes import A4
from reportlab.lib.units import cm
from reportlab.pdfgen import canvas

from mkdocs_exporter.page import Page
from mkdocs_exporter.formats.pdf.renderer import Renderer
from mkdocs_exporter.formats.pdf.preprocessor import Preprocessor
from mkdocs_exporter.logging import logger


# Dark theme colors (matching PDF dark theme)
BG_COLOR = HexColor('#0d1117')
TEXT_COLOR = HexColor('#e6edf3')
ACCENT_COLOR = HexColor('#58a6ff')
MUTED_COLOR = HexColor('#8b949e')
FOOTER_COLOR = HexColor('#6e7681')


def _merge_pdfs_to_bytes(inputs: List[Union[str, bytes]]) -> bytes:
  """Merge PDFs (file paths or bytes) into bytes."""
  writer = PdfWriter()
  for item in inputs:
    if isinstance(item, bytes):
      writer.append(BytesIO(item))
    else:
      writer.append(item)
  output = BytesIO()
  writer.write(output)
  writer.close()
  return output.getvalue()


def _merge_batch_from_files(args: Tuple[List[str], int]) -> Tuple[int, bytes]:
  """Merge batch of file paths. Pickle-safe for ProcessPoolExecutor."""
  batch, batch_idx = args
  return (batch_idx, _merge_pdfs_to_bytes(batch))


def _merge_batch_from_bytes(args: Tuple[List[bytes], int]) -> Tuple[int, bytes]:
  """Merge batch of bytes. Pickle-safe for ProcessPoolExecutor."""
  batch, batch_idx = args
  return (batch_idx, _merge_pdfs_to_bytes(batch))


def _truncate_text(c, text: str, font: str, size: int, max_width: float) -> str:
  """Truncate text to fit within max_width."""
  if c.stringWidth(text, font, size) <= max_width:
    return text

  ellipsis = '...'
  ellipsis_width = c.stringWidth(ellipsis, font, size)
  available = max_width - ellipsis_width

  # Try word boundary first
  words = text.split()
  if len(words) > 1:
    truncated = []
    for word in words:
      test = ' '.join(truncated + [word])
      if c.stringWidth(test, font, size) <= available:
        truncated.append(word)
      else:
        break
    if truncated:
      return ' '.join(truncated) + ellipsis

  # Character truncation fallback
  for i in range(len(text), 0, -1):
    if c.stringWidth(text[:i], font, size) <= available:
      return text[:i] + ellipsis
  return ellipsis


def _draw_wrapped_text(c, text: str, x: float, y: float, max_width: float,
                       font: str = 'Helvetica', size: int = 12,
                       line_height: float = 0.5*cm) -> int:
  """Draw text with word wrapping. Returns number of lines."""
  c.setFont(font, size)
  words = text.split()
  lines = []
  current = []

  for word in words:
    test = ' '.join(current + [word])
    if c.stringWidth(test, font, size) <= max_width:
      current.append(word)
    else:
      if current:
        lines.append(' '.join(current))
      current = [word]
  if current:
    lines.append(' '.join(current))

  for i, line in enumerate(lines):
    c.drawCentredString(x, y - (i * line_height), line)
  return len(lines)


def _create_cover_page(book_title: str, book_description: str) -> PdfReader:
  """Create a cover page PDF."""
  buf = io.BytesIO()
  c = canvas.Canvas(buf, pagesize=A4)
  width, height = A4

  # Dark background
  c.setFillColor(BG_COLOR)
  c.rect(0, 0, width, height, fill=True, stroke=False)

  # Brand title
  c.setFillColor(TEXT_COLOR)
  c.setFont('Helvetica-Bold', 32)
  c.drawCentredString(width/2, height - 6*cm, 'Ready-to-Release')

  c.setFont('Helvetica', 20)
  c.drawCentredString(width/2, height - 7.5*cm, 'Documentation')

  # Book title (accent color)
  if book_title:
    c.setFillColor(ACCENT_COLOR)
    max_width = width - 4*cm

    for size in [28, 24, 20, 18]:
      c.setFont('Helvetica-Bold', size)
      if c.stringWidth(book_title, 'Helvetica-Bold', size) <= max_width:
        c.drawCentredString(width/2, height - 10*cm, book_title)
        break
    else:
      _draw_wrapped_text(c, book_title, width/2, height - 10*cm, max_width, 'Helvetica-Bold', 18, 0.6*cm)
    c.setFillColor(TEXT_COLOR)

  # Accent line
  c.setStrokeColor(ACCENT_COLOR)
  c.setLineWidth(2)
  c.line(4*cm, height - 11.5*cm, width - 4*cm, height - 11.5*cm)

  # Subtitle
  c.setFillColor(MUTED_COLOR)
  c.setFont('Helvetica', 14)
  c.drawCentredString(width/2, height - 13*cm, 'Everything-as-Code Platform')
  c.drawCentredString(width/2, height - 14*cm, 'for Software Delivery Flows')

  # Description
  if book_description:
    c.setFont('Helvetica', 12)
    c.setFillColor(TEXT_COLOR)
    _draw_wrapped_text(c, book_description, width/2, height - 16*cm, width - 8*cm, 'Helvetica', 12, 0.5*cm)

  # Date
  c.setFillColor(MUTED_COLOR)
  c.setFont('Helvetica', 11)
  c.drawCentredString(width/2, 3*cm, datetime.now().strftime('Generated: %B %d, %Y'))

  c.save()
  buf.seek(0)
  return PdfReader(buf)


def _create_toc_pages(toc_entries: list, content_start: int, book_title: str) -> Tuple[PdfReader, list]:
  """Create TOC pages. Returns (reader, links)."""
  buf = io.BytesIO()
  c = canvas.Canvas(buf, pagesize=A4)
  width, height = A4

  y = height - 3.5*cm
  entries_per_page = 35
  count = 0
  links = []
  current_page = 0

  def draw_header(is_first: bool = True):
    nonlocal y
    c.setFillColor(BG_COLOR)
    c.rect(0, 0, width, height, fill=True, stroke=False)
    c.setFillColor(TEXT_COLOR)

    if is_first:
      c.setFont('Helvetica-Bold', 24)
      c.drawString(2*cm, height - 2*cm, 'Table of Contents')
    else:
      c.setFont('Helvetica-Bold', 14)
      c.drawString(2*cm, height - 2*cm, 'Table of Contents (continued)')

    c.setStrokeColor(ACCENT_COLOR)
    c.setLineWidth(1 if is_first else 0.5)
    c.line(2*cm, height - 2.5*cm, width - 2*cm, height - 2.5*cm)
    y = height - 3.5*cm

  draw_header(True)

  for title, page_num, depth in toc_entries:
    if count > 0 and count % entries_per_page == 0:
      c.showPage()
      current_page += 1
      draw_header(False)

    indent = depth * 0.4 * cm
    x = 2*cm + indent

    if depth <= 1:
      font, size = 'Helvetica-Bold', 11
    elif depth == 2:
      font, size = 'Helvetica', 10
    else:
      font, size = 'Helvetica', 9

    c.setFont(font, size)
    c.setFillColor(TEXT_COLOR)

    max_title_width = width - 4*cm - indent - 1.5*cm
    display_title = _truncate_text(c, title, font, size, max_title_width)
    c.drawString(x, y, display_title)

    # Page number
    display_page = page_num + 1
    c.setFillColor(MUTED_COLOR)
    c.setFont('Helvetica', 9)
    c.drawRightString(width - 2*cm, y, str(display_page))

    # Link
    target = page_num + content_start
    links.append((current_page, x, y - 0.1*cm, width - 2*cm - x, 0.4*cm, target))

    # Dotted line
    c.setStrokeColor(HexColor('#30363d'))
    c.setLineWidth(0.3)
    c.setDash(1, 2)
    title_width = c.stringWidth(display_title, font, size)
    line_start = x + title_width + 0.3*cm
    line_end = width - 2.5*cm
    if line_end > line_start:
      c.line(line_start, y + 0.1*cm, line_end, y + 0.1*cm)
    c.setDash()

    y -= 0.5*cm
    count += 1

  c.save()
  buf.seek(0)
  return PdfReader(buf), links


def _create_footer_overlay(page_num, book_title: str, section: str = '') -> PdfReader:
  """Create footer overlay for a page."""
  buf = io.BytesIO()
  c = canvas.Canvas(buf, pagesize=A4)
  width, height = A4
  footer_y = 1.5 * cm

  # Page number center
  c.setFont('Helvetica', 9)
  c.setFillColor(MUTED_COLOR)
  c.drawCentredString(width / 2, footer_y, str(page_num))

  # Book title left
  c.setFont('Helvetica', 8)
  c.setFillColor(FOOTER_COLOR)
  max_width = (width / 2) - 3*cm
  display_title = _truncate_text(c, book_title, 'Helvetica', 8, max_width)
  c.drawString(2*cm, footer_y, display_title)

  # Section right
  if section:
    max_section = (width / 2) - 1*cm
    display_section = _truncate_text(c, section, 'Helvetica', 8, max_section)
    c.drawRightString(width - 2*cm, footer_y, display_section)

  c.save()
  buf.seek(0)
  return PdfReader(buf)


class Aggregator:
  """Aggregates PDF documents together with parallel tree-merge, cover, TOC, and page numbers."""


  def __init__(self, renderer: Renderer, config: dict = {}) -> None:
    """The constructor."""

    self.pages = []
    self.config = config
    self.renderer = renderer
    self.pending_files: List[str] = []
    self.page_titles: List[str] = []  # Track titles for TOC


  def open(self, path: str) -> Aggregator:
    """Opens the aggregator."""

    self.path = path
    self.pending_files = []
    self.page_titles = []

    return self


  def set_pages(self, pages: list[Page]) -> Aggregator:
    """Sets the pages."""

    self.pages = pages
    covers = self.config.get('covers', [])

    for index, page in enumerate(self.pages):
      if covers == 'none':
        self._skip(page, ['front', 'back'])
      elif covers == 'front':
        self._skip(page, ['back'])
      elif covers == 'back':
        self._skip(page, ['front'])
      elif covers == 'limits':
        if len(self.pages) == 1:
          pass
        elif index == 0:
          self._skip(page, ['back'])
        elif index == (len(self.pages) - 1):
          self._skip(page, ['front'])
        else:
          self._skip(page, ['front', 'back'])
      elif covers == 'book':
        if len(self.pages) == 1:
          pass
        elif index == 0:
          self._skip(page, ['back'])
        elif index < (len(self.pages) - 1):
          self._skip(page, ['back'])

    return self


  def preprocess(self, page: Page) -> str:
    """Preprocesses the page."""

    preprocessor = Preprocessor()

    preprocessor.preprocess(self.renderer.preprocess(page, disable=['teleport']))

    if 'front' not in page.formats['pdf']['covers']:
      preprocessor.remove('div.mkdocs-exporter-front-cover')
    if 'back' not in page.formats['pdf']['covers']:
      preprocessor.remove('div.mkdocs-exporter-back-cover')

    preprocessor.teleport()
    preprocessor.metadata({
      'page': sum(page.formats['pdf']['pages'] - page.formats['pdf']['skipped_pages'] for page in self.pages[:page.index]) + 1,
      'pages': sum(page.formats['pdf']['pages'] - page.formats['pdf']['skipped_pages'] for page in self.pages),
    })

    return preprocessor.done()


  def append(self, document: str, title: str = '') -> Aggregator:
    """Queues a document for parallel merge."""

    self.pending_files.append(document)
    self.page_titles.append(title)

    return self


  def save(self, metadata={}) -> Aggregator:
    """Saves the aggregated document with cover, TOC, and page numbers."""

    if not self.pending_files:
      return self

    os.makedirs(os.path.dirname(self.path), exist_ok=True)

    batch_size = self.config.get('merge_batch_size', 50)
    max_workers = self.config.get('merge_workers', 6)
    total_files = len(self.pending_files)

    # Get book info from config
    book_title = self.config.get('book_title', 'Documentation')
    book_description = self.config.get('book_description', '')

    logger.info("[mkdocs-exporter.pdf] Parallel merge: %d files, batch=%d, workers=%d",
                total_files, batch_size, max_workers)

    # Parallel tree-merge
    current_items = self.pending_files
    merge_round = 0
    is_first_round = True

    while len(current_items) > 1:
      merge_round += 1
      batches = [current_items[i:i + batch_size] for i in range(0, len(current_items), batch_size)]
      num_batches = len(batches)

      if num_batches == 1:
        final_bytes = _merge_pdfs_to_bytes(batches[0])
        current_items = [final_bytes]
        break

      results = [None] * num_batches

      if is_first_round:
        with ProcessPoolExecutor(max_workers=min(max_workers, num_batches)) as executor:
          args = [(batch, idx) for idx, batch in enumerate(batches)]
          for idx, merged_bytes in executor.map(_merge_batch_from_files, args):
            results[idx] = merged_bytes
        is_first_round = False
      else:
        with ThreadPoolExecutor(max_workers=min(max_workers, num_batches)) as executor:
          args = [(batch, idx) for idx, batch in enumerate(batches)]
          futures = {executor.submit(_merge_batch_from_bytes, arg): arg[1] for arg in args}
          for future in futures:
            idx, merged_bytes = future.result()
            results[idx] = merged_bytes

      current_items = results
      logger.info("[mkdocs-exporter.pdf] Merge round %d: %d -> %d batches",
                  merge_round, total_files if merge_round == 1 else num_batches, len(current_items))

    # Read merged content
    final_data = current_items[0] if isinstance(current_items[0], bytes) else open(current_items[0], 'rb').read()
    content_reader = PdfReader(BytesIO(final_data))

    # Build TOC entries from page titles and counts
    toc_entries = []
    page_counts = []

    for i, pdf_path in enumerate(self.pending_files):
      try:
        reader = PdfReader(pdf_path)
        count = len(reader.pages)
        page_counts.append(count)

        title = self.page_titles[i] if i < len(self.page_titles) else f'Page {i+1}'
        # Calculate depth from path
        depth = pdf_path.count(os.sep) - self.pending_files[0].count(os.sep) + 1
        depth = max(1, min(depth, 3))

        current_page = sum(page_counts[:-1])
        toc_entries.append((title, current_page, depth))
      except Exception:
        page_counts.append(1)

    # Create cover
    logger.info("[mkdocs-exporter.pdf] Creating cover page...")
    cover_reader = _create_cover_page(book_title, book_description)
    cover_pages = len(cover_reader.pages)

    # Estimate TOC pages
    toc_page_estimate = max(1, (len(toc_entries) + 34) // 35)
    content_start = cover_pages + toc_page_estimate

    # Create TOC
    logger.info("[mkdocs-exporter.pdf] Creating TOC (%d entries)...", len(toc_entries))
    toc_reader, toc_links = _create_toc_pages(toc_entries, content_start, book_title)
    toc_pages = len(toc_reader.pages)

    # Recalculate if needed
    if toc_pages != toc_page_estimate:
      content_start = cover_pages + toc_pages
      toc_reader, toc_links = _create_toc_pages(toc_entries, content_start, book_title)

    # Final assembly
    writer = PdfWriter()

    # Add cover
    for page in cover_reader.pages:
      writer.add_page(page)

    # Add TOC
    for page in toc_reader.pages:
      writer.add_page(page)

    # Add content
    for page in content_reader.pages:
      writer.add_page(page)

    total_pages = len(writer.pages)

    # Add footers
    logger.info("[mkdocs-exporter.pdf] Adding page footers...")

    # TOC footers (roman numerals)
    roman = ['i', 'ii', 'iii', 'iv', 'v', 'vi', 'vii', 'viii', 'ix', 'x']
    for i in range(toc_pages):
      page_idx = cover_pages + i
      num = roman[i] if i < 10 else str(i + 1)
      footer = _create_footer_overlay(num, book_title, 'Table of Contents')
      writer.pages[page_idx].merge_page(footer.pages[0])

    # Content footers (arabic numerals)
    content_start_idx = cover_pages + toc_pages
    for i in range(content_start_idx, total_pages):
      page_num = i - content_start_idx + 1
      # Find section for this page
      section = ''
      cumulative = 0
      for j, count in enumerate(page_counts):
        if cumulative + count > i - content_start_idx:
          section = toc_entries[j][0] if j < len(toc_entries) else ''
          break
        cumulative += count

      footer = _create_footer_overlay(page_num, book_title, section)
      writer.pages[i].merge_page(footer.pages[0])

    # Add TOC links
    logger.info("[mkdocs-exporter.pdf] Adding %d TOC links...", len(toc_links))
    for toc_page_idx, x, y, link_width, link_height, target_page in toc_links:
      page_num = cover_pages + toc_page_idx
      if page_num < len(writer.pages) and target_page < len(writer.pages):
        page = writer.pages[page_num]

        link = DictionaryObject()
        link[NameObject('/Type')] = NameObject('/Annot')
        link[NameObject('/Subtype')] = NameObject('/Link')
        link[NameObject('/Rect')] = ArrayObject([
          FloatObject(x), FloatObject(y),
          FloatObject(x + link_width), FloatObject(y + link_height)
        ])
        link[NameObject('/Border')] = ArrayObject([NumberObject(0), NumberObject(0), NumberObject(0)])
        link[NameObject('/Dest')] = ArrayObject([
          writer.pages[target_page].indirect_reference,
          NameObject('/XYZ'), NumberObject(0), NumberObject(842), NumberObject(0)
        ])

        if '/Annots' not in page:
          page[NameObject('/Annots')] = ArrayObject()
        page['/Annots'].append(link)

    # Add bookmarks
    logger.info("[mkdocs-exporter.pdf] Adding bookmarks...")
    writer.add_outline_item('Table of Contents', cover_pages)

    parent_stack = {}
    for title, page_num, depth in toc_entries:
      target = page_num + content_start
      if depth == 1:
        parent = writer.add_outline_item(title, target)
        parent_stack[1] = parent
        parent_stack[2] = None
      elif depth == 2 and parent_stack.get(1):
        parent = writer.add_outline_item(title, target, parent=parent_stack[1])
        parent_stack[2] = parent
      elif depth >= 3 and parent_stack.get(2):
        writer.add_outline_item(title, target, parent=parent_stack[2])

    # Write output
    writer.add_metadata({'/Producer': 'MkDocs Exporter (with TOC)', **metadata})

    with open(self.path, 'wb') as f:
      writer.write(f)
    writer.close()

    logger.info("[mkdocs-exporter.pdf] Complete: %d pages (cover: %d, TOC: %d, content: %d) -> %s",
                total_pages, cover_pages, toc_pages, total_pages - cover_pages - toc_pages, self.path)

    return self


  def _skip(self, page: Page, covers: list[str]) -> Aggregator:
    """Skip cover pages."""

    for cover in covers:
      if cover in page.formats['pdf']['covers']:
        page.formats['pdf']['covers'].remove(cover)

        page.formats['pdf']['skipped_pages'] = page.formats['pdf']['skipped_pages'] + 1

    return self
