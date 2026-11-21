# Server

Go instant-messaging server. Module: `server/v3`. Services and their configuration
use `server-*` names. Ordinary messages, calls, meetings and existing points-message
integration remain supported. Retired on-chain message types 130–132 are rejected;
this repository does not contain the external points ledger or client applications.

## Build and run

Requires Go **1.26.0+** and GNU `make` (or bash via `./scripts/manage.sh`).

```sh
make
# Optional targeted build:
make build-only TARGETS="server-api seq"
# Configure infrastructure and review config/*.yml before starting.
make start
make check
make stop

CGO_ENABLED=0 PLATFORMS=linux_amd64 make
make pack
# _output/server-linux-amd64.zip
./install.sh _output/server-linux-amd64.zip ./server-release
# On the Linux/amd64 deployment host, after configuring infrastructure:
cd server-release
make start   # or ./scripts/manage.sh start
make check   # or ./scripts/manage.sh check
make stop    # or ./scripts/manage.sh stop
```

The installer accepts a local release ZIP or extracted release directory and a
new destination. It does not download upstream sources, install system packages,
overwrite an existing installation, change credentials, or start services.
The Linux/amd64 release includes binaries, config, `Makefile`, and `scripts/manage.sh`;
it is **not** a source archive.
Use `make start`, `make stop` and `make check` (or `./scripts/manage.sh ...`) inside the extracted release:
the runtime scripts need neither Go nor remote Go dependencies on the deployment host.
`make export` and `make zip` are aliases for `make pack`.

## Containers

```sh
docker build -t server:local .
# One foreground service per container; configure reachable infrastructure first.
docker run --rm -v "$PWD/config:/server/config:ro" server:local \
  /server/bin/server-api -c /server/config
```

The image builds this checkout using Go 1.26.0, without upstream project downloads
or runtime Go dependency installation. See [deployment instructions](deployments/Readme.md)
for Kubernetes, and [build instructions](build/README.md) for image variants.
The image's default command starts only the API; run the other services separately.
Container `localhost` is not the host: configure dependency endpoints accordingly.

## Infrastructure and LiveKit

`docker-compose.yml` runs the infrastructure shared by this server and
[xue-hua-chat](../xue-hua-chat): MongoDB, Redis, etcd, Kafka, MinIO and the
LiveKit SFU used for calls and meetings.

```sh
docker compose up -d                                   # mongo, redis, etcd, kafka, minio, livekit
docker compose --profile livekit-extras up -d          # + livekit-egress (recording), livekit-ingress (RTMP/WHIP)
docker compose --profile caddy up -d                   # + caddy TLS front on :443 for livekit + TURN
```

LiveKit shares the `redis` container (`redis:6379`, db `2`; the IM server uses db `0`,
chat uses db `1`). Its configuration lives in `config/livekit.yaml`,
`config/livekit-egress.yaml`, `config/livekit-ingress.yaml` and `config/caddy.yaml`.
The API key/secret is supplied through `LIVEKIT_API_KEY` / `LIVEKIT_API_SECRET` in
`.env` and must equal `key` / `secret` in chat's `config/livekit.yml`.
Webhooks in `config/livekit.yaml` are `https://jsontodart.cn/chat/meeting/livekit_webhook`
and `https://jsontodart.cn/chat/conference/livekit_webhook` (the `/chat` prefix is
stripped by reverse proxy onto `chat-api-chat` on port `10008`). For TLS, set your domains in `config/caddy.yaml`,
run `deployments/livekit/update_caddy_ip.sh`, and enable `turn` in `config/livekit.yaml`.
Multi-node compose and Kubernetes manifests are under `deployments/livekit/`.

## Fresh deployment

The configuration targets a fresh `server_v3` MongoDB database, `server` object
bucket and `server` discovery root. This project has no separate wallet or on-chain
packet collections to initialize; ordinary message storage remains required.
Points balances and transfer business logic belong to an external service.

Service-specific overrides use `IMENV_SERVER_*`; common overrides such as
`IMENV_REDIS_PASSWORD` retain their names. Review credentials, admin accounts,
callback URLs and infrastructure addresses before the first deployment.
These configuration files do not migrate an existing installation.

## Attribution

Derived from [OpenIM](https://github.com/openimsdk/open-im-server).
Existing OpenIM copyright notices and third-party dependency identities are
preserved. Licensed under Apache-2.0; see [LICENSE](LICENSE).
