#!/usr/bin/env bash
docker-compose up -d
echo "yes" | redis-cli --cluster create $(for ind in `seq 1 2`; do \
                   echo -n "$(docker inspect -f \
                   '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' \
                   redis_redis-$ind\_1)"':6379 '; \
                   done) --cluster-replicas 1 -a test
for ind in $(seq 1 6); do
  docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' redis_redis-$ind\_1;
done