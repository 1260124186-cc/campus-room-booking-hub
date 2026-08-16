# 修复前故障复现（Docker）

## 项目与标准命令
Campus Room Booking Hub 是用于协调校园会议室查询、预约、签到和日占用统计的 Go HTTP 服务。在仓库根目录可执行以下标准命令：

```sh
go build ./...
go run ./cmd/server
go test ./...
```

## 环境构建与编译
已在 `linux/amd64` 和 `linux/arm64` 上分别实际执行以下命令，两个平台的镜像构建和容器内编译均成功：

```sh
docker build --platform linux/amd64 -f benzhi.Dockerfile -t campus-room-booking-hub-bug-001-base:amd64 .
docker run --rm --platform linux/amd64 campus-room-booking-hub-bug-001-base:amd64 go build ./...
docker build --platform linux/arm64 -f benzhi.Dockerfile -t campus-room-booking-hub-bug-001-base:arm64 .
docker run --rm --platform linux/arm64 campus-room-booking-hub-bug-001-base:arm64 go build ./...
```

目标故障在下节的并发预约测试中触发。

## 故障触发步骤
在仓库根目录执行：

```sh
GOCACHE=/private/tmp/campus-room-booking-go-cache-base go test ./internal/transport -run TestConcurrentReservationsHaveSingleWinner -count=1
```

## 实际错误输出
```text
--- FAIL: TestConcurrentReservationsHaveSingleWinner (0.05s)
    http_test.go:143: successful reservations = 12, want 1
FAIL
FAIL	github.com/zhangchengcheng/campus-room-booking-hub/internal/transport	0.587s
FAIL

[exit_code=1]
```

## 期望行为
多个协调人并发提交同一房间、同一天且时间重叠的预约时，只应有一个请求创建成功；其余请求应得到明确的冲突结果，日占用统计不应包含重复的重叠预约。
