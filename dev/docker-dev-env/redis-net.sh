for ind in $(seq 1 6); do
  echo -n docker-dev-env_redis-$ind\_1 " "
  docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' docker-dev-env_redis-$ind\_1;
done

docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' docker-dev-env_redis-insight_1;