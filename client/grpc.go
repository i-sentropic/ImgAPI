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

func Upload(client proto.ImgAPIClient, fileName string) (*proto.UploadResponse, metadata.MD) {
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	resp, err := client.Upload(c, &proto.UploadRequest{
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


func Download(client proto.ImgAPIClient, fileID string, fileExtension string) (*proto.DownloadResponse, metadata.MD) {
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	if value, ok := contentType[fileExtension]; ok {
		c = metadata.AppendToOutgoingContext(c, "MIME-type", value)
	}

	defer cancel()

	var header metadata.MD

	resp, err := client.Download(c, &proto.DownloadRequest{
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
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Delete(c, &proto.DeleteRequest{FileId: fileID})
	if err != nil {
		log.Fatal(err)
	}
	return resp
}

var contentType = map[string]string{
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"bmp":  "image/bmp",
	"png":  "image/png",
	"tif":  "image/tiff",
	"tiff": "image/tiff",
	"gif":  "image/gif",
}
