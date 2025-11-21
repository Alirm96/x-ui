# Backend (Go)

## Overview
The backend is written in Go, using the Gin web framework. It exposes a RESTful API for the frontend, manages the SQLite database, and controls the Xray-core process.

### Main Responsibilities
- Handle API requests (controllers)
- Business logic (services)
- Database access (models)
- Xray-core config generation and process management
- Error handling and logging

## Key Directories & Files
- `main.go`: Application entry point, server setup
- `web/controller/`: API endpoints (e.g., outbound, inbound, user)
- `web/service/`: Business logic, Xray integration, migration
- `database/`: DB models, migrations
- `xray/`: Xray-core process, config generation
- `logger/`: Logging utilities

## API Routing Example
```go
// web/controller/outbound.go
func RegisterOutboundRoutes(router *gin.RouterGroup) {
    router.POST("/import", ImportOutbound)
    router.POST("/migrate", MigrateOutbound)
    // ...
}
```

## Xray-core Integration
- Generates config from DB
- Starts/stops/reloads Xray process
- Handles protocol-specific logic (VLESS, VMess, Trojan, SS)

---

See [API Reference](./api.md) for endpoint details.