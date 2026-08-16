#!/usr/bin/env bash
set -euo pipefail

project_image="${PROJECT_IMAGE:-campus-room-booking-hub:benzhi}"
platform="${1:-linux/amd64}"
container_name="campus-room-booking-hub-benzhi-${platform##*/}"

docker build --platform "$platform" -f benzhi.Dockerfile -t "$project_image-$platform" .
docker rm -f "$container_name" >/dev/null 2>&1 || true
docker run -d --rm --name "$container_name" --platform "$platform" -p 18080:8080 "$project_image-$platform" >/dev/null
trap 'docker rm -f "$container_name" >/dev/null 2>&1 || true' EXIT

for _ in $(seq 1 30); do
  if curl -fsS http://127.0.0.1:18080/healthz >/dev/null; then
    break
  fi
  sleep 1
done

docker exec "$container_name" go build ./...
curl -fsS http://127.0.0.1:18080/healthz
curl -fsS http://127.0.0.1:18080/rooms >/dev/null
