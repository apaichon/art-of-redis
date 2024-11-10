#!/bin/bash

start_app() {
  app_name=$1
  echo "Starting $app_name..."
  case $app_name in
    leaderboard)
      cd ../leaderboard/cmd/server && go run main.go --port 9002 &
      ;;
    sales-analytics)
      cd ../sales-analytics/cmd/server && go run main.go --port 9003 &
      ;;
    lucky-draw)
      cd ../lucky-draw/cmd/server && go run main.go --port 9004 &
      ;;
    frontend)
      frontend_path="$(pwd)/../../frontend"  # ใช้เส้นทางปัจจุบัน
      cd "$frontend_path" || exit 1  # ตรวจสอบการเปลี่ยนไดเรกทอรี
      npm install  # ติดตั้ง dependencies
      npm run dev -- --port 5000 &
      ;;
    session-management)
      session_management_path="../session-management/cmd/server"
      cd $session_management_path && go run main.go --port 9001 &
      ;;
    *)
      echo "Usage: $0 {leaderboard|sales-analytics|lucky-draw|frontend|session-management}"
      exit 1
      ;;
  esac
}

stop_app() {
  app_name=$1
  echo "Stopping $app_name..."
  case $app_name in
    leaderboard)
      pkill -f "leaderboard"
      ;;
    sales-analytics)
      pkill -f "sales-analytics.*9003"
      ;;
    lucky-draw)
      pkill -f "lucky-draw.*9004"
      ;;
    frontend)
      pkill -f "npm"
      ;;
    session-management)
      pkill -f "session-management.*9001"
      ;;
    *)
      echo "Usage: $0 {leaderboard|sales-analytics|lucky-draw|frontend|session-management}"
      exit 1
      ;;
  esac
}

case $1 in
  start)
    start_app $2
    ;;
  stop)
    stop_app $2
    ;;
  *)
    echo "Usage: $0 {start|stop} {leaderboard|sales-analytics|lucky-draw|frontend|session-management}"
    exit 1
    ;;
esac

