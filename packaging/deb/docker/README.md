# Debian Package Builder

Docker image for building OneMount Debian packages. Extends the base image with Debian packaging tools.

## Usage

### Build Debian Package

```bash
# Using Docker Compose (recommended)
docker compose -f docker/compose/docker-compose.build.yml --profile deb run --rm build-deb

# Output: dist/onemount_*.deb
```

### Build Binaries Only

```bash
docker compose -f docker/compose/docker-compose.build.yml --profile binaries run --rm build-binaries

# Output: build/binaries/onemount, build/binaries/onemount-launcher
```

### Clean Build Artifacts

```bash
docker compose -f docker/compose/docker-compose.build.yml --profile clean run --rm clean-build
```

## Build Entrypoint Commands

The image uses `docker/scripts/build-entrypoint.sh` which provides:

- `binaries` - Build OneMount binaries
- `deb` - Build Debian package
- `clean` - Clean build artifacts
- `help` - Show help message

### Options

- `--verbose` - Enable verbose output
- `--output DIR` - Output directory (default: /workspace/build)

## Output Locations

- **Binaries**: `build/binaries/`
- **Debian packages**: `dist/`

## See Also

- Base image: `packaging/docker/Dockerfile`
- Build compose: `docker/compose/docker-compose.build.yml`
- Build entrypoint: `docker/scripts/build-entrypoint.sh`
