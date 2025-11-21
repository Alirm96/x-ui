# Database

## Overview
x-ui uses SQLite as its default database for storing users, outbounds, inbounds, and other configuration data. The database schema is defined in Go structs and managed via the backend.

## Key Concepts
- **Models:** Defined in `database/model/`
- **Migrations:** Handled automatically on startup
- **Relationships:** Outbounds, users, and inbounds are linked via foreign keys

## Main Tables
- `outbound`: Proxy outbounds (protocol, address, port, etc.)
- `inbound`: Proxy inbounds
- `user`: User accounts and credentials
- `settings`: Application settings

## Example Model (Go)
```go
// database/model/model.go
type Outbound struct {
    ID       int
    Protocol string
    Address  string
    Port     int
    // ...
}
```

## Extending the Schema
- Add new fields to the Go struct
- Update migration logic if needed
- Restart the backend to apply changes

---

See [Backend](./backend.md) for how models are used.