protoc -I ./setuDumpProto picdump.proto --go_out=plugins=grpc:./transmit
protoc -I ./setuDumpProto msgforward.proto --go_out=plugins=grpc:./transmit
go build