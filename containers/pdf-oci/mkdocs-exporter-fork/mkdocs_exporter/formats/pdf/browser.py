from __future__ import annotations

import os
import asyncio
from tempfile import NamedTemporaryFile
from typing import List, Tuple

from playwright.async_api import async_playwright, Page

from mkdocs_exporter.logging import logger


class Browser:
  """A web browser instance with optimized parallel PDF rendering."""

  # Aggressive Chromium flags for faster headless rendering
  args = [
    '--disable-background-timer-throttling',
    '--disable-renderer-backgrounding',
    '--disable-backgrounding-occluded-windows',
    '--allow-file-access-from-files',
    '--font-render-hinting=none',
    '--disable-gpu',
    '--disable-software-rasterizer',
    '--disable-extensions',
    '--disable-sync',
    '--disable-translate',
    '--no-first-run',
    '--disable-default-apps',
    '--disable-hang-monitor',
    '--disable-prompt-on-repost',
    '--disable-client-side-phishing-detection',
    '--disable-popup-blocking',
    '--disable-component-update',
    '--metrics-recording-only',
    '--safebrowsing-disable-auto-update',
  ]


  @property
  def launched(self):
    """Has the browser been launched?"""
    return self._launched


  def __init__(self, options: dict = {}):
    """The constructor."""

    self.browser = None
    self.context = None
    self._launched = False
    self.playwright = None
    self.lock = asyncio.Lock()
    self.debug = options.get('debug', False)
    self.headless = options.get('headless', True)
    self.timeout = options.get('timeout', 60_000)
    self.args = self.args + options.get('args', [])
    self.page_pool: List[Page] = []
    self.pool_size = options.get('pool_size', 6)
    self.pool_lock = asyncio.Lock()
    self.levels = {
      'warn': 'warning',
      'error': 'error',
      'info': 'info',
      'debug': 'debug'
    }


  async def launch(self) -> 'Browser':
    """Launches the browser and pre-warms page pool."""

    if self.launched:
      return self

    async with self.lock:
      if self.launched:
        return self

      logger.info('[mkdocs-exporter.pdf] Launching browser with pool_size=%d...', self.pool_size)

      self.playwright = await async_playwright().start()
      self.browser = await self.playwright.chromium.launch(headless=self.headless, args=self.args)
      self.context = await self.browser.new_context()
      self.context.on('console', self.log)

      # Pre-warm page pool
      self.page_pool = []
      for _ in range(self.pool_size):
        page = await self.context.new_page()
        self.page_pool.append(page)

      self._launched = True

    return self


  async def _acquire_page(self) -> Page:
    """Get a page from pool or create new one."""
    async with self.pool_lock:
      if self.page_pool:
        return self.page_pool.pop()
    return await self.context.new_page()


  async def _release_page(self, page: Page) -> None:
    """Return page to pool for reuse."""
    async with self.pool_lock:
      if len(self.page_pool) < self.pool_size:
        # Clear page state for reuse
        try:
          await page.goto('about:blank', wait_until='commit', timeout=5000)
          self.page_pool.append(page)
          return
        except Exception:
          pass
    # Pool full or page unusable - close it
    try:
      await page.close()
    except Exception:
      pass


  async def close(self) -> 'Browser':
    """Closes the browser and page pool."""

    # Close pooled pages
    for page in self.page_pool:
      try:
        await page.close()
      except Exception:
        pass
    self.page_pool = []

    if self.context:
      await self.context.close()
    if self.browser:
      await self.browser.close()
    if self.playwright:
      await self.playwright.stop()

    self._launched = False

    return self


  async def print(self, html: str) -> Tuple[bytes, int]:
    """Prints HTML to PDF using temp file in ramdisk."""

    page = await self._acquire_page()

    # Use ramdisk for temp file (TMPDIR=/dev/shm set in Dockerfile)
    # This gives file:// URL access while minimizing I/O overhead
    file = NamedTemporaryFile(suffix='.html', mode='w+', encoding='utf-8', delete=False)

    try:
      file.write(html)
      file.close()

      # Use 'load' instead of 'networkidle' - faster for local files
      await page.goto('file://' + file.name, wait_until='load', timeout=self.timeout)

      # Wait for mkdocs-exporter ready signal
      await page.locator('body[mkdocs-exporter="true"]').wait_for(timeout=self.timeout)

      pages = int(await page.locator('body').get_attribute('mkdocs-exporter-pages', timeout=self.timeout) or 0)
      pdf = await page.pdf(prefer_css_page_size=True, print_background=True, display_header_footer=False)

      return (pdf, pages)

    finally:
      await self._release_page(page)
      try:
        os.unlink(file.name)
      except OSError:
        pass


  async def print_batch(self, html_list: List[str]) -> List[Tuple[bytes, int]]:
    """Render multiple pages in parallel using page pool."""
    tasks = [self.print(html) for html in html_list]
    return await asyncio.gather(*tasks)


  async def log(self, msg):
    """Logs a message coming from the browser."""

    prefix = '[mkdocs-exporter]'
    text = msg.text
    level = self.levels.get(msg.type, 'info')

    if text.startswith(prefix):
      text = msg.text[len(prefix):].strip()

    if self.debug or level == 'error':
      getattr(logger, level)(f"[mkdocs-exporter.pdf.browser] ({msg.type}) {await msg.page.title()}\n{text}")
