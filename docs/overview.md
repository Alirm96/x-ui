# Project Overview

x-ui is a modern, web-based management panel for Xray-core and related proxy protocols (V2Ray, VLESS, Trojan, Shadowsocks, etc.). It provides a user-friendly interface for configuring, managing, and monitoring proxy outbounds/inbounds, users, and server settings. The project is designed for extensibility, maintainability, and ease of use.

## Key Features
- Web UI for managing Xray-core
- Support for multiple protocols: VLESS, VMess, Trojan, Shadowsocks
- User and Outbound management
- Import/export of configuration links
- Docker support
- Multi-platform (Linux, Windows, macOS)
- Database-backed configuration
- Migration tools for protocol changes

## Main Components
- **Backend:** Go (Gin, SQLite, Xray integration)
- **Frontend:** Vue.js (Single Page Application)
- **Database:** SQLite (default), extensible
- **Xray-core:** Managed as a subprocess

---

Continue to [Architecture](./architecture.md) for a detailed breakdown.