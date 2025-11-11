# Production Docker Configuration

Base Docker image and production deployment configuration for OneMount.

For development and testing, see `docker/` directory.

## Files

- **Dockerfile** - Base image (Ubuntu 24.04, Go 1.24.2, FUSE3, build tools)
- **docker-compose.yml** - Production deployment configuration
- **.dockerignore** - Build context exclusions

## Base Image

The base image provides the foundation for all OneMount containers:

- Ubuntu 24.04 LTS
- Go 1.24.2
- FUSE3 support
- GUI dependencies (WebKit2GTK)
- Build tools with CGO support
- IPv4-only networking

Used by:
- `docker/Dockerfile.test-runner` - Test execution
- `docker/Dockerfile.github-runner` - CI/CD runners
- `packaging/deb/docker/Dockerfile` - Debian package builder

## Production Deployment

### Start Production Container

```bash
docker compose -f packaging/docker/docker-compose.yml up -d
```

### Configuration

The production container includes:
- FUSE support for filesystem operations
- Resource limits (2GB RAM, 2 CPUs)
- Health checks (mount point monitoring)
- Automatic restart policy
- Persistent volumes for data, config, and cache

### Volumes

- `onemount-data` - OneDrive mount point
- `onemount-config` - Configuration files
- `onemount-cache` - Cache directory

### Environment Variables

- `ONEMOUNT_VERSION` - Version tag (default: latest)
- `ONEMOUNT_LOG_LEVEL` - Log level (default: info)
- `ONEMOUNT_MOUNT_POINT` - Mount point (default: /mnt/onedrive)

## See Also

- Development Docker: `docker/README.md`
- Build configuration: `docker/compose/docker-compose.build.yml`
- Test configuration: `docker/compose/docker-compose.test.yml`
