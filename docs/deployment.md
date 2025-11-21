# Deployment

## Overview
x-ui can be deployed on Linux, Windows, and macOS. Docker is the recommended method for production deployments.

## Docker Deployment
1. Build or pull the Docker image
2. Use `docker-compose.yml` for multi-service setup
3. Configure environment variables and volumes
4. Start with `docker-compose up -d`

## Manual Deployment
- Build Go backend: `go build -o x-ui main.go`
- Build frontend assets: `npm run build` (if using Node.js)
- Copy binaries and assets to server
- Configure systemd or init scripts (see `x-ui.service`)

## Updating
- Pull latest code
- Rebuild and restart services
- Run DB migrations if needed

## Troubleshooting
- Check logs in `logger/` and Docker logs
- Validate Xray config before applying

---

See [Contributing](./contributing.md) for how to get involved.