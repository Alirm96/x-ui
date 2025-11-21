# X-UI Development Quick Reference

## 🚀 Getting Started

```bash
# First time setup
make dev-up              # Start development environment
make dev-install-xray    # Install xray binary (first time only)
```

Access: http://localhost:54321 (admin/admin)

## 📝 Common Commands

```bash
make dev-up          # Start dev environment
make dev-down        # Stop dev environment
make dev-restart     # Restart dev environment
make dev-logs        # View logs (Ctrl+C to exit)
make dev-shell       # Enter container shell
make dev-status      # Check environment status
```

## 🔧 Development Workflow

### Editing Go Code
1. Edit `.go` file
2. Save file
3. Air auto-rebuilds (~5-10 seconds)
4. Test changes

### Editing HTML/JS/CSS
1. Edit files in `web/html/` or `web/assets/`
2. Save file
3. Changes instant (refresh browser)

### Editing Translations
1. Edit `.toml` files in `web/translation/`
2. Save file
3. Requires: `make dev-restart`

## 🗄️ Database

```bash
make dev-db-backup    # Backup database
make dev-db-reset     # Reset database (DELETES ALL DATA!)
```

## 🐛 Debugging

```bash
# View logs
make dev-logs

# Enter container
make dev-shell

# Check Air status
docker exec x-ui-dev cat /app/tmp/build-errors.log

# Manual restart
make dev-restart
```

## 📁 Important Directories

```
web/
├── controller/    # HTTP handlers (Go hot-reload)
├── service/       # Business logic (Go hot-reload)
├── job/           # Background jobs (Go hot-reload)
├── html/          # Templates (instant update)
├── assets/        # CSS/JS (instant update)
└── translation/   # i18n files (needs restart)
```

## 🔐 Default Credentials

- **Username:** `admin`
- **Password:** `admin`

⚠️ Change immediately after first login!

## 🆘 Troubleshooting

| Problem                     | Solution                         |
| --------------------------- | -------------------------------- |
| Container won't start       | `make dev-down && make dev-up`   |
| Code changes not reflecting | Check `make dev-logs` for errors |
| Port 54321 in use           | `sudo lsof -i :54321`            |
| Xray not working            | `make dev-install-xray`          |
| Slow rebuild                | `make dev-clean && make dev-up`  |

## 📖 Full Documentation

See [DEVELOPMENT.md](DEVELOPMENT.md) for comprehensive guide.

## 💡 Tips

- Use `make` commands (not docker compose directly)
- Keep logs open in separate terminal: `make dev-logs`
- Test API with curl before frontend work
- Commit often - hot-reload makes iteration fast!

---

**Quick Test:**
```bash
curl http://localhost:54321/login
```
