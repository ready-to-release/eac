# Changelog

All notable changes to **Documentation Site** will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Calendar Versioning](https://calver.org/).

## [Unreleased]

### Changed

- Simplified docs module to only build HTML site
- Moved PDF book generation to separate books module
- Updated CI pipeline to build site only (removed --all flag)
- Updated release pipeline to deploy site only (no PDF attachments)
