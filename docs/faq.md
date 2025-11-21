# FAQ

## General
**Q: What is x-ui?**
A: x-ui is a web-based management panel for Xray-core, supporting multiple proxy protocols and user management.

**Q: Which platforms are supported?**
A: Linux, Windows, macOS. Docker is recommended for deployment.

## Development
**Q: How do I start the backend?**
A: Run `go run main.go` or use Air for hot reload.

**Q: How do I build the frontend?**
A: Use `npm run build` or the provided scripts in `web/assets/js/`.

**Q: Where are the main config files?**
A: See `config/` for app config, and `xray/` for Xray-core integration.

## Troubleshooting
**Q: Xray-core fails to start, what should I do?**
A: Check the generated config, logs in `logger/`, and ensure all required fields are set.

**Q: Importing a link fails, how can I debug?**
A: Review the error message for details. See [Error Handling](./error_handling.md) for more info.

## Contribution
**Q: How can I contribute?**
A: Fork the repo, create a branch, make changes, and submit a pull request. See [Contributing](./contributing.md).

---

For more, see the full documentation in this folder.