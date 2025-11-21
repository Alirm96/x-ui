# Default Inbounds Feature

## Overview

The X-UI application now automatically creates default SOCKS and HTTP proxy inbounds on first run, providing an out-of-the-box experience for users. These default inbounds are automatically configured to route traffic through the best available outbound when auto-routing is enabled.

## Default Inbounds Configuration

### SOCKS Proxy (Port 20808)
- **Tag**: `default-socks`
- **Protocol**: SOCKS5
- **Listen Address**: `127.0.0.1` (localhost only)
- **Port**: `20808`
- **Authentication**: No authentication required
- **UDP Support**: Enabled
- **Sniffing**: Enabled for HTTP and TLS traffic

### HTTP Proxy (Port 20809)
- **Tag**: `default-http`
- **Protocol**: HTTP
- **Listen Address**: `127.0.0.1` (localhost only)
- **Port**: `20809`
- **Authentication**: No authentication required
- **Timeout**: 300 seconds
- **Sniffing**: Enabled for HTTP and TLS traffic

## How It Works

### 1. First-Run Initialization

When the application starts for the first time:
1. After database initialization, the `InitDefaultInbounds()` function is called
2. The function checks if default inbounds already exist (by checking for tags `default-socks` and `default-http`)
3. If they don't exist, it creates both SOCKS and HTTP inbounds with the configurations above
4. The inbounds are automatically enabled and ready to use

### 2. Auto-Routing Integration

When auto-routing is enabled:
1. The `OutboundAutoRouteJob` periodically tests all enabled outbounds
2. It selects the best performing outbound based on latency
3. It automatically creates/updates routing rules that direct traffic from the default inbounds to the best outbound
4. When the best outbound changes over time, the routing rules are automatically updated

### 3. Routing Rules

The routing rule for default inbounds looks like this:

```json
{
  "type": "field",
  "inboundTag": ["default-socks", "default-http"],
  "outboundTag": "best-outbound-tag"
}
```

This rule ensures that any traffic coming from the default SOCKS or HTTP inbounds is routed through the currently selected best outbound.

### 4. Automatic Routing Rule Management

The routing rule for default inbounds is **automatically managed** by the system:

**When Auto-Routing is Enabled:**
- A routing rule is automatically created directing traffic from default inbounds to the best outbound
- The rule appears in `/xui/routing` with a green "Auto" badge
- The rule **cannot be deleted manually** - attempting to delete it shows a warning
- The rule **can be reordered** - you can move it up/down/first/last like any other rule
- When the best outbound changes, the rule automatically updates to use the new best outbound

**When Auto-Routing is Disabled:**
- The routing rule is automatically removed from the configuration
- Default inbounds continue to exist but have no special routing

**Visual Indicators:**
- **Green "Auto" Badge**: Marks the rule as system-managed in the routing table
- **Disabled Delete Button**: The delete option is grayed out and shows a tooltip
- **Warning Message**: If you try to delete it, you'll see: "This is an automatic routing rule for default inbounds. It cannot be deleted manually. Disable auto-routing in outbound settings to remove it."

## Usage

### For End Users

1. **Install and Run X-UI**: On first run, the default inbounds are automatically created
2. **Configure Outbounds**: Add your proxy server outbounds in the `/xui/outbounds` page
3. **Enable Auto-Routing**: In outbound settings, enable auto-routing and configure test schedule
4. **Use the Proxies**: 
   - Configure your browser/applications to use:
     - SOCKS5: `127.0.0.1:20808`
     - HTTP: `127.0.0.1:20809`
5. **View Routing**: Check `/xui/routing` to see the automatic routing rule (marked with "Auto" badge)

### For Developers

#### Code Structure

**Initialization (`web/service/inbound.go`)**:
```go
func (s *InboundService) InitDefaultInbounds() error
```
- Called from `main.go` after database initialization
- Checks if default inbounds exist, creates them if not
- Uses standard `AddInbound` method to ensure proper validation and setup

**Routing Rules (`web/service/inbound.go`)**:
```go
func (s *InboundService) CreateDefaultRoutingRules(bestOutboundTag string) error
```
- Called by `OutboundAutoRouteJob` when best outbound is selected
- Creates or updates routing rules for default inbounds
- Ensures traffic flows through the best available outbound

**Auto-Routing Job (`web/job/outbound_job.go`)**:
```go
type OutboundAutoRouteJob struct
```
- Runs on schedule (default: every 4 hours)
- Gets best outbound based on test results
- Applies routing rules for default inbounds
- Restarts xray to apply configuration changes

## Configuration Options

### Changing Default Ports

If you need to use different ports, you can modify the default inbounds through the UI:
1. Go to `/xui/inbounds`
2. Find the default SOCKS or HTTP inbound
3. Edit the port number
4. Save changes

### Disabling Default Inbounds

If you don't want to use the default inbounds:
1. Go to `/xui/inbounds`
2. Find the default inbound you want to disable
3. Toggle the "Enable" switch to off
4. Save changes

### Removing Default Inbounds

To permanently remove default inbounds:
1. Go to `/xui/inbounds`
2. Find the default inbound
3. Delete it using the delete button
4. The inbound will not be recreated unless you reset the database

## Technical Details

### Database Schema

Default inbounds are stored in the `inbounds` table with:
- `tag`: `default-socks` or `default-http` (unique identifier)
- `protocol`: `socks` or `http`
- `port`: `10808` or `10809`
- `listen`: `127.0.0.1` (localhost only for security)
- `enable`: `true` (enabled by default)

### Security Considerations

1. **Localhost Only**: Default inbounds listen on `127.0.0.1` only, preventing external access
2. **No Authentication**: For localhost access, authentication is not required (you can add it manually if needed)
3. **Sniffing Enabled**: Traffic is automatically detected and routed appropriately

### Auto-Routing Behavior

- **Best Outbound Selection**: Based on lowest latency from recent tests
- **Automatic Updates**: Routing rules update when best outbound changes
- **Xray Restart**: When routing rules change, xray is automatically restarted to apply changes
- **Fallback**: If no suitable outbound is found, traffic uses the previous routing configuration

## Troubleshooting

### Ports Already in Use

**Error**: `bind: address already in use`

**Solution**:
1. Check if another application is using ports 10808 or 10809:
   ```bash
   lsof -i :10808
   lsof -i :10809
   ```
2. Either stop the conflicting application or change the default inbound ports in the UI

### Routing Not Working

**Issue**: Traffic not flowing through best outbound

**Checklist**:
1. Verify auto-routing is enabled in outbound settings
2. Check that outbounds have been tested successfully
3. Verify routing rules exist in `/xui/xray` configuration
4. Check xray logs for errors
5. Ensure xray service is running

### Default Inbounds Not Created

**Issue**: Default inbounds don't appear on first run

**Possible Causes**:
1. Database initialization failed
2. Ports 10808/10809 were already in use
3. Check application logs for errors

**Solution**:
```bash
# Check logs
docker compose -f docker-compose.dev.yml logs | grep -i "default"

# Restart application
make dev-restart
```

## API Reference

### Checking Default Inbounds Status

```bash
# Get all inbounds
curl -X POST http://localhost:54321/xui/API/inbounds/list \
  -H "Content-Type: application/json" \
  -d '{"userId": 1}'
```

### Getting Routing Configuration

```bash
# Get xray configuration
curl -X POST http://localhost:54321/xui/API/xray/config \
  -H "Content-Type: application/json"
```

## Future Enhancements

Potential improvements for this feature:
1. **Customizable Default Ports**: Allow users to configure default ports via environment variables
2. **Authentication Options**: Optional authentication for default inbounds
3. **Multiple Protocol Support**: Add VMESS/VLESS default inbounds
4. **Load Balancing**: Round-robin or weighted distribution across multiple outbounds
5. **Geo-Routing**: Route based on destination country/region
6. **Traffic Statistics**: Per-inbound traffic monitoring and analytics

## Related Documentation

- [Development Guide](DEVELOPMENT.md) - Setting up development environment
- [Auto-Routing Feature](AUTO_ROUTING.md) - Details on auto-routing functionality
- [Outbound Management](OUTBOUNDS.md) - Managing outbound configurations
- [Routing Rules](ROUTING.md) - Advanced routing configuration

## Support

For issues or questions:
1. Check the troubleshooting section above
2. Review application logs
3. Open an issue on GitHub
4. Consult the community forums
