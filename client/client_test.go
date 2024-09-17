package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/h2non/filetype"
	"github.com/i-sentropic/imgAPI/pkg/src"
)

func TestAPIFetch(t *testing.T) {
	fetchRequestData := src.FetchRequestData{
		OriginalFileName: "badger-011",
		Url:              "https://www.wildsheffield.com/wp-content/uploads/2018/09/wildlifetrusts_40678106689-e1537524864604-1050x750.jpg",
	}
	fetchImageRequest := src.FetchImageRequest{
		Payload: []src.FetchRequestData{fetchRequestData},
	}
	data, err := json.Marshal(&fetchImageRequest)
	if err != nil {
		log.Fatal(err)
	}
	fetchReqFileName := "FetchRequest.json"
	err = os.WriteFile(fetchReqFileName, data, 0644)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := Post("http://localhost:8080/fetch", fetchReqFileName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
	tests := []FileData{}
	tests = append(tests, resp.Payload...)

	//load the bytes from the downloaded file and compare to a get
	for _, test := range tests {
		t.Run("Test file equality", func(t *testing.T) {
			resp, err := http.Get(fetchRequestData.Url)
			if err != nil {
				log.Fatal("Unable to down file from url")
			}
			bodyBytes, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()

			err = GetFile("http://localhost:8080/download/", test.FileId)
			if err != nil {
				t.Errorf("unable to download file %v (%v) from server", test.FileId, test.OriginalFileName)
			}

			fileName := fmt.Sprintf("%v.%v", test.FileId, test.FileExtension)
			bodyBytes2, err := os.ReadFile(fileName)
			if err != nil {
				log.Fatal("unable to open file after")
			}

			if !bytes.Equal(bodyBytes, bodyBytes2) {
				t.Errorf("file downloaded from server and file from API not equal")
			}
			os.Remove(fmt.Sprintf("%v.%v", test.FileId, test.FileExtension))
		})
	}
	os.Remove(fetchReqFileName)

}

func TestAPIUploadDownload(t *testing.T) {
	//compile test cases
	tests := []FileData{}
	resp, err := UploadFile("http://localhost:8080/upload", "image", "badger-004.jpg")
	if err != nil {
		log.Fatal(err)
	}
	tests = append(tests, resp.Payload...)
	resp, err = UploadMultipleFiles("http://localhost:8080/upload", "image", []string{"badger-001.jpg", "badger-002.jpg", "badger-003.jpg"})
	if err != nil {
		log.Fatal(err)
	}
	tests = append(tests, resp.Payload...)

	//fetch from server and assert the same as the file loaded
	for _, test := range tests {
		t.Run(fmt.Sprintf("Test %v.%v equal to %v.%v", test.OriginalFileName, test.FileExtension, test.FileId, test.FileExtension), func(t *testing.T) {
			err := GetFile("http://localhost:8080/download/", test.FileId)
			if err != nil {
				t.Errorf("unable to download file %v (%v) from server", test.FileId, test.OriginalFileName)
			}
			f1 := fmt.Sprintf("%v.%v", test.FileId, test.FileExtension)
			f2 := fmt.Sprintf("%v.%v", test.OriginalFileName, test.FileExtension)
			if !filesEqual(f1, f2) {
				t.Errorf("files not equal")
			}
			os.Remove(fmt.Sprintf("%v.%v", test.FileId, test.FileExtension))
		})
	}
}

func TestDelete(t *testing.T) {
	resp, err := UploadMultipleFiles("http://localhost:8080/upload", "image", []string{"badger-001.jpg", "badger-002.jpg", "badger-003.jpg"})
	if err != nil {
		log.Fatal(err)
	}
	testFiles := append([]FileData{}, resp.Payload...)
	type TestCase struct {
		FileId   string
		Expected bool
	}

	//compile struct for request to delete files sent to server, starting with a file not held on the db
	filename := "thisfiledoesntexist"
	failingtestcase := TestCase{
		FileId:   filename,
		Expected: false,
	}
	tests := []TestCase{failingtestcase}

	failingrequest := src.DeleteRequestData{FileID: filename}
	deleteRequestData := []src.DeleteRequestData{failingrequest}

	for _, test := range testFiles {
		deleteRequestData = append(deleteRequestData, src.DeleteRequestData{FileID: test.FileId})
		testcase := TestCase{
			FileId:   test.FileId,
			Expected: true,
		}
		tests = append(tests, testcase)
	}

	//generate delete request struct and marshal to JSON, save and then send as a request
	deleteReq := src.DeleteImageRequest{
		Payload: deleteRequestData,
	}
	bytes, err := json.Marshal(&deleteReq)
	if err != nil {
		log.Fatal(err)
	}
	deleteReqFileName := "DeleteRequest.json"
	err = os.WriteFile(deleteReqFileName, bytes, 0644)
	if err != nil {
		log.Fatal(err)
	}

	//send delete request
	resp, err = Post("http://localhost:8080/delete", deleteReqFileName)
	if err != nil {
		log.Fatal(err)
	}
	for idx, test := range tests {
		response := resp.Payload
		t.Run(fmt.Sprintf("Test %v", test.FileId), func(t *testing.T) {
			if response[idx].FileId != test.FileId {
				t.Errorf("Expected filename: %v, got: %v", test.FileId, response[idx].FileId)
			}
			if response[idx].Success != test.Expected {
				t.Errorf("Expected: %v, got %v", test.Expected, response[idx].Success)
			}
		})
	}
	os.Remove(deleteReqFileName)

}

func TestChangeFileFormat(t *testing.T) {
	originalFileExtension := "jpg"
	originalFileName := "badger-004"
	requestFile := fmt.Sprintf("%v.%v", originalFileName, originalFileExtension)
	resp, err := UploadFile("http://localhost:8080/upload", "image", requestFile)
	if err != nil {
		log.Fatal(err)
	}
	testFile := resp.Payload[0]
	tests := []struct {
		FileExtension string
		Expected      string
	}{
		{"fhfh", originalFileExtension},
		{"jpg", "jpg"},
		{"jpeg", "jpg"},
		{"tif", "tif"},
		{"tiff", "tif"},
		{"bmp", "bmp"},
		{"png", "png"},
		{"gif", "gif"},
	}

	//fetch from server and assert the file type matches the requested one
	for _, test := range tests {
		t.Run(fmt.Sprintf("Test file type: %v against expected: %v", test.FileExtension, test.Expected), func(t *testing.T) {
			requestFile := fmt.Sprintf("%v.%v", testFile.FileId, test.FileExtension)
			err := GetFile("http://localhost:8080/download/", requestFile)
			if err != nil {
				t.Errorf(fmt.Sprintf("unable to download file %v from server", requestFile))
			}
			//load file
			f1, err := os.ReadFile(fmt.Sprintf("%v.%v", testFile.FileId, test.Expected))
			if err != nil {
				t.Errorf(fmt.Sprintf("unable to load file from directory %v", requestFile))
			}
			err = os.Remove(fmt.Sprintf("%v.%v", testFile.FileId, test.Expected))
			if err != nil {
				t.Errorf("unable to delete file after loading bytes into memory")
			}
			//get file extension
			fileType, _ := filetype.Match(f1)
			if fileType.Extension != test.Expected {
				t.Errorf(fmt.Sprintf("Expected: %v, Got: %v", test.Expected, fileType.Extension))
			}
		})
	}

}

func filesEqual(file1 string, file2 string) bool {
	f1, err1 := os.ReadFile(file1)
	if err1 != nil {
		log.Fatal(err1)
	}
	f2, err2 := os.ReadFile(file2)
	if err2 != nil {
		log.Fatal(err2)
	}
	return bytes.Equal(f1, f2)
}
