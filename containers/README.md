# Containers

This directory contains Docker containers for the CLI project.

## Available Containers

### MkDocs Container

Location: `containers/mkdocs/`

**Purpose:** Build and serve project documentation

**Features:**

- MkDocs with Material theme
- All required plugins
- Live reload development server
- Static site builder

**Quick Start:**

```bash
cd containers/mkdocs

# Serve documentation
./serve.sh

# Build static site
./build.sh

# Interactive shell
./shell.sh
```

See [containers/mkdocs/README.md](mkdocs/README.md) for full documentation.

## General Usage

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+

### Directory Structure

```
containers/
├── README.md              # This file
└── mkdocs/                # MkDocs documentation container
    ├── .Dockerfile        # Container definition
    ├── requirements.txt   # Python dependencies
    ├── docker-compose.yml # Compose configuration
    ├── .dockerignore      # Build exclusions
    ├── serve.sh           # Dev server script
    ├── build.sh           # Build script
    ├── shell.sh           # Interactive shell script
    └── README.md          # Detailed documentation
```

## Common Commands

### MkDocs

```bash
# Start development server
cd containers/mkdocs && docker-compose up

# Build static site
cd containers/mkdocs && docker-compose run --rm mkdocs-build

# Interactive shell
cd containers/mkdocs && ./shell.sh
```

## Adding New Containers

When adding a new container to this directory:

1. **Create subdirectory:**

   ```bash
   mkdir containers/<container-name>
   ```

2. **Add files:**
   - `.Dockerfile` - Container definition
   - `docker-compose.yml` - Compose configuration (optional)
   - `README.md` - Usage documentation
   - Helper scripts (optional)

3. **Update this README:**
   Add entry to "Available Containers" section

4. **Document:**
   - Purpose and features
   - Quick start commands
   - Link to detailed README

## Best Practices

### Dockerfile Naming

Use `.Dockerfile` (with leading dot) to:

- Distinguish from application Dockerfiles
- Keep containers directory organized
- Work with .dockerignore patterns

### Multi-Service Setup

For complex setups, use docker-compose.yml:

```yaml
version: '3.8'

services:
  service1:
    build:
      context: .
      dockerfile: .Dockerfile
    # ...

  service2:
    # ...
```

### Helper Scripts

Provide bash scripts for common operations:

- `serve.sh` - Start development server
- `build.sh` - Build artifacts
- `shell.sh` - Interactive shell
- `test.sh` - Run tests

Make them executable:

```bash
chmod +x *.sh
```

### Documentation

Each container should have its own README.md with:

- Purpose and features
- Prerequisites
- Quick start
- Detailed usage
- Configuration
- Troubleshooting
- Advanced topics

## Security

### Non-Root Users

Run containers as non-root users:

```dockerfile
RUN useradd -m -u 1000 -s /bin/bash appuser
USER appuser
```

### No Secrets in Images

Never include secrets in images:

- Use environment variables
- Use Docker secrets
- Mount secret files

### Minimal Images

Use minimal base images:

- `alpine` for small size
- `slim` for better compatibility
- Official images when available

### Pin Versions

Pin exact versions in requirements:

```txt
# Good
mkdocs==1.5.3

# Bad
mkdocs>=1.5
```

## Performance

### Layer Caching

Order Dockerfile instructions for optimal caching:

```dockerfile
# 1. System dependencies (rarely change)
RUN apt-get update && apt-get install -y ...

# 2. Application dependencies (occasionally change)
COPY requirements.txt .
RUN pip install -r requirements.txt

# 3. Application code (frequently changes)
COPY . .
```

### Multi-Stage Builds

Use multi-stage builds for smaller images:

```dockerfile
FROM base AS builder
# Build steps

FROM base AS runtime
COPY --from=builder /app /app
```

### BuildKit

Enable BuildKit for faster builds:

```bash
DOCKER_BUILDKIT=1 docker build ...
```

## Maintenance

### Update Dependencies

```bash
# Check for updates
docker-compose run --rm <service> pip list --outdated

# Update requirements
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

### Prune System

```bash
# Remove unused containers
docker container prune

# Remove unused images
docker image prune

# Remove unused volumes
docker volume prune

# Remove everything unused
docker system prune -a
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Build Container

on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build container
        run: |
          cd containers/mkdocs
          docker-compose build
```

### GitLab CI

```yaml
build-container:
  image: docker:latest
  services:
    - docker:dind
  script:
    - cd containers/mkdocs
    - docker-compose build
```

## Resources

**Docker:**

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Dockerfile Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)

**Guides:**

- [Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [BuildKit](https://docs.docker.com/build/buildkit/)
- [Docker Security](https://docs.docker.com/engine/security/)

## Future Containers

Potential containers to add:

- **Go Build Container** - Build Go MCP servers
- **Node Build Container** - Build VSCode extension
- **Test Container** - Run automated tests
- **CI Container** - Unified CI/CD environment
- **Deployment Container** - Package for deployment

## Contributing

When contributing containers:

1. Follow the directory structure
2. Include comprehensive README
3. Add helper scripts for common tasks
4. Document all environment variables
5. Include examples in README
6. Test on Linux, macOS, and Windows (if applicable)
7. Update this main README

## Questions?

See individual container READMEs for specific questions:

- [MkDocs Container](mkdocs/README.md)
