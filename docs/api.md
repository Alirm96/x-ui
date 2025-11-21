# API Reference

## Overview
The backend exposes a RESTful API for all frontend operations. Endpoints are grouped by resource (outbound, inbound, user, etc.).

## Example Endpoints
- `POST /api/outbound/import` - Import outbounds from links
- `POST /api/outbound/migrate` - Migrate VLESS outbounds
- `GET /api/outbound/list` - List all outbounds
- `POST /api/inbound/add` - Add inbound
- `GET /api/user/list` - List users

## Request/Response Format
- **Request:** JSON body for POST/PUT
- **Response:**
  ```json
  {
    "success": true,
    "msg": "Operation successful",
    "obj": {...}
  }
  ```
- Errors are returned with `success: false` and a descriptive `msg`.

## Adding New Endpoints
- Create handler in `web/controller/`
- Implement logic in `web/service/`
- Register route in `main.go` or router setup

---

See [Error Handling](./error_handling.md) for error response patterns.