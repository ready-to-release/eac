"""
MkDocs Macros for Documentation Footer Generation

This module provides dynamic macros for generating Diátaxis-based navigation footers
throughout the documentation. The footer adapts based on the current page's location
and section (Tutorials, How-to Guides, Explanation, Reference).

Usage in documentation files:
    {{ diataxis_footer() }}
"""

import os
from pathlib import Path


def define_env(env):
    """
    Define custom macros for MkDocs

    This function is called by mkdocs-macros-plugin to register custom macros.
    """

    # Diátaxis section descriptions
    SECTION_DESCRIPTIONS = {
        'tutorials': 'learning-oriented guides that take you through steps to complete a project',
        'how-to-guides': 'task-oriented recipes that guide you through solving specific problems',
        'explanation': 'understanding-oriented discussion that clarifies concepts',
        'reference': 'information-oriented technical descriptions of the system'
    }

    # Diátaxis section display names
    SECTION_NAMES = {
        'tutorials': 'Tutorials',
        'how-to-guides': 'How-to Guides',
        'explanation': 'Explanation',
        'reference': 'Reference'
    }

    def detect_section(file_path):
        """
        Detect which Diátaxis section a file belongs to based on its path.

        Args:
            file_path (str): Relative path from docs/ root (e.g., "explanation/continuous-delivery/index.md")

        Returns:
            str: Section name ('tutorials', 'how-to-guides', 'explanation', 'reference')
                 or None if not in a recognized section
        """
        # Normalize path separators
        path_parts = file_path.replace('\\', '/').split('/')

        # First part of path is the section
        if len(path_parts) > 0:
            first_part = path_parts[0].lower()
            if first_part in SECTION_DESCRIPTIONS:
                return first_part

        return None

    def calculate_depth(file_path):
        """
        Calculate the depth of a file from the docs/ root.

        Args:
            file_path (str): Relative path from docs/ root

        Returns:
            int: Depth level - accounts for MkDocs converting files to directories
                 - index.md files: depth based on directory nesting
                 - other .md files: depth includes the file itself as a directory
        """
        # Normalize path separators
        normalized_path = file_path.replace('\\', '/')

        # Split path into parts
        parts = normalized_path.split('/')

        # MkDocs converts markdown files to directories with index.html
        # - docs/tutorials/index.md → /tutorials/ (depth 1)
        # - docs/tutorials/quick-start.md → /tutorials/quick-start/ (depth 2)

        filename = parts[-1] if parts else ''

        if filename == 'index.md':
            # For index files, depth is based on directory nesting only
            depth = len(parts) - 1
        else:
            # For non-index files, the file itself becomes a directory
            depth = len(parts)

        return max(1, depth)

    def generate_nav_links(depth, current_section):
        """
        Generate navigation links for all Diátaxis sections.

        Args:
            depth (int): Depth level of current file
            current_section (str): Current section name

        Returns:
            str: Markdown-formatted navigation links
        """
        # Calculate relative path prefix based on depth
        prefix = '../' * depth

        sections = ['tutorials', 'how-to-guides', 'explanation', 'reference']
        links = []

        for section in sections:
            display_name = SECTION_NAMES[section]
            link_path = f'{prefix}{section}/'

            if section == current_section:
                # Current section - bold link to section index
                links.append(f'**[{display_name}]({link_path})**')
            else:
                # Other sections - regular link
                links.append(f'[{display_name}]({link_path})')

        return ' | '.join(links)

    @env.macro
    def diataxis_footer():
        """
        Generate a Diátaxis navigation footer for the current page.

        This macro automatically:
        - Detects which section the current page belongs to
        - Calculates the correct relative paths based on file depth
        - Highlights the current section in bold
        - Includes the appropriate "You are here" description

        Returns:
            str: Complete footer HTML/Markdown
        """
        # Get current page from MkDocs context
        page = env.page

        if not page or not page.file:
            return '<!-- Footer: Unable to determine page context -->'

        # Get file path relative to docs/ root
        file_path = page.file.src_uri  # e.g., "explanation/continuous-delivery/index.md"

        # Detect section
        section = detect_section(file_path)

        if not section:
            return '<!-- Footer: Page not in a recognized Diátaxis section -->'

        # Calculate depth
        depth = calculate_depth(file_path)

        # Generate navigation links
        nav_links = generate_nav_links(depth, section)

        # Get section description
        description = SECTION_DESCRIPTIONS[section]
        section_name = SECTION_NAMES[section]

        # Generate footer with smaller font for visual distinction
        # Convert markdown-style links to HTML for the footer
        sections_list = ['tutorials', 'how-to-guides', 'explanation', 'reference']
        prefix = '../' * depth

        html_links = []
        for sec in sections_list:
            display_name = SECTION_NAMES[sec]
            link_path = f'{prefix}{sec}/'

            if sec == section:
                # Current section - bold link
                html_links.append(f'<strong><a href="{link_path}">{display_name}</a></strong>')
            else:
                # Other sections - regular link
                html_links.append(f'<a href="{link_path}">{display_name}</a>')

        nav_html = ' | '.join(html_links)

        # Footer is styled via CSS using 'article hr:last-of-type ~ p' selector
        # The --- creates an <hr> and paragraphs after it are automatically styled
        footer = f"""---

<p><em>{nav_html}</em></p>

<p><strong>You are here:</strong> {section_name} — {description}.</p>"""

        return footer

    @env.macro
    def diataxis_footer_debug():
        """
        Debug version of diataxis_footer that shows calculation details.
        Useful for troubleshooting path issues.

        Returns:
            str: Footer with debug information
        """
        page = env.page

        if not page or not page.file:
            return '<!-- Debug: No page context -->'

        file_path = page.file.src_uri
        section = detect_section(file_path)
        depth = calculate_depth(file_path)

        debug_info = f"""---

**Debug Information:**
- File path: `{file_path}`
- Detected section: `{section}`
- Calculated depth: `{depth}`
- Path prefix: `{'../' * depth}`

{diataxis_footer()}"""

        return debug_info
