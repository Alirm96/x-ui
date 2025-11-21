# Error Handling

## Overview
x-ui provides consistent error handling across backend and frontend. Errors are reported to users with actionable messages, and detailed logs are kept for developers.

## Backend
- API responses use `{ success: false, msg: "...", obj: null }`
- Common error sources: invalid config, DB errors, Xray process failures, import validation
- Use descriptive, user-friendly messages (e.g., "Invalid trojan link: missing password")
- Log technical details in `logger/`

## Frontend
- Displays error messages from API in modals, alerts, or notifications
- Guides users to fix issues (e.g., import errors, config validation)
- Handles network and server errors gracefully

## Improving Error Messages
- Always specify what is wrong (e.g., which field, which link)
- Suggest corrective action if possible
- Log stack traces and details for developers

---

See [Frontend](./frontend.md) for UI error display patterns.