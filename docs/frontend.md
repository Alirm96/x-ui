# Frontend (Vue.js)

## Overview
The frontend is a Single Page Application (SPA) built with Vue.js. It provides a modern, responsive UI for managing Xray-core and related resources.

### Main Responsibilities
- User interface for managing outbounds, inbounds, users, settings
- Import/export of configuration links
- Real-time status and logs
- Error and notification display

## Key Directories & Files
- `web/html/xui/`: Vue.js templates (e.g., outbounds.html)
- `web/assets/js/`: JavaScript (Vue components, utils)
- `web/assets/css/`: Stylesheets

## UI Structure
- **Pages:** Outbounds, Inbounds, Users, Settings, Logs
- **Components:** Modals, tables, forms, notifications
- **State Management:** Local state, API-driven

## API Communication
- Uses `HttpUtil` (in `utils.js`) for AJAX requests
- Handles API responses, error messages, and updates UI accordingly

---

See [Error Handling](./error_handling.md) for user feedback patterns.