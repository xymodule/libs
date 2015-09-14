# libs
It's include the public common libs in gonet2

#  pb.go产生方式
protoc  ./*.proto --go_out=plugins=grpc:./

# etcd目录结构
etcd目录结构采用 http://gliderlabs.com/registrator/latest/ 提供的结构:

/backends/service_xxx/service_id  --->  ip:port

在services目录中执行init.sh，上传所有的有效名字到etcd
