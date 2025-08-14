# Local test

## Postgres

docker compose file:

```yaml
services:
  postgres:
    image: postgres:16
    container_name: postgres
    environment:
      POSTGRES_USER: appuser
      POSTGRES_PASSWORD: apppw
      POSTGRES_DB: appdb
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./initdb/fill.sql:/docker-entrypoint-initdb.d/fill.sql:ro
    restart: unless-stopped

volumes:
  postgres_data:
```

run/stop:

```shell
docker logs --tail=200 postgres
docker compose -f test/mysql/docker-compose.yml up -d
docker compose -f test/mysql/docker-compose.yml down -v
```

test connection:

```shell
psql postgres://appuser:apppw@localhost:5432/appdb
```