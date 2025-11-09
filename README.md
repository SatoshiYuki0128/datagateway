# datagateway
1. install docker desktop
2. execute docker-compose.yml

# 1) 安裝 protoc-gen-go 與 protoc-gen-go-grpc（如果尚未安裝）
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.29.0
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

# 2) 生成 pb.go
protoc --go_out=. --go-grpc_out=. proto/user.proto
