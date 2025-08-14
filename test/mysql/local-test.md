# Local test

## Mysql

docker compose file:

```yaml
services:
  db:
    image: mysql:8.4
    container_name: mysql
    restart: unless-stopped
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: rootpw
      MYSQL_DATABASE: appdb
      MYSQL_USER: appuser
      MYSQL_PASSWORD: apppw
      TZ: "America/New_York"
    # optional but fine to keep:
    command: >
      --character-set-server=utf8mb4
      --collation-server=utf8mb4_0900_ai_ci
    healthcheck:
      test: ["CMD-SHELL", "mysqladmin ping -h 127.0.0.1 -prootpw || exit 1"]
      interval: 10s
      timeout: 3s
      retries: 5
    volumes:
      - mysql_data:/var/lib/mysql
      - ./initdb:/docker-entrypoint-initdb.d:ro
volumes:
  mysql_data:
```

run/stop:

```shell
docker rm -f mysql
docker logs --tail=200 mysql
docker volume ls
docker volume rm rel8tmp_mysql_data  # wipe old DB so init script runs
docker compose -f test/mysql/docker-compose.yml up -d

docker compose -f test/mysql/docker-compose.yml down -v
```

test connection:

```shell
mysql -h 127.0.0.1 -P 3306 -u appuser -papppw appdb
```