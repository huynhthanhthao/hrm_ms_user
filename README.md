# hrm_user_service

# Change module name 
sed -i 's|"user/|"github.com/huynhthanhthao/hrm_user_service/|g' $(find . -name '*.go')

# generate proto
protoc --go_out=. --go-grpc_out=. proto/user.proto