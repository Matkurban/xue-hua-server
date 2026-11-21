# Server

## 本机编译

```shell
make
```

## 编译指定目标

```shell
make build-only TARGETS="server-api seq"
```

## 编译特定平台

以下是编译为 `linux_amd64` 平台的命令，如果需要编译其他平台替换 `PLATFORMS` 即可：

```shell
CGO_ENABLED=0 PLATFORMS=linux_amd64 make
```

## 打包

把 Linux 平台的二进制和运行脚本打包成压缩包，上传到服务器后直接解压使用（无需安装 Go）：

```shell
make pack
```

## 运行容器

```shell
docker compose up -d
```

## 运行

```shell
make start
```

或直接运行脚本（无 make 环境时）：

```shell
./scripts/manage.sh start
```

## 检查

```shell
make check
```

## 停止

```shell
make stop
```

## 停止容器

```shell
docker compose down
```

## 单元测试

```shell
make test
```
