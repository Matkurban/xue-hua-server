# Building Server

Run from the repository root with Go 1.26.0+ and Mage:

```sh
mage
mage buildOnly "server-api seq"
CGO_ENABLED=0 PLATFORMS=linux_amd64 mage
mage pack
```

Linux services are written to `_output/bin/platforms/linux/amd64/server-*`;
tools are under `_output/bin/tools/linux/amd64/`.
The portable runtime archive is `_output/server-linux-amd64.zip`.
It includes a compiled Linux/amd64 `mage` launcher, Mage startup files, config
and module metadata, not the full source. Use `./mage start`, `./mage stop` and
`./mage check` in the extracted release without installing Go or downloading
Go dependencies on the deployment host. `mage export` is another packaging command.
See the root README for fresh-deployment configuration.

## Local images

```sh
docker build -t server:local .
# Optional service-specific default command, using the image just built:
docker build -f build/images/server-rpc-user/Dockerfile -t server-rpc-user:local .
```

All service recipes derive from the locally built `server:local` image.
Never substitute an upstream service image: its binaries and config names differ.
Kubernetes uses the unified image with an explicit foreground command per service.
The root Dockerfile builds current source with Go 1.26.0 and ships only static
service binaries, config, CA certificates, timezone data and LICENSE.
Override build version with `--build-arg VERSION=your-version`; otherwise it is `dev`.
No Docker daemon, push or live deployment is required for source-level checks.

Third-party base images, dependency paths and original OpenIM license attribution
remain unchanged in identity. See the root LICENSE.
