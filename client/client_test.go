package client

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/h2non/filetype"
	"github.com/i-sentropic/imgAPI/pkg/src"
)

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
