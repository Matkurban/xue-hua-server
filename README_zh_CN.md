# Server

## 本机编译

```shell
mage
```

## 编译特定平台
以下是编译为`linux_amd64`平台的命令，如果需要编译其他平台的替换此属性就可以
```shell
CGO_ENABLED=0 PLATFORMS=linux_amd64 mage
```

## 打包
把 Linux 平台的二进制打包成二进制，上传到服务器后就可以直接解压使用的
```shell
mage pack
```

## 运行容器

```shell
docker compose up -d
```

## 运行

```shell
mage start
```

## 二进制运行

```shell
./mage start
```

## 长期运行

```shell
nohup ./mage start >> _output/logs/server.log 2>&1 &
```

## 检查

```shell
mage check
```

## 二进制检查

```shell
./mage check
```

## 停止

```shell
mage stop
```

## 停止容器

```shell
docker compose down
```

## 二进制停止

```shell
./mage stop
```
