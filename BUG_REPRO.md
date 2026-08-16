# 修复前故障复现（Docker）

## 项目与标准命令

Campus Room Booking Hub 是一个用于浏览校园会议室、提交预约和查看占用情况的 Go HTTP 服务。

在仓库根目录执行：

```sh
go build ./...
go run ./cmd/server
go test ./...
```

## 环境构建与编译

```sh
docker build --platform linux/amd64 -f benzhi.Dockerfile -t campus-room-booking-hub:bug003-base-amd64 .
docker run --rm --platform linux/amd64 campus-room-booking-hub:bug003-base-amd64 go build ./...
docker build --platform linux/arm64 -f benzhi.Dockerfile -t campus-room-booking-hub:bug003-base-arm64 .
docker run --rm --platform linux/arm64 campus-room-booking-hub:bug003-base-arm64 go build ./...
```

两个平台的镜像构建和容器内编译应能完成；下述故障由房间目录查询结果被调用方修改后触发。

## 故障触发步骤

从仓库根目录执行：

```sh
go test ./internal/store -run TestListedRoomEquipmentDoesNotChangeCatalog -count=1
```

## 实际错误输出

```text
--- FAIL: TestListedRoomEquipmentDoesNotChangeCatalog (0.02s)
    memory_test.go:18: listed room equipment changed the catalog
FAIL
FAIL	github.com/zhangchengcheng/campus-room-booking-hub/internal/store	0.094s
FAIL
```

## 期望行为

调用方对一次房间目录查询结果中的设备清单进行临时标注或修改时，后续查询和预约校验仍应读取原有设备清单，不应在服务重启前持续出现未实际配置的设备。
