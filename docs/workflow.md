# Development Workflow

## Getting Started
1. Clone the repository
2. Install Go, Node.js, and dependencies
3. Build frontend assets (Vue.js)
4. Run backend (Go)
5. Access the web UI in your browser

## Hot Reload (Development)
- Use [Air](https://github.com/cosmtrek/air) for Go hot reload
- Use `npm run dev` or similar for frontend hot reload

## Code Structure
- Follow the directory conventions (see [Architecture](./architecture.md))
- Keep business logic in `web/service/`, API in `web/controller/`
- Use clear, descriptive commit messages

## Branching & PRs
- Use feature branches for new features
- Submit pull requests with clear descriptions
- Reference related issues in PRs

## Code Review
- Automated and manual review
- Ensure tests pass before merging

---

See [Testing](./testing.md) for test workflow.