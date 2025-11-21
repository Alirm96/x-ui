# X-UI Development Guide

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Git
- (Optional) VS Code with Go extension

### Initial Setup

1. **Clone the repository** (if not already done)
   ```bash
   git clone https://github.com/Alirm96/x-ui.git
   cd x-ui
   ```

2. **Create environment file**
   ```bash
   cp .env.example .env
   # Edit .env if needed
   ```

3. **Start development environment**
   ```bash
   # Using Makefile (recommended)
   make dev-up
   
   # Or using docker compose directly
   docker compose -f docker-compose.dev.yml up -d
   ```

4. **Check logs**
   ```bash
   make dev-logs
   # Or: docker logs x-ui-dev -f
   ```

5. **Access the panel**
   - URL: http://localhost:54321
   - Default credentials: `admin` / `admin`
   - **⚠️ Change password immediately after first login!**

## 🔥 Hot Reload Development

The development environment uses [Air](https://github.com/cosmtrek/air) for automatic Go code reloading.

### How It Works
- **Go files** (`.go`): Automatically rebuild on save (~5-10 seconds)
- **HTML/JS/CSS**: Changes reflected immediately (volume-mounted)
- **Translations** (`.toml`): Require container restart

### Development Workflow

1. **Edit Go code**
   ```bash
   vim web/controller/xui.go
   # Save file - Air detects and rebuilds automatically
   ```

2. **Watch rebuild progress**
   ```bash
   make dev-logs
   # Look for "building..." and "running..." messages
   ```

3. **Test changes**
   ```bash
   curl http://localhost:54321/xui/routing
   ```

### What Gets Hot-Reloaded?
✅ Controllers (`web/controller/*.go`)  
✅ Services (`web/service/*.go`)  
✅ Models (`database/model/*.go`)  
✅ Jobs (`web/job/*.go`)  
✅ HTML templates (`web/html/**/*.html`)  
✅ Static assets (`web/assets/**/*`)  

❌ Docker configuration changes (requires restart)  
❌ Air configuration (`.air.toml`) changes (requires restart)

## 📁 Project Structure

```
x-ui/
├── .air.toml                    # Air hot-reload config
├── docker-compose.yml           # Production config
├── docker-compose.dev.yml       # Development config (use this!)
├── Dockerfile                   # Production image build
├── .env.example                 # Environment variables template
├── main.go                      # Application entry point
├── config/                      # App configuration
├── database/                    # Database models
│   └── model/
├── web/                         # Web application
│   ├── controller/              # HTTP request handlers
│   ├── service/                 # Business logic
│   ├── job/                     # Background jobs (cron)
│   ├── html/                    # Templates (hot-reload)
│   ├── assets/                  # Static files (hot-reload)
│   └── translation/             # i18n files
├── xray/                        # Xray integration
├── db/                          # SQLite database (gitignored)
├── cert/                        # SSL certificates (gitignored)
└── bin/                         # Binaries (xray, x-ui)
```

## 🛠️ Common Development Tasks

### Starting/Stopping

```bash
# Start development container
make dev-up

# Stop development container
make dev-down

# Restart development container
make dev-restart

# View logs
make dev-logs

# Stop and remove everything (including volumes)
docker compose -f docker-compose.dev.yml down -v
```

### Database Management

```bash
# Reset database (deletes all data!)
make dev-db-reset

# Backup database
make dev-db-backup

# Access SQLite CLI
docker exec -it x-ui-dev sqlite3 /etc/x-ui/x-ui.db
```

### Debugging

```bash
# Enter container shell
make dev-shell

# Check running processes
docker exec x-ui-dev ps aux

# Check Air status
docker exec x-ui-dev cat /app/tmp/build-errors.log

# View real-time rebuild
make dev-logs
```

### Testing API Endpoints

```bash
# Login and save session
curl -c /tmp/cookies.txt -X POST http://localhost:54321/login \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin&password=admin"

# Test authenticated endpoint
curl -b /tmp/cookies.txt http://localhost:54321/xui/routing

# Test outbound API
curl -b /tmp/cookies.txt -X POST http://localhost:54321/xui/API/outbounds \
  -H "Content-Type: application/json" \
  -d '{"remark":"test","protocol":"vmess","address":"example.com","port":443}'
```

## 🐛 Troubleshooting

### Container won't start
```bash
# Check logs for errors
docker logs x-ui-dev

# Check if port 54321 is already in use
sudo lsof -i :54321

# Remove old container and try again
docker rm -f x-ui-dev
make dev-up
```

### Code changes not reflecting
```bash
# Check if Air is running
docker exec x-ui-dev ps aux | grep air

# Check Air build errors
docker exec x-ui-dev cat /app/tmp/build-errors.log

# Force rebuild by touching a file
touch main.go

# Or restart container
make dev-restart
```

### "Permission denied" errors
```bash
# Fix db directory permissions
sudo chown -R $(id -u):$(id -g) db/

# Fix cert directory permissions
sudo chown -R $(id -u):$(id -g) cert/
```

### Xray not starting
```bash
# Check if xray binary exists
docker exec x-ui-dev ls -la /app/bin/xray*

# Download xray manually
docker exec -w /app/bin x-ui-dev bash -c \
  'wget -q https://github.com/XTLS/Xray-core/releases/download/v25.9.11/Xray-linux-64.zip && \
   unzip -q Xray-linux-64.zip && rm Xray-linux-64.zip && chmod +x xray && \
   ln -sf xray xray-linux-amd64'
```

### Build is too slow
```bash
# Clear Go module cache
docker volume rm x-ui-dev-go-mod-cache

# Clear Air tmp cache
docker volume rm x-ui-dev-air-tmp

# Rebuild from scratch
make dev-down
make dev-up
```

## 🔧 Configuration Files

### `.air.toml`
Air hot-reload configuration. Key settings:
- `cmd`: Build command with flags
- `bin`: Output binary path
- `exclude_dir`: Directories to ignore
- `include_ext`: File extensions to watch
- `delay`: Milliseconds before rebuild

### `docker-compose.dev.yml`
Development Docker Compose config with:
- Air image for hot-reload
- Volume mounts for code
- Go module caching
- Health checks
- Debug mode enabled

## 📝 Adding New Features

### 1. Add New Controller Method
```go
// web/controller/my_controller.go
func (a *MyController) newFeature(c *gin.Context) {
    html(c, "my-feature.html", "pages.myFeature.title", nil)
}
```

### 2. Register Route
```go
// web/controller/my_controller.go - initRouter()
g.GET("/my-feature", a.newFeature)
```

### 3. Create HTML Template
```html
<!-- web/html/xui/my-feature.html -->
{{ template "header" . }}
<!-- Your content -->
{{ template "footer" . }}
```

### 4. Add Translations
```toml
# web/translation/translate.en_US.toml
[pages.myFeature]
"title" = "My Feature"
"description" = "Feature description"
```

### 5. Save and Test
Air will automatically rebuild. Test at: `http://localhost:54321/xui/my-feature`

## 🎯 Best Practices

1. **Always use `make` commands** for common tasks
2. **Watch logs** when developing (`make dev-logs`)
3. **Commit often** - hot-reload makes iteration fast
4. **Test API with curl** before frontend integration
5. **Use descriptive commit messages**
6. **Don't commit** `db/`, `cert/`, or `.env` files
7. **Update translations** for all languages when adding UI text
8. **Check errors** in Air logs after saving Go files

## 🌐 VS Code Integration

### Recommended Extensions
- `golang.go` - Go language support
- `ms-azuretools.vscode-docker` - Docker support
- `esbenp.prettier-vscode` - Code formatting
- `dbaeumer.vscode-eslint` - JavaScript linting

### Workspace Settings
See `.vscode/settings.json` for configured:
- Go formatting on save
- Auto imports
- Test flags
- Build tags

## 🔐 Security Notes

⚠️ **Development environment only!** Do NOT use in production:
- Default credentials enabled
- Debug mode active
- Verbose logging
- No rate limiting
- Hot-reload overhead

For production deployment, use `docker-compose.yml` and build a proper image.

## 📚 Additional Resources

- [Air Documentation](https://github.com/cosmtrek/air)
- [Gin Framework](https://gin-gonic.com/docs/)
- [Xray Documentation](https://xtls.github.io/)
- [Go Best Practices](https://golang.org/doc/effective_go)

## 💡 Tips & Tricks

### Fast Container Shell Access
```bash
alias xui-shell='docker exec -it x-ui-dev sh'
```

### Watch Logs with Filtering
```bash
make dev-logs | grep -i "error\|warning\|panic"
```

### Quick Database Query
```bash
docker exec x-ui-dev sqlite3 /etc/x-ui/x-ui.db "SELECT * FROM outbounds;"
```

### Auto-reload Browser on Changes
Use browser extensions like LiveReload or BrowserSync

---

**Happy coding! 🚀**

For questions or issues, check the main [README.md](README.md) or open an issue on GitHub.
