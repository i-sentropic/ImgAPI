package client

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/i-sentropic/imgAPI/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	address = "localhost:8089"
)

// func main() {
// 	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		log.Fatalf("did not connect: %v", err)
// 	}
// 	defer conn.Close()
// 	c := proto.NewImgAPIClient(conn)
// 	fileName := "badger-001.jpg"
// 	deets, header := Upload(c, fileName)

// 	fmt.Println(deets.FileId, header)

// 	fileID := "66e94a9b39115792a34b0c9a.gif"
// 	file, header := Download(c, fileID)
// 	fileName = "test"
// 	// open input file
// 	fi, err := os.Create(fileName)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// close fi on exit and check for its returned error
// 	defer func() {
// 		if err := fi.Close(); err != nil {
// 			panic(err)
// 		}
// 	}()

// 	fileType, _ := filetype.Match(file.ImageData)
// 	fmt.Println(fileType.Extension, header)

// 	fi.Write(file.ImageData)

// 	res := Delete(c, "66e94aa739115792a34b0c9c")
// 	fmt.Println(res)
// }

func Upload(client proto.ImgAPIClient, fileName string) (*proto.UploadResponse, metadata.MD) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fh, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()

	stat, err := fh.Stat()
	if err != nil {
		log.Fatal(err)
	}
	size := stat.Size()

	data, err := io.ReadAll(fh)
	if err != nil {
		log.Fatal(err)
	}

	var header metadata.MD

	resp, err := client.Upload(ctx, &proto.UploadRequest{
		Header: &proto.FileHeader{
			Name:     fileName,
			FileSize: size,
		},
		ImageData: data,
	},
		grpc.Header(&header),
	)
	if err != nil {
		log.Fatal(err)
	}

	return resp, header
}

func Download(client proto.ImgAPIClient, fileID string) (*proto.DownloadResponse, metadata.MD) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var header metadata.MD

	resp, err := client.Download(ctx, &proto.DownloadRequest{
		FileId: fileID,
	},
		grpc.Header(&header),
	)
	if err != nil {
		log.Fatal(err)
	}
	return resp, header
}

func Delete(client proto.ImgAPIClient, fileID string) *proto.DeleteResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Delete(ctx, &proto.DeleteRequest{FileId: fileID})
	if err != nil {
		log.Fatal(err)
	}
	return resp
}
