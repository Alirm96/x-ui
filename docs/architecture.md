# Architecture

## High-Level Structure

```
+-------------------+
|    Frontend (UI)  |  <--- Vue.js SPA (web/html/xui, web/assets/js)
+-------------------+
           |
           | HTTP/REST API
           v
+-------------------+
|   Backend (API)   |  <--- Go (web/controller, web/service)
+-------------------+
           |
           | Xray config, process control
           v
+-------------------+
|   Xray-core       |  <--- Managed subprocess
+-------------------+
           |
           | SQLite DB
           v
+-------------------+
|   Database        |  <--- SQLite (database/)
+-------------------+
```

## Key Directories
- `web/` - Backend API, controllers, services, web assets
- `web/html/xui/` - Vue.js UI templates
- `web/assets/js/` - Frontend JS (Vue, utils)
- `database/` - DB models, migrations
- `xray/` - Xray-core integration, config generation
- `config/` - App config, versioning
- `main.go` - Entry point

## Data Flow
1. User interacts with the web UI (Vue.js)
2. UI sends REST API requests to Go backend
3. Backend processes requests, updates DB, manages Xray-core
4. Backend returns data/status to UI
5. UI updates accordingly

---

See [Backend](./backend.md) and [Frontend](./frontend.md) for more details.