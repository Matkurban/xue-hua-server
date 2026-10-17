# Docker publish workflows (ported from open-im-server v3.8.3-patch.18)

## What was added
- `publish-docker-image.yml` — multi-arch build of the repo root `Dockerfile` as `xue-hua-server`, with **separate primary-registry and ACR pushes** (ACR filters `*.aliyuncs.com` tags; `BUILDX_NO_DEFAULT_ATTESTATIONS=1` for ACR).
- `docker-build-and-release-services-images.yml` — per-service matrix via `build/images/server/docker-compose.build.yml` + `build/build_single.sh`, same ACR split as upstream #3815.
- `build/build_single.sh` — adapted from upstream (compose path → `build/images/server/`).
- `build/images/server/Dockerfile` + `docker-compose.build.yml` — builds each `server-*` binary independently.

## Secrets required
- `DOCKER_USERNAME` / `DOCKER_PASSWORD`
- `GITHUB_TOKEN` (automatic for GHCR)
- `ALIREGISTRY_USERNAME` / `ALIREGISTRY_TOKEN` for `registry.cn-hangzhou.aliyuncs.com/xuehua`

## Notes
- Image namespace uses `xuehua` on ACR (not `openimsdk`).
- OpenIM compose smoke-test / `openim-docker` checkout was **not** ported; this fork uses its own compose stack.
- Per-service Dockerfiles under `build/images/server-*` that only `FROM server:local` are unchanged; CI uses the new `build/images/server/Dockerfile`.
