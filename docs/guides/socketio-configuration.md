# Socket.IO Configuration Guide

This document describes the Socket.IO realtime notification configuration options available in OneMount.

## Overview

OneMount supports realtime change notifications from Microsoft Graph via Socket.IO connections. This reduces the need for frequent polling and provides near-instantaneous updates when files change in OneDrive.

## Configuration Options

### Basic Configuration

The realtime configuration is specified in the `realtime` section of the OneMount configuration file:

```yaml
realtime:
  enabled: true
  pollingOnly: false
  resource: "/me/drive/root"
  fallbackIntervalSeconds: 1800
  clientState: ""
```

### Configuration Fields

#### `enabled` (boolean)
- **Default**: `false`
- **Description**: Controls whether realtime notifications are active. When `false`, OneMount relies entirely on periodic delta polling.
- **Example**: `enabled: true`

#### `pollingOnly` (boolean)
- **Default**: `false`
- **Description**: Forces delta polling even when realtime subscriptions are configured. This disables the Socket.IO transport while keeping the realtime infrastructure active. Useful for debugging or environments where WebSocket connections are problematic.
- **Example**: `pollingOnly: true`

#### `resource` (string)
- **Default**: `"/me/drive/root"`
- **Description**: Specifies the Microsoft Graph resource path to monitor for changes.
- **Valid Values**:
  - `"/me/drive/root"` - Personal OneDrive root
  - `"/drives/{drive-id}"` - Specific shared drive
  - `"/drives/{drive-id}/root"` - Shared drive root
- **Example**: `resource: "/me/drive/root"`

#### `fallbackIntervalSeconds` (integer)
- **Default**: `1800` (30 minutes)
- **Range**: `30` to `7200` seconds (30 seconds to 2 hours)
- **Description**: The polling interval in seconds when Socket.IO is unavailable or degraded. When Socket.IO is healthy, polling occurs much less frequently (every 30+ minutes).
- **Example**: `fallbackIntervalSeconds: 900`

#### `clientState` (string)
- **Default**: Auto-generated random token
- **Description**: A validation token echoed in notification events. Used to verify that notifications are intended for this client instance. If empty, a random token is generated automatically.
- **Example**: `clientState: "my-custom-client-state"`

## Command Line Options

Several command-line flags can override configuration file settings:

### `--polling-only`
Forces delta polling even if realtime subscriptions are configured.
```bash
onemount --polling-only /mnt/onedrive
```

### `--realtime-fallback-seconds`
Override the realtime fallback polling interval.
```bash
onemount --realtime-fallback-seconds 600 /mnt/onedrive
```

## Behavior Modes

### Realtime Enabled (`enabled: true`, `pollingOnly: false`)
- OneMount establishes a Socket.IO connection to Microsoft Graph
- Change notifications are received immediately
- Delta polling occurs infrequently (every 30+ minutes) as a backup
- Falls back to configured interval if Socket.IO fails

### Polling Only (`enabled: true`, `pollingOnly: true`)
- Socket.IO connection is not established
- Delta polling occurs at the fallback interval
- Realtime infrastructure remains active for potential future use

### Realtime Disabled (`enabled: false`)
- No Socket.IO connection attempts
- Delta polling occurs at regular intervals (typically 5 minutes)
- Minimal overhead from realtime infrastructure

## Health States

The Socket.IO transport reports health states that affect polling behavior:

- **Healthy**: Socket.IO working normally, infrequent polling
- **Degraded**: Connection issues, increased polling frequency
- **Failed**: Socket.IO not working, fallback to configured interval
- **Unknown**: Initial state before connection attempts

## Troubleshooting

### Connection Issues
If Socket.IO connections fail:
1. Check network connectivity to `*.graph.microsoft.com`
2. Verify WebSocket traffic is not blocked by firewalls
3. Enable polling-only mode as a workaround: `pollingOnly: true`

### High Polling Frequency
If polling occurs too frequently:
1. Check Socket.IO health in `onemount --stats`
2. Increase `fallbackIntervalSeconds` if needed
3. Verify realtime is properly enabled and not in polling-only mode

### Authentication Errors
If realtime subscriptions fail to authenticate:
1. Verify OAuth tokens are valid and not expired
2. Check that the specified resource path is accessible
3. Ensure the Microsoft Graph API permissions include subscription access

## Statistics and Monitoring

Use `onemount --stats /mnt/onedrive` to view realtime status:

```
Realtime Notifications:
  Mode: socketio
  Status: healthy
  Missed heartbeats: 0
  Consecutive failures: 0
  Reconnect count: 2
  Last heartbeat: 2024-12-12T10:30:45Z
```

## Best Practices

1. **Enable realtime for better responsiveness**: Set `enabled: true` for near-instant change detection
2. **Use appropriate fallback intervals**: 30 minutes (1800s) is usually sufficient for most use cases
3. **Monitor connection health**: Check stats regularly to ensure Socket.IO is working properly
4. **Have a polling fallback**: Always configure a reasonable fallback interval in case Socket.IO fails
5. **Test in your environment**: Some corporate networks may block WebSocket connections

## Security Considerations

- Client state tokens are used for validation but are not security credentials
- Socket.IO connections use the same OAuth tokens as regular Graph API calls
- WebSocket traffic should be encrypted (WSS) in production environments
- Consider network security policies when enabling realtime notifications

## Migration from Webhooks

OneMount previously supported webhook-based notifications, which have been replaced by Socket.IO:

- **Old**: Required publicly accessible webhook endpoints
- **New**: Uses outbound WebSocket connections to Microsoft Graph
- **Benefits**: No need for public endpoints, better firewall compatibility, more reliable

If you have old webhook configuration, it will be ignored. Update your configuration to use the `realtime` section instead.

## Quick Troubleshooting

### Socket.IO Not Working
1. Check stats: `onemount --stats /mount/path`
2. Look for "Status: failed" in Realtime section
3. Try polling-only mode: `onemount --polling-only /mount/path`

### High Polling Frequency
1. Verify Socket.IO status is "healthy"
2. Check `fallbackIntervalSeconds` is reasonable (1800+ seconds)
3. Ensure `pollingOnly: false` if you want Socket.IO

### Corporate Network Issues
1. Set `pollingOnly: true` in configuration
2. Use shorter `fallbackIntervalSeconds` (300-900 seconds)
3. Contact IT about WebSocket access to `*.graph.microsoft.com`

For comprehensive troubleshooting, see the [Troubleshooting Guide](../user/troubleshooting-guide.md).