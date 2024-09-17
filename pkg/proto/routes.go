package proto

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/h2non/filetype"
	"github.com/i-sentropic/imgAPI/pkg/db"
	"github.com/i-sentropic/imgAPI/pkg/src"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ImgAPI struct {
	UnimplementedImgAPIServer
}

func (s *ImgAPI) Upload(c context.Context, req *UploadRequest) (*UploadResponse, error) {
	header := *req.GetHeader()
	originalFileName := header.GetName()

	originalFileName = src.CleanFileName(originalFileName)
	buf := bytes.NewBuffer(nil)

	_, err := buf.Write(req.GetImageData())
	if err != nil {
		return &UploadResponse{}, errors.New("nil value byte slice")
	}

	//Generate meta data and file name
	bodyBytes := buf.Bytes()
	fileType, _ := filetype.Match(bodyBytes)
	fileExt := fileType.Extension
	fileName := time.Now().Format(time.RFC3339) + "_" + originalFileName
	transformHistory := []string{}
	fileMetaData := &src.MetaDataFields{originalFileName, transformHistory, fileExt}

	response, err := src.SaveToDB(&bodyBytes, fileMetaData, fileName)

	if err != nil {
		header := metadata.Pairs()
		grpc.SetHeader(c, header)
		return &UploadResponse{}, err
	}

	fileSize := strconv.Itoa(response["fileSize"].(int))

	headerResponse := metadata.Pairs(
		"fileId", response["fileId"].(string),
		"fileSize", fileSize,
		"fileExtension", response["fileExtension"].(string),
		"originalFileName", response["originalFileName"].(string),
		"transformHistory", strings.Join(response["transformHistory"].([]string), " | "),
	)
	grpc.SetHeader(c, headerResponse)

	resp := UploadResponse{FileId: response["fileId"].(string)}

	return &resp, nil
}

func (s *ImgAPI) Download(c context.Context, req *DownloadRequest) (*DownloadResponse, error) {
	fileID := req.FileId
	buf, foundFile, err := src.GetFileFromDB(fileID)
	if err != nil {
		return &DownloadResponse{}, err
	}

	md, ok := metadata.FromIncomingContext(c)
	if ok {
		key := strings.Join(md.Get("mime-type"), "")
		formatString, ok := contentType[key]
		if ok {
			format := src.ImgMap[formatString]
			fmt.Println(format)
			buf = src.ConvertImage(buf, format)
		}
	}

	fileType, _ := filetype.Match(buf.Bytes())

	headerResponse := metadata.Pairs(
		"fileExtension", fileType.Extension,
		"originalFileName", foundFile.Metadata.OriginalFileName,
		"transformHistory", strings.Join(foundFile.Metadata.TransformHistory, " | "),
	)
	grpc.SetHeader(c, headerResponse)
	return &DownloadResponse{ImageData: buf.Bytes()}, nil

}

func (s *ImgAPI) Delete(c context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	fileID := req.FileId
	deleteRequestData := src.DeleteRequestData{FileID: fileID}

	bucket, err := gridfs.NewBucket(db.DB.Database("ImgAPI"), db.Opt)
	if err != nil {
		return &DeleteResponse{FileId: fileID, Success: false}, err
	}
	result := src.DeleteFileFomDB(bucket, deleteRequestData)
	return &DeleteResponse{FileId: fileID, Success: result["success"].(bool)}, err
}

var contentType = map[string]string{
	"image/jpeg": "jpg",
	"image/bmp":  "bmp",
	"image/png":  "png",
	"image/tiff": "tif",
	"image/gif":  "gif",
}
