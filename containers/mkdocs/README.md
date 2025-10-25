# MkDocs Container

Docker container for building and serving MkDocs documentation with Material theme.

## Contents

- `.Dockerfile` - Container definition
- `requirements.txt` - Python dependencies
- `docker-compose.yml` - Docker Compose configuration
- `.dockerignore` - Files to exclude from build context

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+ (included with Docker Desktop)

## Quick Start

### Serve Documentation (Development)

```bash
# From containers/mkdocs/ directory
docker-compose up

# Or from project root
docker-compose -f containers/mkdocs/docker-compose.yml up
```

Open http://localhost:8000 in your browser.

### Build Documentation (Production)

```bash
# Build static site
docker-compose run --rm mkdocs-build

# Or with profile
docker-compose --profile build up mkdocs-build
```

Output will be in `site/` directory.

## Usage

### Option 1: Docker Compose (Recommended)

**Serve with live reload:**
```bash
cd containers/mkdocs
docker-compose up
```

**Build static site:**
```bash
cd containers/mkdocs
docker-compose run --rm mkdocs-build
```

**Run custom command:**
```bash
docker-compose run --rm mkdocs mkdocs --help
```

### Option 2: Docker CLI

**Build image:**
```bash
cd containers/mkdocs
docker build -f .Dockerfile -t cli-mkdocs:latest .
```

**Serve documentation:**
```bash
docker run --rm -it \
  -p 8000:8000 \
  -v $(pwd)/../..:/docs \
  cli-mkdocs:latest \
  mkdocs serve --dev-addr=0.0.0.0:8000
```

**Build static site:**
```bash
docker run --rm \
  -v $(pwd)/../..:/docs \
  cli-mkdocs:latest \
  mkdocs build --clean --strict
```

### Option 3: Helper Scripts

Use the provided helper scripts (see below).

## Helper Scripts

### serve.sh - Development Server

```bash
# From containers/mkdocs/
./serve.sh

# From project root
./containers/mkdocs/serve.sh
```

Starts MkDocs development server with live reload.

### build.sh - Build Static Site

```bash
# From containers/mkdocs/
./build.sh

# From project root
./containers/mkdocs/build.sh
```

Builds static site to `site/` directory.

### shell.sh - Interactive Shell

```bash
# From containers/mkdocs/
./shell.sh

# Inside container:
mkdocs --version
mkdocs build
mkdocs serve
```

Opens interactive shell in the container.

## Container Details

### Base Image
- `python:3.11-slim` - Lightweight Python 3.11

### Installed Packages
- `mkdocs` 1.5.3
- `mkdocs-material` 9.5.3 (Material theme)
- `mkdocs-git-revision-date-localized-plugin` (Git dates)
- `mkdocs-minify-plugin` (Minification)
- `mkdocs-redirects` (URL redirects)
- `pymdown-extensions` (Enhanced markdown)

### Ports
- `8000` - MkDocs development server

### Volumes
- `/docs` - Mounted project root

### User
- Non-root user `mkdocs` (UID 1000)

## Configuration

### Update Dependencies

Edit `requirements.txt`:

```txt
mkdocs==1.5.3
mkdocs-material==9.5.3
# Add more packages...
```

Rebuild:
```bash
docker-compose build --no-cache
```

### Change Port

Edit `docker-compose.yml`:

```yaml
ports:
  - "8080:8000"  # Changed from 8000:8000
```

### Add Environment Variables

Edit `docker-compose.yml`:

```yaml
environment:
  - PYTHONUNBUFFERED=1
  - MY_VAR=value
```

## Development Workflow

### 1. Start Development Server

```bash
cd containers/mkdocs
docker-compose up
```

Server starts at http://localhost:8000 with live reload.

### 2. Edit Documentation

Edit files in `docs/` directory. Changes appear automatically.

### 3. Build for Production

```bash
docker-compose run --rm mkdocs-build
```

Static site created in `site/` directory.

### 4. Deploy

Upload `site/` directory to your web server or hosting platform.

## Troubleshooting

### Container Won't Start

**Check Docker is running:**
```bash
docker ps
```

**Check logs:**
```bash
docker-compose logs mkdocs
```

**Rebuild container:**
```bash
docker-compose build --no-cache
docker-compose up
```

### Port Already in Use

**Check what's using port 8000:**
```bash
# Linux/macOS
lsof -i :8000

# Windows
netstat -ano | findstr :8000
```

**Use different port:**
Edit `docker-compose.yml` and change `"8000:8000"` to `"8080:8000"`.

### Permission Errors

**Check file permissions:**
```bash
ls -la docs/
```

**Fix permissions:**
```bash
# Linux/macOS
chmod -R 755 docs/

# Or run container as your user
docker-compose run --user $(id -u):$(id -g) mkdocs mkdocs serve
```

### Build Fails

**Check mkdocs.yml syntax:**
```bash
docker-compose run --rm mkdocs mkdocs build --strict
```

**Check for broken links:**
```bash
docker-compose run --rm mkdocs mkdocs build --strict --verbose
```

### Changes Not Appearing

**Hard refresh browser:**
- Windows/Linux: `Ctrl + F5`
- macOS: `Cmd + Shift + R`

**Clear cache:**
```bash
docker-compose down -v
docker-compose up
```

**Restart container:**
```bash
docker-compose restart mkdocs
```

## Advanced Usage

### Custom Build Commands

```bash
# Build with verbose output
docker-compose run --rm mkdocs mkdocs build --verbose

# Build to custom directory
docker-compose run --rm mkdocs mkdocs build --site-dir custom-site

# Clean build
docker-compose run --rm mkdocs mkdocs build --clean --strict
```

### Deploy to GitHub Pages

```bash
# Build and deploy
docker-compose run --rm mkdocs mkdocs gh-deploy
```

### Run Tests

```bash
# Validate configuration
docker-compose run --rm mkdocs mkdocs build --strict

# Check for warnings
docker-compose run --rm mkdocs mkdocs build --strict --verbose 2>&1 | grep -i warning
```

### Multi-Stage Build (Optimized)

For production, create an optimized multi-stage build:

```dockerfile
# Build stage
FROM python:3.11-slim as builder
WORKDIR /build
COPY requirements.txt .
RUN pip install --user -r requirements.txt
COPY . .
RUN mkdocs build

# Runtime stage
FROM nginx:alpine
COPY --from=builder /build/site /usr/share/nginx/html
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Build Docs

on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build documentation
        run: |
          cd containers/mkdocs
          docker-compose run --rm mkdocs-build

      - name: Deploy
        # Upload site/ directory
```

### GitLab CI

```yaml
build-docs:
  image: docker:latest
  services:
    - docker:dind
  script:
    - cd containers/mkdocs
    - docker-compose run --rm mkdocs-build
  artifacts:
    paths:
      - site/
```

## Maintenance

### Update Dependencies

```bash
# Check for updates
docker-compose run --rm mkdocs pip list --outdated

# Update requirements.txt
# Rebuild
docker-compose build --no-cache
```

### Clean Up

```bash
# Stop and remove containers
docker-compose down

# Remove images
docker-compose down --rmi all

# Remove volumes
docker-compose down -v

# Full cleanup
docker-compose down -v --rmi all --remove-orphans
```

## Security

### Non-Root User

Container runs as non-root user `mkdocs` (UID 1000) for security.

### No Cache

Build uses `PIP_NO_CACHE_DIR=1` to avoid cache poisoning.

### Read-Only Filesystem

For production, mount filesystem as read-only:

```yaml
volumes:
  - ../../:/docs:ro  # Read-only
```

## Performance

### Build Cache

Speed up builds with BuildKit:

```bash
DOCKER_BUILDKIT=1 docker-compose build
```

### Layer Caching

Keep `requirements.txt` stable to leverage layer caching.

### Volume Performance

For better performance on macOS/Windows:

```yaml
volumes:
  - ../../:/docs:cached
```

## Resources

**Documentation:**
- [MkDocs](https://www.mkdocs.org/)
- [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/)
- [Docker Compose](https://docs.docker.com/compose/)

**Container:**
- Base image: `python:3.11-slim`
- Size: ~200MB
- Platforms: linux/amd64, linux/arm64

## License

Same as parent project.
