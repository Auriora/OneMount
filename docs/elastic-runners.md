# 🚀 Elastic GitHub Actions Runners

This guide shows how to set up auto-scaling (elastic) GitHub Actions runners for OneMount, providing dynamic scaling based on workflow queue length.

## 🎯 Overview

The elastic runner system automatically:
- **Scales up** when workflows are queued (default: 2+ queued jobs)
- **Scales down** when queue is empty (with configurable cooldown)
- **Maintains minimum runners** for immediate availability
- **Respects maximum limits** to control resource usage

## ⚡ Quick Start

### 1. Prerequisites

Ensure you have:
- ✅ Existing runner setup (`.env` file configured)
- ✅ Docker access to remote host (172.16.1.104:2376)
- ✅ `jq` installed for JSON parsing

### 2. Complete Setup

```bash
# One-command setup
./scripts/setup-elastic-runners.sh setup
```

This will:
1. Build the elastic runner image
2. Configure the elastic manager
3. Start the auto-scaling system

### 3. Monitor Status

```bash
# Check current status
./scripts/elastic-runner-manager.sh status

# Start monitoring dashboard
./scripts/setup-elastic-runners.sh start-dashboard
# Visit: http://172.16.1.104:8080
```

## 🔧 Configuration

### Environment Variables

Configure scaling behavior with these variables:

```bash
# Scaling limits
MIN_RUNNERS=1              # Minimum runners (always running)
MAX_RUNNERS=5              # Maximum runners (resource limit)

# Scaling triggers
SCALE_UP_THRESHOLD=2       # Queue length to trigger scale up
SCALE_DOWN_THRESHOLD=0     # Queue length to trigger scale down

# Timing
CHECK_INTERVAL=30          # Check frequency (seconds)
COOLDOWN_PERIOD=300        # Cooldown between scaling actions (seconds)

# Docker host
DOCKER_HOST=172.16.1.104:2376  # Remote Docker host
```

### Custom Configuration Example

```bash
# High-capacity setup
MIN_RUNNERS=2 MAX_RUNNERS=10 SCALE_UP_THRESHOLD=3 \
./scripts/elastic-runner-manager.sh monitor
```

## 🛠️ Management Commands

### Elastic Manager

```bash
# Start auto-scaling (runs continuously)
./scripts/elastic-runner-manager.sh monitor

# Check current status
./scripts/elastic-runner-manager.sh status

# Manual scaling
./scripts/elastic-runner-manager.sh scale-up 3
./scripts/elastic-runner-manager.sh scale-down 1

# Clean up all elastic runners
./scripts/elastic-runner-manager.sh cleanup
```

### System Management

```bash
# Start/stop elastic system
./scripts/setup-elastic-runners.sh start
./scripts/setup-elastic-runners.sh stop

# Rebuild image
./scripts/setup-elastic-runners.sh build

# Install as systemd service
./scripts/setup-elastic-runners.sh install-service
```

## 📊 Monitoring

### Real-time Monitoring

```bash
# Watch status continuously
watch -n 10 './scripts/elastic-runner-manager.sh status'

# Follow elastic manager logs
DOCKER_HOST=tcp://172.16.1.104:2376 docker logs -f onemount-elastic-manager
```

### Web Dashboard

```bash
# Start dashboard
./scripts/setup-elastic-runners.sh start-dashboard

# Access at: http://172.16.1.104:8080
```

### Systemd Service

```bash
# Install service
./scripts/setup-elastic-runners.sh install-service

# Control service
sudo systemctl start onemount-elastic-runner
sudo systemctl status onemount-elastic-runner
sudo systemctl stop onemount-elastic-runner

# View logs
journalctl -u onemount-elastic-runner -f
```

## 🔄 How It Works

### Scaling Logic

1. **Monitor**: Checks GitHub Actions queue every 30 seconds
2. **Scale Up**: When queue ≥ threshold, adds runners up to maximum
3. **Scale Down**: When queue ≤ threshold, removes idle runners to minimum
4. **Cooldown**: Waits 5 minutes between scaling actions

### Runner Lifecycle

```
Queue: 0 → MIN_RUNNERS (1) running
Queue: 2 → Scale up to 2 runners
Queue: 5 → Scale up to 5 runners (MAX_RUNNERS)
Queue: 0 → Scale down to 1 runner (MIN_RUNNERS)
```

### Container Management

- **Naming**: `onemount-runner-elastic-1`, `onemount-runner-elastic-2`, etc.
- **Labels**: `self-hosted,Linux,onemount-testing,optimized,elastic`
- **Volumes**: Separate workspace and work volumes per runner
- **Cleanup**: Automatic removal of stopped containers

## 🎛️ Advanced Configuration

### Custom Scaling Strategy

Create a custom configuration file:

```bash
# .env.elastic
MIN_RUNNERS=2
MAX_RUNNERS=8
SCALE_UP_THRESHOLD=1
SCALE_DOWN_THRESHOLD=0
CHECK_INTERVAL=15
COOLDOWN_PERIOD=180
```

Use with:
```bash
source .env.elastic
./scripts/elastic-runner-manager.sh monitor
```

### Integration with Existing Runners

The elastic system works alongside your existing static runners:
- Static runners: `onemount-runner-remote-optimized`
- Elastic runners: `onemount-runner-elastic-*`

Both use the same labels and can handle the same workflows.

## 🚨 Troubleshooting

### Common Issues

**Runners not scaling up:**
```bash
# Check GitHub API access
curl -H "Authorization: token $GITHUB_TOKEN" \
     https://api.github.com/repos/$GITHUB_REPOSITORY/actions/runs

# Check Docker connectivity
DOCKER_HOST=tcp://172.16.1.104:2376 docker ps
```

**Containers not starting:**
```bash
# Check container logs
DOCKER_HOST=tcp://172.16.1.104:2376 docker logs onemount-runner-elastic-1

# Check image availability
DOCKER_HOST=tcp://172.16.1.104:2376 docker images | grep onemount-github-runner
```

**Manager not responding:**
```bash
# Restart elastic manager
./scripts/setup-elastic-runners.sh stop
./scripts/setup-elastic-runners.sh start

# Check manager logs
DOCKER_HOST=tcp://172.16.1.104:2376 docker logs onemount-elastic-manager
```

### Debug Mode

```bash
# Enable verbose logging
set -x
./scripts/elastic-runner-manager.sh status
set +x
```

## 📈 Performance Benefits

### Resource Efficiency
- **Idle Cost**: Only minimum runners consume resources when idle
- **Peak Capacity**: Automatically scales to handle workflow bursts
- **Cost Savings**: Reduces unnecessary runner uptime

### Workflow Performance
- **Faster Starts**: Minimum runners provide immediate availability
- **Parallel Execution**: Multiple runners handle concurrent workflows
- **Queue Reduction**: Auto-scaling prevents workflow queuing

### Example Scenarios

**Low Activity** (nights/weekends):
- 1 runner active (MIN_RUNNERS)
- Immediate response for occasional workflows

**High Activity** (peak development):
- 5 runners active (MAX_RUNNERS)
- Parallel execution of multiple workflows
- No queuing delays

**Burst Activity** (releases):
- Rapid scale-up to handle deployment workflows
- Automatic scale-down after completion

## 🔗 Integration

### Workflow Configuration

Your workflows automatically use elastic runners with existing labels:

```yaml
runs-on: [self-hosted, Linux, onemount-testing, optimized]
```

Both static and elastic runners will pick up these jobs.

### Monitoring Integration

Integrate with your monitoring stack:

```bash
# Prometheus metrics endpoint (future enhancement)
curl http://172.16.1.104:8080/metrics

# JSON status API (future enhancement)
curl http://172.16.1.104:8080/api/status
```

## 🎉 Success!

Your elastic runner system is now:
- ✅ **Auto-scaling** based on workflow demand
- ✅ **Resource efficient** with minimum idle cost
- ✅ **Highly available** with immediate response
- ✅ **Easily manageable** with simple commands

Monitor your workflows and adjust scaling parameters as needed for optimal performance!
