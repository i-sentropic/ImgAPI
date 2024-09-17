package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype"
	"github.com/i-sentropic/imgAPI/pkg/db"
	"github.com/i-sentropic/imgAPI/pkg/src"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

func uploadImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		form, _ := c.MultipartForm()
		files := form.File["image"]
		var responseText []map[string]interface{}

		for _, fileheader := range files {
			file, err := fileheader.Open()
			header := fileheader

			if err != nil {
				log.Fatal(err)
				c.JSON(http.StatusBadRequest, err.Error())
				return
			}
			defer file.Close()

			buf := bytes.NewBuffer(nil)
			if _, err := io.Copy(buf, file); err != nil {
				log.Fatal(err)
				c.JSON(http.StatusBadRequest, err.Error())
				return
			}

			//Generate meta data and file name
			bodyBytes := buf.Bytes()
			fileType, _ := filetype.Match(bodyBytes)
			fileExt := fileType.Extension
			originalFileName := src.CleanFileName(header.Filename)
			fileName := time.Now().Format(time.RFC3339) + "_" + originalFileName
			transformHistory := []string{}

			fileMetaData := &src.MetaDataFields{originalFileName, transformHistory, fileExt}

			response, err := src.SaveToDB(&bodyBytes, fileMetaData, fileName)

			if err != nil {
				continue
			}

			responseText = append(responseText, response)

		}
		c.JSON(http.StatusOK, map[string]interface{}{"success": true, "payload": responseText})
	}
}

func transformImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqData := src.TransformImageRequest{}
		c.ShouldBindJSON(&reqData)
		var responseText []map[string]interface{}

		for _, request := range reqData.Payload {

			result := map[string]interface{}{"fileId": request.FileID}

			buf, foundFile, err := src.GetFileFromDB(request.FileID)
			if err != nil {
				result["success"] = false
				responseText = append(responseText, result)
				continue
			}

			buf, transformHistory, err := src.ExecuteTransform(buf, request, foundFile)
			if err != nil {
				result["success"] = false
				responseText = append(responseText, result)
				continue
			}

			bodyBytes := buf.Bytes()
			fileType, _ := filetype.Match(bodyBytes)
			fileExt := fileType.Extension
			originalFileName := foundFile.Metadata.OriginalFileName
			fileMetaData := &src.MetaDataFields{originalFileName, transformHistory, fileExt}

			fileName := time.Now().Format(time.RFC3339) + "_" + originalFileName

			result, err = src.SaveToDB(&bodyBytes, fileMetaData, fileName)
			if err != nil {
				result["success"] = false
				responseText = append(responseText, result)
				continue
			}

			result["success"] = true
			result["sourceFileId"] = request.FileID
			responseText = append(responseText, result)
		}
		c.JSON(http.StatusOK, map[string]interface{}{"success": true, "payload": responseText})
	}
}

func fetchImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqData := src.FetchImageRequest{}
		c.ShouldBindJSON(&reqData)
		var responseText []map[string]interface{}

		for _, requestData := range reqData.Payload {
			bodyBytes, err := src.SendRequest(requestData)
			if err != nil {
				continue
			}

			fileType, _ := filetype.Match(*bodyBytes)
			fileExt := fileType.Extension
			originalFileName := requestData.OriginalFileName
			fileName := time.Now().Format(time.RFC3339) + "_" + originalFileName
			transformHistory := []string{}
			fileMetaData := &src.MetaDataFields{originalFileName, transformHistory, fileExt}

			response, err := src.SaveToDB(bodyBytes, fileMetaData, fileName)

			if err != nil {
				continue
			}

			responseText = append(responseText, response)

		}
		c.JSON(http.StatusOK, map[string]interface{}{"success": true, "payload": responseText})
	}
}

func serveImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := strings.TrimPrefix(c.Request.URL.Path, "/download/")

		conv, format := src.CheckFileExtension(fileID)

		buf, foundFile, err := src.GetFileFromDB(fileID)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		//convert to new image type
		if conv {
			buf = src.ConvertImage(buf, format)
		}

		fileType, _ := filetype.Match(buf.Bytes())
		contentType := http.DetectContentType(buf.Bytes())
		c.Writer.Header().Add("Content-Type", contentType)
		c.Writer.Header().Add("Content-Length", strconv.Itoa(len(buf.Bytes())))
		c.Writer.Header().Add("File-Extension", fileType.Extension)

		metaData, _ := json.Marshal(&foundFile.Metadata)
		c.Writer.Header().Add("Source-File-Meta-Data", string(metaData))

		c.Writer.Write(buf.Bytes())
	}
}

func deleteImage() gin.HandlerFunc {
	return func(c *gin.Context) {

		reqData := src.DeleteImageRequest{}
		c.ShouldBindJSON(&reqData)
		var responseText []map[string]interface{}

		bucket, _ := gridfs.NewBucket(db.DB.Database("ImgAPI"), db.Opt)

		for _, req := range reqData.Payload {
			result := src.DeleteFileFomDB(bucket, req)
			responseText = append(responseText, result)
		}
		c.JSON(http.StatusOK, map[string]interface{}{"success": true, "payload": responseText})
	}
}
