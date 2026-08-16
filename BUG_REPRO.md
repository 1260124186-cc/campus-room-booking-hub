# 修复前故障复现（Docker）

## 项目与标准命令

Campus Room Booking Hub 是用于浏览校园会议室、提交预约、签到和查询每日占用情况的 Go HTTP 服务。在仓库根目录可执行以下标准命令：

```sh
go build ./...
go test ./...
go run ./cmd/server
```

## 环境构建与编译

修复前状态在两个目标平台均能完成镜像构建和容器内编译：

```sh
docker build --platform linux/amd64 -f benzhi.Dockerfile -t campus-room-booking-hub:repro-amd64 .
docker run --rm --platform linux/amd64 campus-room-booking-hub:repro-amd64 go build ./...
docker build --platform linux/arm64 -f benzhi.Dockerfile -t campus-room-booking-hub:repro-arm64 .
docker run --rm --platform linux/arm64 campus-room-booking-hub:repro-arm64 go build ./...
```

`linux/amd64` 与 `linux/arm64` 的镜像构建和容器内 `go build ./...` 均成功。

## 故障触发步骤

在仓库根目录先构建上面的 `linux/arm64` 镜像，再执行：

```sh
docker run --rm --platform linux/arm64 campus-room-booking-hub:repro-arm64 go test ./internal/transport -run TestCanceledReservationDoesNotPersist -count=1
```

## 实际错误输出

```text
--- FAIL: TestCanceledReservationDoesNotPersist (0.00s)
    http_test.go:157: expected cancellation error
FAIL
FAIL	github.com/zhangchengcheng/campus-room-booking-hub/internal/transport	0.003s
FAIL
exit_status=1
```

## 期望行为

取消预约请求后，调用方应收到取消结果，并且当天占用统计中不应出现该预约时段；后续用户仍应能够预约该时段。
