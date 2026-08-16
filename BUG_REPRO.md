# 修复前故障复现（Docker）

## 项目与标准命令

Campus Room Booking Hub 是用于查询校园会议室、创建预约和查看占用情况的 Go HTTP 服务。在仓库根目录可执行以下标准命令：

```sh
go build ./...
go run ./cmd/server
go test ./...
```

修复前，`go test ./...` 会因房间目录请求失败而失败。

## 环境构建与编译

已在两个平台分别完成镜像构建和容器内编译：

```sh
docker build --platform linux/amd64 -f benzhi.Dockerfile -t campus-room-booking-hub-bug-002-base:amd64 .
docker run --rm --platform linux/amd64 campus-room-booking-hub-bug-002-base:amd64 go build ./...
docker build --platform linux/arm64 -f benzhi.Dockerfile -t campus-room-booking-hub-bug-002-base:arm64 .
docker run --rm --platform linux/arm64 campus-room-booking-hub-bug-002-base:arm64 go build ./...
```

两个平台的镜像构建和容器内 `go build ./...` 均成功。故障由下一节的房间目录验证触发。

## 故障触发步骤

先按上一节构建 `linux/amd64` 镜像，再从仓库根目录执行：

```sh
docker run --rm --platform linux/amd64 -e GOCACHE=/tmp/campus-room-booking-go-cache campus-room-booking-hub-bug-002-base:amd64 go test ./internal/transport -run TestHealthAndRooms -count=1
```

## 实际错误输出

```text
--- FAIL: TestHealthAndRooms (0.02s)
panic: runtime error: invalid memory address or nil pointer dereference [recovered, repanicked]
[signal SIGSEGV: segmentation violation code=0x1 addr=0x8 pc=0x5f14f3]

goroutine 21 [running]:
testing.tRunner.func1.2({0x675700, 0x901cb0})
	/usr/local/go/src/testing/testing.go:1974 +0x232
testing.tRunner.func1()
	/usr/local/go/src/testing/testing.go:1977 +0x349
panic({0x675700?, 0x901cb0?})
	/usr/local/go/src/runtime/panic.go:860 +0x13a
github.com/zhangchengcheng/campus-room-booking-hub/internal/store.cloneRoom(...)
	/app/internal/store/memory.go:164
github.com/zhangchengcheng/campus-room-booking-hub/internal/store.(*MemoryStore).ListRooms(0x12ae70e26f90, {0x6ded08?, 0x932920?})
	/app/internal/store/memory.go:49 +0x333
github.com/zhangchengcheng/campus-room-booking-hub/internal/service.(*Service).ListRooms(...)
	/app/internal/service/service.go:26
github.com/zhangchengcheng/campus-room-booking-hub/internal/transport.(*Handler).rooms(0x12ae70e040b0, {0x6dea60, 0x12ae70d980c0}, 0x12ae70d92280)
	/app/internal/transport/http.go:35 +0x77
net/http.HandlerFunc.ServeHTTP(0x12ae70e500c0?, {0x6dea60?, 0x12ae70d980c0?}, 0x12ae70d4ee58?)
	/usr/local/go/src/net/http/server.go:2284 +0x29
net/http.(*ServeMux).ServeHTTP(0x6ded08?, {0x6dea60, 0x12ae70d980c0}, 0x12ae70d92280)
	/usr/local/go/src/net/http/server.go:2826 +0x1c7
github.com/zhangchengcheng/campus-room-booking-hub/internal/transport_test.perform({0x6dd0e0, 0x12ae70e500c0}, {0x6c030a, 0x3}, {0x6c099c, 0x6}, {0x0, 0x0, 0x0})
	/app/internal/transport/http_test.go:28 +0x17c
github.com/zhangchengcheng/campus-room-booking-hub/internal/transport_test.TestHealthAndRooms(0x12ae70e64488)
	/app/internal/transport/http_test.go:40 +0xda
testing.tRunner(0x12ae70e64488, 0x6d8f90)
	/usr/local/go/src/testing/testing.go:2036 +0xea
created by testing.(*T).Run in goroutine 1
	/usr/local/go/src/testing/testing.go:2101 +0x4c5
FAIL	github.com/zhangchengcheng/campus-room-booking-hub/internal/transport	0.154s
FAIL
```

## 期望行为

房间目录请求应成功返回全部会议室；未分配负责人的会议室应以缺少负责人信息的状态正常显示，其他会议室仍可选择并继续预约。
