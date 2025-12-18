# r2r cleanup

Clean up old extension images and containers to free disk space.

## Syntax

```bash
r2r cleanup [flags]
```

## Description

The `cleanup` command removes old Docker images and stopped containers from extension usage. Use this to reclaim disk space.

**What it cleans:**

- Old extension images (previous versions)
- Stopped containers
- Dangling images
- Build cache

**What it preserves:**

- Currently configured images in `.r2r/r2r-cli.yml`
- Running containers
- Recent images

## Examples

### Basic Cleanup

```bash
r2r cleanup
```

**Output:**

```
Cleaning up R2R resources...

Removed:
  - 3 old extension images (freed 1.2 GB)
  - 5 stopped containers
  - Build cache (freed 450 MB)

Total space freed: 1.65 GB
```

### Check Space Before Cleanup

```bash
# Check Docker disk usage
docker system df

# Run cleanup
r2r cleanup

# Verify space freed
docker system df
```

## Disk Space

### Checking Disk Usage

```bash
# Overall Docker disk usage
docker system df

# Extension images
docker images | grep ready-to-release
```

### When to Run Cleanup

- Disk space running low
- After major extension updates
- Monthly maintenance
- Before pulling large new images

## What Gets Removed

**Extension images:**

- Images not in current `.r2r/r2r-cli.yml`
- Old versions
- Untagged images

**Containers:**

- Stopped extension containers
- Failed container runs

**Build cache:**

- Intermediate build layers
- Unused build cache

## See Also

- [R2R CLI Overview](index.md) - Command overview
- [install command](install.md) - Install extensions
- [verify command](verify.md) - Verify system health
- [Configuration Reference](configuration.md) - Extension configuration
