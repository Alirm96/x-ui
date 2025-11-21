# Configuration & Xray Integration

## Overview
x-ui manages Xray-core as a subprocess, generating its configuration dynamically from the database. It supports multiple protocols and ensures compatibility with Xray's requirements.

## Config Generation
- Reads outbounds, inbounds, users from DB
- Generates a valid Xray JSON config
- Handles protocol-specific quirks (e.g., VLESS users array)
- Writes config to disk before (re)starting Xray

## Xray Process Management
- Start, stop, reload Xray-core
- Monitor process health
- Log Xray output for debugging

## Protocol Support
- **VLESS**: Multi-user, migration logic
- **VMess**: Standard support
- **Trojan**: Strict link format validation
- **Shadowsocks**: Cipher and password handling

## Troubleshooting
- Check logs in `logger/` and Xray output
- Use migration tools for protocol changes
- Validate config with Xray-core before applying

---

See [Error Handling](./error_handling.md) for config error reporting.