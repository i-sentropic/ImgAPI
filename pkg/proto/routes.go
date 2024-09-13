package proto

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/h2non/filetype"
	"github.com/i-sentropic/imgAPI/pkg/src"
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

	headerResponse := metadata.Pairs("fileSize", fileSize,
		"fileExtension", response["fileExtension"].(string),
		"originalFileName", response["originalFileName"].(string),
	)
	grpc.SetHeader(c, headerResponse)

	resp := UploadResponse{FileId: response["fileId"].(string)}

	return &resp, nil
}

func (s *ImgAPI) Download(c context.Context, req *DownloadRequest) (*DownloadResponse, error) {
	fileID := req.FileId
	conv, format := src.CheckFileExtension(fileID)

	buf, foundFile, err := src.GetFileFromDB(fileID)
	if err != nil {
		return &DownloadResponse{}, err
	}

	//convert to new image type
	if conv {
		buf = src.ConvertImage(buf, format)
	}

	fileType, _ := filetype.Match(buf.Bytes())
	fileExtension := fileType.Extension
	fileMetaData := foundFile.Metadata

	headerResponse := metadata.Pairs("fileExtension", fileExtension,
		"originalFileName", fileMetaData.OriginalFileName)
	grpc.SetHeader(c, headerResponse)

}
