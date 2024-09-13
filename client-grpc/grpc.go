package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/i-sentropic/imgAPI/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	address = "localhost:8089"
)

func main() {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewImgAPIClient(conn)
	fileName := "badger-001.jpg"
	deets, header := Upload(c, fileName)

	fmt.Println(deets.FileId, header)

	fileID := "66e46e52188265959abc16ab"
	file, header := Download(c, fileID)
	fileName = fileID + ".jpg"
	// open input file
	fi, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	fileData := file.ImageData
	fi.Write(fileData)

}

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

	log.Printf("[Upload] Will send file %q with size %d\n", fh.Name(), size)

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
