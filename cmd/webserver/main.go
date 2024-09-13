package main

import (
	"fmt"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/i-sentropic/imgAPI/pkg/proto"
	"github.com/i-sentropic/imgAPI/pkg/routes"
	"google.golang.org/grpc"
)

func main() {
	server := gin.Default()

	routes.RegisterRoutes(server)

	go server.Run(":8080")

	lis, err := net.Listen("tcp", ":8089")
	if err != nil {
		fmt.Print(err)
	}

	ImgAPIServer := &proto.ImgAPI{}

	grpcServer := grpc.NewServer()
	proto.RegisterImgAPIServer(grpcServer, ImgAPIServer)
	err = grpcServer.Serve(lis)
	if err != nil {
		fmt.Print("impossible to serve", err)
	}

}
