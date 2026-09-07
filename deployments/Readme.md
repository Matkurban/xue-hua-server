# Server Kubernetes deployment

These are **single-instance development templates**, not a highly available
production database topology. They use the `server` namespace and existing
third-party MongoDB, Redis, Kafka and MinIO images. Review image availability,
storage class, credentials, external object-storage URL, callbacks and ingress
before deployment. These templates target a [fresh deployment](../README.md#fresh-deployment),
not a migration of existing storage. Existing example credentials are preserved,
not rotated by this change.

## Build and prepare (from repository root)

```sh
docker build -t server:local .
go run ./deployments/render
go run ./deployments/render -check
```

Load this **newly built local image** into every node of your development cluster
(for example, `kind load docker-image server:local` for an existing kind cluster).
No upstream service image provides these renamed binaries. For a separately
managed registry image, use `go run ./deployments/render -image YOUR_BUILT_IMAGE`;
the renderer never builds, downloads or pushes images.

The renderer copies current runtime configuration from `config/` and generates
`deploy/server-config.yml` plus application Deployments and Services. It changes
only cluster discovery, dependency addresses, injected credential placeholders
and fixed listening/metrics ports. Re-run it after editing `config/`.
It preserves notification, webhook, rate-limiter and object-storage settings;
review externally reachable URLs yourself. A ConfigMap may contain other sensitive
application configuration: do not expose the generated file publicly.

Each application runs `/server/bin/server-* -c /config` as its foreground process.
All pods mount the complete ConfigMap and set `CONFIG_PATH=/config`.
The current runtime detects Kubernetes via the mounted service-account directory,
so `automountServiceAccountToken` must remain enabled. The namespace-scoped
`server` account can read Services, Endpoints and Pods for discovery. RPC ports
must be named exactly `grpc` in both Services and containers.
Common `IMENV_*` credential overrides refer to the matching Secret keys.

## Apply only after review

```sh
kubectl apply -f deployments/deploy/clusterRole.yml
kubectl -n server apply -f deployments/deploy/
kubectl -n server get pods,services,pvc
```

The optional ingress assumes an installed ingress-nginx controller with class
`nginx`. `/server-api/...` preserves the API suffix and
`/server-msggateway/...` routes WebSockets. Configure host/TLS and external URLs
for your environment. Infrastructure Services and PVCs keep their generic names;
the namespace isolates them. These storage manifests use one writer per volume,
not replicas sharing one data directory.

Changing Secret values does not rotate existing database users. MongoDB initializes
the application user from `server-mongo-secret` for a fresh `server_v3` database;
existing users retain their credentials and must be migrated explicitly.
Do not rely on initialization to copy old messages or object data.

Image builds, server startup, live Kubernetes rollout and data migration require
working infrastructure. Static manifest checks alone do not verify those outcomes.
