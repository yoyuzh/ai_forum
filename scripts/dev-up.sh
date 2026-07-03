#!/usr/bin/env sh
set -eu

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$ROOT"

wait_for() {
  name="$1"
  shift
  i=0
  until "$@" >/dev/null 2>&1; do
    i=$((i + 1))
    if [ "$i" -gt 90 ]; then
      echo "timed out waiting for $name" >&2
      return 1
    fi
    sleep 2
  done
}

docker compose up -d mysql redis rabbitmq elasticsearch

wait_for mysql docker compose exec -T mysql mysqladmin ping -h 127.0.0.1 -uroot -pai_forum_root
wait_for redis docker compose exec -T redis redis-cli ping
wait_for rabbitmq docker compose exec -T rabbitmq rabbitmq-diagnostics -q ping
wait_for elasticsearch curl -fsS http://127.0.0.1:9200/_cluster/health

MYSQL_HOST=127.0.0.1 \
MYSQL_PORT=3306 \
MYSQL_USERNAME=root \
MYSQL_PASSWORD=ai_forum_root \
MYSQL_DATABASE=ai_forum \
  make migrate-up

docker compose up -d api-server worker-service outbox-publisher
docker compose up -d --force-recreate nginx

wait_for readyz curl -fsS http://127.0.0.1:19091/readyz

docker compose up -d web admin

wait_for web curl -fsS http://127.0.0.1:5173/posts
wait_for admin curl -fsS http://127.0.0.1:5174/login

echo "ai_forum dev stack is ready"
