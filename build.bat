protoc -I ./setuDumpProto picdump.proto --go_out=plugins=grpc:./picdump
go build