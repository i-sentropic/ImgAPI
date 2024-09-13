package src

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/i-sentropic/imgAPI/pkg/db"
	"github.com/sunshineplan/imgconv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// custom struct to unmarshal BSON data
type GridfsFileData struct {
	FileName string         `bson:"filename"`
	Length   int64          `bson:"length"`
	Metadata MetaDataFields `bson:"metadata"`
}

type MetaDataFields struct {
	OriginalFileName string   `bson:"originalFileName"`
	TransformHistory []string `bson:"transformHistory"`
	FileExtension    string   `bson:"fileExtension"`
}

func DeleteFileFomDB(bucket *gridfs.Bucket, req DeleteRequestData) map[string]interface{} {
	result := map[string]interface{}{"fileId": req.FileID}
	objID, err := primitive.ObjectIDFromHex(req.FileID)
	if err != nil {
		result["success"] = false
		return result
	}
	err = bucket.Delete(objID)
	if err != nil {
	result["success"] = false
		return result
	}
	result["success"] = true
	return result
}

func GetFileFromDB(fileID string) (*bytes.Buffer, GridfsFileData, error) {
	//remove any file extensions
	fileID = strings.Split(fileID, ".")[0]

	objID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return nil, GridfsFileData{}, err
	}

	bucket, err := gridfs.NewBucket(db.DB.Database("ImgAPI"), db.Opt)
	if err != nil {
		return nil, GridfsFileData{}, err
	}

	buf := bytes.NewBuffer(nil)
	_, err = bucket.DownloadToStream(objID, buf)
	if err != nil {
		return nil, GridfsFileData{}, err
	}

	//Find file metadata
	filter := bson.M{"_id": bson.M{"$eq": objID}}
	cursor, err := bucket.Find(filter)
	if err != nil {
		return nil, GridfsFileData{}, err
	}
	var foundFiles []GridfsFileData
	if err = cursor.All(context.TODO(), &foundFiles); err != nil {
		return nil, GridfsFileData{}, err
	}
	foundFile := foundFiles[0]

	return buf, foundFile, nil
}

func SaveToDB(bodyBytes *[]byte, fileMetaData *MetaDataFields, fileName string) (map[string]interface{}, error) {
	bucket, err := gridfs.NewBucket(db.DB.Database("ImgAPI"), db.Opt)
	if err != nil {
		return nil, err
	}

	uploadStream, err := bucket.OpenUploadStream(fileName, &options.UploadOptions{
		Metadata: map[string]interface{}{
			"fileExtension":    fileMetaData.FileExtension,
			"originalFileName": fileMetaData.OriginalFileName,
			"transformHistory": fileMetaData.TransformHistory,
		}},
	)

	if err != nil {
		return nil, err
	}

	defer uploadStream.Close()

	fileSize, err := uploadStream.Write(*bodyBytes)
	if err != nil {
		return nil, err
	}

	fileId, err := json.Marshal(uploadStream.FileID)
	if err != nil {
		return nil, err
	}

	responseText := map[string]interface{}{
		"fileId":           strings.Trim(string(fileId), `"`),
		"fileSize":         fileSize,
		"fileExtension":    fileMetaData.FileExtension,
		"originalFileName": fileMetaData.OriginalFileName,
		"transformHistory": fileMetaData.TransformHistory}

	return responseText, nil
}

func SendRequest(req FetchRequestData) (*[]byte, error) {
	buf := &bytes.Buffer{}
	request, err := http.NewRequest("GET", req.Url, buf)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil || resp.StatusCode != 200 {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &bodyBytes, nil
}

func CleanFileName(fileName string) string {
	file := strings.Split(fileName, ".")
	return file[0]
}

func CheckFileExtension(fileName string) (bool, *imgconv.FormatOption) {
	match, _ := regexp.MatchString(".(gif|jpe?g|tiff?|png|bmp)$", fileName)
	if match {
		fileName = strings.Split(fileName, ".")[1]
		return true, ImgMap[fileName]
	}
	return false, nil
}

var ImgMap = map[string]*imgconv.FormatOption{
	"jpg":  &imgconv.FormatOption{Format: imgconv.JPEG},
	"jpeg": &imgconv.FormatOption{Format: imgconv.JPEG},
	"tif":  &imgconv.FormatOption{Format: imgconv.TIFF},
	"tiff": &imgconv.FormatOption{Format: imgconv.TIFF},
	"bmp":  &imgconv.FormatOption{Format: imgconv.BMP},
	"png":  &imgconv.FormatOption{Format: imgconv.PNG},
	"gif":  &imgconv.FormatOption{Format: imgconv.GIF},
}

var TransformMap = map[string]struct{}{
	"watermark": struct{}{},
	"resize":    struct{}{},
}
