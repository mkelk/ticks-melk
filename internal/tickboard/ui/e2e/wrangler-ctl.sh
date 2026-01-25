#!/bin/bash
#
# Wrangler dev orchestration for integration testing
#
# Commands:
#   start  - Start wrangler dev in background, wait for ready, print URL
#   stop   - Stop wrangler dev
#   status - Check if wrangler dev is running
#
# Environment variables:
#   WRANGLER_PORT - Port to use (default: 8787, will try alternatives if busy)
#   WRANGLER_TIMEOUT - Timeout in seconds waiting for ready (default: 30)
#
# Usage:
#   ./wrangler-ctl.sh start   # Starts and prints URL
#   ./wrangler-ctl.sh stop    # Stops wrangler dev
#   ./wrangler-ctl.sh status  # Prints running status
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKER_DIR="$(cd "${SCRIPT_DIR}/../../../../cloud/worker" && pwd)"
PIDFILE="${SCRIPT_DIR}/.wrangler.pid"
PORTFILE="${SCRIPT_DIR}/.wrangler.port"
LOGFILE="${SCRIPT_DIR}/.wrangler.log"

# Default settings
DEFAULT_PORT=${WRANGLER_PORT:-8787}
TIMEOUT=${WRANGLER_TIMEOUT:-30}

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
  echo -e "${GREEN}[wrangler-ctl]${NC} $1" >&2
}

log_warn() {
  echo -e "${YELLOW}[wrangler-ctl]${NC} $1" >&2
}

log_error() {
  echo -e "${RED}[wrangler-ctl]${NC} $1" >&2
}

# Check if a port is in use
port_in_use() {
  local port=$1
  lsof -i ":${port}" >/dev/null 2>&1
}

# Find an available port starting from the given port
find_available_port() {
  local port=$1
  local max_attempts=10
  local attempt=0

  while [ $attempt -lt $max_attempts ]; do
    if ! port_in_use $port; then
      echo $port
      return 0
    fi
    log_warn "Port $port is in use, trying next..."
    port=$((port + 1))
    attempt=$((attempt + 1))
  done

  log_error "Could not find available port after $max_attempts attempts"
  return 1
}

# Wait for wrangler to be ready
wait_for_ready() {
  local port=$1
  local timeout=$2
  local url="http://localhost:${port}/health"
  local elapsed=0
  local interval=1

  log_info "Waiting for wrangler dev to be ready at ${url}..."

  while [ $elapsed -lt $timeout ]; do
    if curl -s -f "${url}" >/dev/null 2>&1; then
      log_info "Wrangler dev is ready!"
      return 0
    fi
    sleep $interval
    elapsed=$((elapsed + interval))
  done

  log_error "Timeout waiting for wrangler dev to be ready after ${timeout}s"
  return 1
}

# Get the PID of the running wrangler process
get_pid() {
  if [ -f "$PIDFILE" ]; then
    local pid=$(cat "$PIDFILE")
    if ps -p "$pid" >/dev/null 2>&1; then
      echo "$pid"
      return 0
    fi
  fi
  return 1
}

# Get the port of the running wrangler process
get_port() {
  if [ -f "$PORTFILE" ]; then
    cat "$PORTFILE"
    return 0
  fi
  return 1
}

# Start wrangler dev
cmd_start() {
  # Check if already running
  if pid=$(get_pid); then
    local port=$(get_port)
    log_warn "Wrangler dev already running (PID: $pid, port: $port)"
    echo "http://localhost:${port}"
    return 0
  fi

  # Find available port
  local port=$(find_available_port $DEFAULT_PORT)
  if [ -z "$port" ]; then
    return 1
  fi

  log_info "Starting wrangler dev on port ${port}..."

  # Start wrangler dev in background
  cd "$WORKER_DIR"
  nohup npx wrangler dev --port "$port" > "$LOGFILE" 2>&1 &
  local pid=$!

  # Save PID and port
  echo "$pid" > "$PIDFILE"
  echo "$port" > "$PORTFILE"

  log_info "Started wrangler dev (PID: $pid)"

  # Wait for ready
  if ! wait_for_ready "$port" "$TIMEOUT"; then
    log_error "Failed to start wrangler dev. Check logs at: $LOGFILE"
    cmd_stop
    return 1
  fi

  # Print URL
  local url="http://localhost:${port}"
  log_info "Wrangler dev running at: ${url}"
  echo "$url"
}

# Stop wrangler dev
cmd_stop() {
  if pid=$(get_pid); then
    log_info "Stopping wrangler dev (PID: $pid)..."

    # Kill the process and its children (wrangler spawns child processes)
    pkill -P "$pid" 2>/dev/null || true
    kill "$pid" 2>/dev/null || true

    # Wait for process to exit
    local timeout=10
    local elapsed=0
    while ps -p "$pid" >/dev/null 2>&1 && [ $elapsed -lt $timeout ]; do
      sleep 1
      elapsed=$((elapsed + 1))
    done

    # Force kill if still running
    if ps -p "$pid" >/dev/null 2>&1; then
      log_warn "Process didn't exit gracefully, force killing..."
      kill -9 "$pid" 2>/dev/null || true
    fi

    log_info "Wrangler dev stopped"
  else
    log_info "Wrangler dev is not running"
  fi

  # Clean up files
  rm -f "$PIDFILE" "$PORTFILE"
}

# Check status
cmd_status() {
  if pid=$(get_pid); then
    local port=$(get_port)
    local url="http://localhost:${port}"

    # Check health endpoint
    if curl -s -f "${url}/health" >/dev/null 2>&1; then
      log_info "Wrangler dev is running and healthy (PID: $pid, port: $port)"
      echo "$url"
      return 0
    else
      log_warn "Wrangler dev process exists but not responding (PID: $pid)"
      return 1
    fi
  else
    log_info "Wrangler dev is not running"
    return 1
  fi
}

# Show usage
usage() {
  echo "Usage: $0 {start|stop|status}"
  echo ""
  echo "Commands:"
  echo "  start   Start wrangler dev, wait for ready, print URL"
  echo "  stop    Stop wrangler dev"
  echo "  status  Check if wrangler dev is running"
  echo ""
  echo "Environment variables:"
  echo "  WRANGLER_PORT    Port to use (default: 8787)"
  echo "  WRANGLER_TIMEOUT Timeout in seconds (default: 30)"
}

# Main
case "${1:-}" in
  start)
    cmd_start
    ;;
  stop)
    cmd_stop
    ;;
  status)
    cmd_status
    ;;
  *)
    usage
    exit 1
    ;;
esac
