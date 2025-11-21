#!/bin/bash
# X-UI Development Helper Scripts

set -e

COMPOSE_FILE="docker-compose.dev.yml"
CONTAINER_NAME="x-ui-dev"
DB_PATH="./db/x-ui.db"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if container is running
is_running() {
    docker ps --filter "name=$CONTAINER_NAME" --format "{{.Names}}" | grep -q "$CONTAINER_NAME"
}

# Start development environment
dev_up() {
    log_info "Starting development environment..."
    
    # Create required directories
    mkdir -p db cert bin
    
    # Check if already running
    if is_running; then
        log_warning "Container already running"
        return 0
    fi
    
    docker compose -f "$COMPOSE_FILE" up -d
    
    log_success "Development environment started"
    log_info "Waiting for services to be ready..."
    sleep 5
    
    # Show logs
    docker logs --tail 20 "$CONTAINER_NAME"
    
    echo ""
    log_success "X-UI Dev is running at http://localhost:54321"
    log_info "Default credentials: admin / admin"
    log_info "Use 'make dev-logs' to view logs"
}

# Stop development environment
dev_down() {
    log_info "Stopping development environment..."
    docker compose -f "$COMPOSE_FILE" down
    log_success "Development environment stopped"
}

# Restart development environment
dev_restart() {
    log_info "Restarting development environment..."
    docker compose -f "$COMPOSE_FILE" restart
    log_success "Development environment restarted"
    sleep 3
    docker logs --tail 20 "$CONTAINER_NAME"
}

# View logs
dev_logs() {
    docker logs -f "$CONTAINER_NAME"
}

# Enter container shell
dev_shell() {
    if ! is_running; then
        log_error "Container is not running. Start it with 'make dev-up'"
        exit 1
    fi
    
    log_info "Entering container shell..."
    docker exec -it "$CONTAINER_NAME" sh
}

# Reset database
dev_db_reset() {
    log_warning "This will DELETE ALL DATA in the database!"
    read -p "Are you sure? (yes/no): " confirm
    
    if [ "$confirm" != "yes" ]; then
        log_info "Aborted"
        exit 0
    fi
    
    log_info "Stopping container..."
    docker compose -f "$COMPOSE_FILE" stop
    
    log_info "Removing database..."
    rm -f "$DB_PATH" "${DB_PATH}-shm" "${DB_PATH}-wal"
    
    log_info "Starting container..."
    docker compose -f "$COMPOSE_FILE" start
    
    log_success "Database reset complete"
}

# Backup database
dev_db_backup() {
    if [ ! -f "$DB_PATH" ]; then
        log_error "Database file not found: $DB_PATH"
        exit 1
    fi
    
    BACKUP_DIR="./db/backups"
    mkdir -p "$BACKUP_DIR"
    
    BACKUP_FILE="$BACKUP_DIR/x-ui-$(date +%Y%m%d-%H%M%S).db"
    
    log_info "Backing up database to $BACKUP_FILE..."
    cp "$DB_PATH" "$BACKUP_FILE"
    
    log_success "Database backed up successfully"
    log_info "Backup location: $BACKUP_FILE"
}

# Show status
dev_status() {
    echo ""
    log_info "Development Environment Status"
    echo "=================================="
    
    if is_running; then
        log_success "Container: Running"
        
        # Get container info
        UPTIME=$(docker inspect -f '{{.State.StartedAt}}' "$CONTAINER_NAME" | xargs date +%s -d)
        CURRENT=$(date +%s)
        DIFF=$((CURRENT - UPTIME))
        MINS=$((DIFF / 60))
        
        echo "  Uptime: ${MINS} minutes"
        echo "  URL: http://localhost:54321"
        
        # Check if xray exists
        if docker exec "$CONTAINER_NAME" test -f /app/bin/xray 2>/dev/null; then
            log_success "Xray: Installed"
        else
            log_warning "Xray: Not installed"
        fi
        
        # Check database
        if [ -f "$DB_PATH" ]; then
            SIZE=$(du -h "$DB_PATH" | cut -f1)
            log_success "Database: $SIZE"
        else
            log_warning "Database: Not created yet"
        fi
        
    else
        log_error "Container: Not running"
        log_info "Start with: make dev-up"
    fi
    
    echo ""
}

# Install xray binary
dev_install_xray() {
    if ! is_running; then
        log_error "Container is not running. Start it with 'make dev-up'"
        exit 1
    fi
    
    log_info "Installing xray binary..."
    
    docker exec "$CONTAINER_NAME" sh -c '
        cd /app/bin
        apt-get update -qq && apt-get install -y -qq wget unzip
        wget -q https://github.com/XTLS/Xray-core/releases/download/v25.9.11/Xray-linux-64.zip
        unzip -q Xray-linux-64.zip
        rm Xray-linux-64.zip
        chmod +x xray
        # Create symlinks for all supported architectures
        ln -sf xray xray-linux-amd64
        ln -sf xray xray-linux-arm64
        ln -sf xray xray-linux-arm
        ln -sf xray xray-linux-i386
    '
    
    log_success "Xray installed successfully"
}

# Clean build cache
dev_clean() {
    log_warning "This will remove build cache and temp files"
    read -p "Continue? (yes/no): " confirm
    
    if [ "$confirm" != "yes" ]; then
        log_info "Aborted"
        exit 0
    fi
    
    log_info "Cleaning build cache..."
    
    # Stop container
    docker compose -f "$COMPOSE_FILE" down
    
    # Remove volumes
    docker volume rm x-ui-dev-go-mod-cache x-ui-dev-air-tmp 2>/dev/null || true
    
    # Remove tmp directory
    rm -rf tmp/
    
    log_success "Build cache cleaned"
    log_info "Restart with: make dev-up"
}

# Show help
show_help() {
    cat << EOF
X-UI Development Helper Script

Usage: ./dev.sh [command]

Commands:
  up          Start development environment
  down        Stop development environment
  restart     Restart development environment
  logs        View container logs
  shell       Enter container shell
  status      Show environment status
  
  db-reset    Reset database (deletes all data!)
  db-backup   Backup database
  
  install-xray  Install xray binary in container
  clean         Clean build cache and temp files
  
  help        Show this help message

Examples:
  ./dev.sh up
  ./dev.sh logs
  ./dev.sh shell
  ./dev.sh db-backup

For more information, see DEVELOPMENT.md
EOF
}

# Main command dispatcher
case "${1:-help}" in
    up)
        dev_up
        ;;
    down)
        dev_down
        ;;
    restart)
        dev_restart
        ;;
    logs)
        dev_logs
        ;;
    shell)
        dev_shell
        ;;
    status)
        dev_status
        ;;
    db-reset)
        dev_db_reset
        ;;
    db-backup)
        dev_db_backup
        ;;
    install-xray)
        dev_install_xray
        ;;
    clean)
        dev_clean
        ;;
    help|*)
        show_help
        ;;
esac
