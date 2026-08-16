# 修复前故障复现（Docker）

## 项目与标准命令

校园房间预约服务提供房间查询、预约、签到和日占用统计接口。在仓库根目录可执行：

```bash
go build ./...
go run ./cmd/server
go test ./...
```

## 环境构建与编译

已实际执行以下命令。`linux/arm64` 和 `linux/amd64` 的镜像均构建成功，且各自容器内的 `go build ./...` 均成功。

```bash
docker build --platform linux/arm64 -f benzhi.Dockerfile -t campus-room-booking-hub:repro-arm64 .
docker run --rm --platform linux/arm64 campus-room-booking-hub:repro-arm64 go build ./...
docker build --platform linux/amd64 -f benzhi.Dockerfile -t campus-room-booking-hub:repro-amd64 .
docker run --rm --platform linux/amd64 campus-room-booking-hub:repro-amd64 go build ./...
```

## 故障触发步骤

在仓库根目录执行以下 Docker 命令，模拟扫描不存在的预约编号进行签到：

```bash
docker run --rm --platform linux/arm64 campus-room-booking-hub:repro-arm64 go test ./internal/transport -run TestCheckInUnknownBookingReturnsNotFound -count=1
```

## 实际错误输出

```text
--- FAIL: TestCheckInUnknownBookingReturnsNotFound (0.00s)
    http_test.go:168: unknown check-in status = 500
FAIL
FAIL	github.com/zhangchengcheng/campus-room-booking-hub/internal/transport	0.011s
FAIL
```

## 期望行为

扫描不存在的预约编号进行签到时，接口应返回“预约不存在”的客户端可识别结果，而不是 HTTP 500 服务器错误；工作人员无需将该情况误判为网络或服务故障并反复重试。
