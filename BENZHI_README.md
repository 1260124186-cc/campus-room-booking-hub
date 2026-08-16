# Docker 验收说明

Campus Room Booking Hub 是一个本地运行的校园会议室预约 HTTP 服务，支持房间目录、预约、签到和按日占用统计。服务使用内存保存运行期间的数据，不依赖外部数据库或在线服务。

## 本地标准命令

从仓库根目录执行：

```sh
go build ./...
go run ./cmd/server
go test ./...
```

服务默认监听容器端口 `8080`，健康检查为 `GET /healthz`，房间目录接口为 `GET /rooms`。

## 工具链固定

实际 Dockerfile 为 `benzhi.Dockerfile`。`go.mod` 固定 Go 语言版本为 `1.26`，Dockerfile 的构建和运行阶段都设置 `GOTOOLCHAIN=local`，镜像在容器内从源码执行 `go mod download` 与 `go build ./...`，不复制宿主机二进制。

## 双架构脚本

脚本 `build_benzhi_docker.sh` 的第一个参数是 Docker 平台，脚本依次构建镜像、启动容器、等待健康接口、在容器内执行 `go build ./...`，再请求健康接口和真实房间目录接口：

```sh
./build_benzhi_docker.sh linux/amd64
./build_benzhi_docker.sh linux/arm64
```

## 手工验收

```sh
docker build --platform linux/amd64 -f benzhi.Dockerfile -t campus-room-booking-hub:amd64 .
docker run -d --rm --name campus-room-booking-hub-amd64 -p 18080:8080 campus-room-booking-hub:amd64
docker exec campus-room-booking-hub-amd64 go build ./...
curl -fsS http://127.0.0.1:18080/healthz
curl -fsS http://127.0.0.1:18080/rooms
docker rm -f campus-room-booking-hub-amd64
```

将上述命令中的 `linux/amd64`、镜像标签和容器名替换为 `linux/arm64` 对应值即可验收另一平台。通过标准是镜像构建、容器启动、容器内 `go build ./...` 和接口请求均返回退出码 `0`；健康接口返回 HTTP `200` 且 JSON 状态为 `ok`，房间目录接口返回 HTTP `200`。
