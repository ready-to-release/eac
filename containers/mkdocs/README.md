# MkDocs Docker Container

**Environment**: `mkdocs-docker`  
**Contract**: `contracts/environments/0.1.0/environments.yml`

## Purpose

Provides a Docker container for serving project documentation using MkDocs with the Material theme.

## Dependencies

- Docker
- `@deps:docker`

## Usage

### Build the image

```bash
docker build -t eac-mkdocs -f containers/mkdocs/.Dockerfile containers/mkdocs
```

### Run the container

```bash
docker run -d \
  --name eac-mkdocs \
  -p 8000:8000 \
  -v $(pwd)/docs:/docs \
  eac-mkdocs
```

### Stop the container

```bash
docker stop eac-mkdocs
docker rm eac-mkdocs
```

## Environment Tags

- `@L2` - Local environment  
- `@env:mkdocs-docker` - MkDocs Docker environment
- `@deps:docker` - Requires Docker

## Container Details

- **Base Image**: python:3.11-slim
- **MkDocs Version**: 1.5.3
- **Material Theme**: 9.5.3
- **Port**: 8000
- **Working Directory**: /docs

## Integration with Tests

BDD tests tagged with `@env:mkdocs-docker` will automatically use this container when the test infrastructure detects the tag.
