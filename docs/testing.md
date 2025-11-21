# Testing

## Overview
Testing is essential for maintaining code quality and reliability. x-ui uses a combination of unit, integration, and manual tests.

## Backend (Go)
- Unit tests for services and controllers
- Use Go's built-in `testing` package
- Place tests in the same package with `_test.go` suffix
- Run tests: `go test ./...`

## Frontend (Vue.js)
- Unit/component tests (if set up)
- Manual testing via browser
- Linting for code quality

## End-to-End
- Manual E2E tests for UI/API integration
- Use Docker for environment parity

## Adding Tests
- Write clear, isolated test cases
- Mock dependencies where possible
- Cover edge cases and error handling

---

See [Deployment](./deployment.md) for release process.