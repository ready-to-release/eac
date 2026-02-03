from __future__ import annotations

import os

from urllib.parse import unquote

from mkdocs_exporter.page import Page
from mkdocs_exporter.formats.pdf.browser import Browser
from mkdocs_exporter.renderer import Renderer as BaseRenderer
from mkdocs_exporter.formats.pdf.preprocessor import Preprocessor


# Native Chrome PDF script - signals ready after images load
# Chrome handles @page rules natively; ReportLab adds page numbers in aggregator
READY_SCRIPT = '''
(function() {
  function signalReady() {
    document.body.setAttribute('mkdocs-exporter-pages', '1');
    document.body.setAttribute('mkdocs-exporter', 'true');
  }

  var images = document.images;
  var loaded = 0;
  var total = images.length;

  if (total === 0) {
    signalReady();
    return;
  }

  function checkComplete() {
    loaded++;
    if (loaded >= total) signalReady();
  }

  for (var i = 0; i < total; i++) {
    if (images[i].complete) {
      checkComplete();
    } else {
      images[i].addEventListener('load', checkComplete);
      images[i].addEventListener('error', checkComplete);
    }
  }
})();
'''


class Renderer(BaseRenderer):
  """The renderer."""


  def __init__(self, browser: Browser = None, options: dict = None):
    """The constructor."""

    self.options: dict = options
    self.scripts: list[str] = []
    self.stylesheets: list[str] = []
    self.browser = browser or Browser(self.options.get('browser', {}))


  def add_stylesheet(self, path: str) -> Renderer:
    """Adds a stylesheet to the renderer."""

    self.stylesheets.append(path)

    return self


  def add_script(self, path: str) -> Renderer:
    """Adds a script to the renderer."""

    self.scripts.append(path)

    return self


  def cover(self, template: str, location: str) -> Renderer:
    """Renders a cover."""

    content = template.strip('\n')

    return f'<div class="mkdocs-exporter-{location}-cover" data-decompose="true">{content}</div>' + '\n'


  def preprocess(self, page: Page, disable: list = []) -> str:
    """Preprocesses a page, returning HTML that can be printed."""

    preprocessor = Preprocessor(theme=page.theme)
    base = os.path.dirname(page.file.abs_dest_path)
    root = base.replace(unquote(page.url).rstrip('/'), '', 1).rstrip('/')

    preprocessor.preprocess(page.html)
    preprocessor.set_attribute('details:not([open])', 'open', 'open')
    page.theme.preprocess(preprocessor)

    for stylesheet in self.stylesheets:
      with open(stylesheet, 'r', encoding='utf-8') as file:
        preprocessor.stylesheet(file.read())
    for script in self.scripts:
      with open(script, 'r', encoding='utf-8') as file:
        preprocessor.script(file.read(), path=stylesheet)

    if 'teleport' not in disable:
      preprocessor.teleport()

    preprocessor.script(READY_SCRIPT)
    preprocessor.update_links(base, root)

    if self.options.get('url'):
      preprocessor.rewrite_links(page.abs_url, self.options['url'])

    return preprocessor.done()


  async def render(self, page: str | Page) -> tuple[bytes, int]:
    """Renders a page as a PDF document."""

    if not self.browser.launched:
      await self.browser.launch()

    html = page if isinstance(page, str) else self.preprocess(page)

    return await self.browser.print(html)


  async def dispose(self) -> None:
    """Dispose of the renderer."""

    if self.browser:
      await self.browser.close()
