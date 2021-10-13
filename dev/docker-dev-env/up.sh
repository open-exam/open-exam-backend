#!/usr/bin/env bash
docker-compose up -d

echo "yes" | redis-cli --cluster create $(for ind in `seq 1 6`; do \
                   echo -n "$(docker inspect -f \
                   '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' \
                   docker-dev-env_redis-$ind\_1)"':6379 '; \
                   done) --cluster-replicas 1 -a test

for ind in $(seq 1 6); do
  echo -n docker-dev-env_redis-$ind\_1 " "
  docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' docker-dev-env_redis-$ind\_1;
done

sleep 1s

docker exec -it $(docker container ls  | grep 'docker-dev-env_db_1' | awk '{print $1}') sh -c 'mysql -u open_exam -popen_exam open_exam < /up.sql'

echo -n docker-dev-env_db_1 " "
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' docker-dev-env_db_1;