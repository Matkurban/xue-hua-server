# Server

Go 即时通信服务端，模块为 `server/v3`，服务程序和配置统一使用 `server-*`。
保留普通聊天、通话、会议及现有积分消息接入；拒绝已退役的链上消息类型
130–132。本仓库不包含外部积分账本及客户端。

## 编译与运行

需要 Go **1.26.0+** 和 [Mage](https://magefile.org/)。

```sh
mage
# 可选：仅编译指定目标
mage buildOnly "server-api seq"
# 先配置基础设施并检查 config/*.yml
mage start
mage check
mage stop

CGO_ENABLED=0 PLATFORMS=linux_amd64 mage
mage pack
./install.sh _output/server-linux-amd64.zip ./server-release
# 在 Linux/amd64 部署主机完成基础设施配置后：
cd server-release
./mage start
nohup ./mage start >> _output/logs/server.log 2>&1 &
./mage check
./mage stop
```

安装脚本仅接受本地发布 ZIP 或解压目录，以及尚不存在的目标目录。
不会下载上游项目、安装系统软件、覆盖旧安装、修改凭据或自动启动服务。
发布包面向 Linux/amd64，包含二进制、配置、已编译的 `mage` 启动器、
Mage 运行文件及模块元数据，**不是源码包**。
在解压目录使用 `./mage start/stop/check`，部署主机无需 Go，也无需下载 Go 依赖。
源码目录使用已安装的 Mage；`mage export` 同样用于打包发布。

## 容器

```sh
docker build -t server:local .
docker run --rm -v "$PWD/config:/server/config:ro" server:local \
  /server/bin/server-api -c /server/config
```

镜像使用 Go 1.26.0 从当前项目构建，不下载上游服务项目，不在运行时安装 Go 依赖。
默认仅以前台方式启动 API，其余服务需分别启动；容器的 localhost 不是宿主机，
必须配置可达的基础设施地址。
参见 [Kubernetes 部署](deployments/Readme.md) 和 [构建说明](build/README.md)。

## 首次部署

配置使用全新的 MongoDB 数据库 `server_v3`、对象存储桶 `server` 和服务发现根目录
`server`。本项目没有独立的钱包或链上红包集合初始化代码，普通聊天存储继续保留。
积分账本和转账业务逻辑由外部业务服务负责。

服务专属环境变量使用 `IMENV_SERVER_*`，公共变量如 `IMENV_REDIS_PASSWORD` 不变。
首次部署前需检查凭据、管理员账号、回调地址和基础设施连接配置。
这些配置用于新部署，不会迁移已有环境的数据。

## 版权与许可

基于 [OpenIM](https://github.com/openimsdk/open-im-server)，保留原有版权声明及
真实第三方依赖标识。采用 Apache-2.0，参见 [LICENSE](LICENSE)。
