package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type APIResponse struct {
	Payload []FileData `json:"payload"`
	Success bool       `json:"success"`
}

type FileData struct {
	FileExtension    string   `json:"fileExtension"`
	FileId           string   `json:"fileId"`
	FileSize         string   `json:"fileSize"`
	OriginalFileName string   `json:"originalFileName"`
	TransformHistory []string `json:"transformHistory"`
	Success          bool     `json:"success"`
}

type APIHeaderResponse struct {
}

// func main() {
// 	//UploadFile("http://localhost:8080/upload", "image", "badger-004.jpg")
// 	//UploadMultipleFiles("http://localhost:8080/upload", "image", []string{"badger-001.jpg", "badger-002.jpg", "badger-003.jpg"})
// 	// GetFile("http://localhost:8080/download/", "66e2ddc8f754d67716625e31")
// 	// GenericPost("http://localhost:8080/fetch", "fetchImageRequest.json")
// 	GenericPost("http://localhost:8080/transform", "transformImageRequest.json")
// 	//GenericPost("http://localhost:8080/delete", "deleteImageRequest.json")
// }

func GenericGet(url string, fileRef string, fileExt string) error {
	resp, err := http.Get(url + fileRef + fileExt)
	if err != nil {
		return err
	}
	fmt.Println(resp)
	return nil
}

func generateResponseObject(resp *http.Response) (APIResponse, error) {
	respData := &APIResponse{}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return APIResponse{}, err
	}
	json.Unmarshal(bodyBytes, respData)
	return *respData, nil
}

func GenericPost(url string, fileRef string) (APIResponse, error) {
	jsonbytes, err := os.ReadFile(fileRef)
	if err != nil {
		return APIResponse{}, err
	}

	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonbytes))
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(request)

	if err != nil || resp.StatusCode != 200 {
		return APIResponse{}, err
	}

	defer resp.Body.Close()

	respData, err := generateResponseObject(resp)
	if err != nil {
		return respData, err
	}

	return respData, nil
}

func GetFile(url string, fileID string) error {
	body := &bytes.Buffer{}
	request, _ := http.NewRequest("GET", url+fileID, body)
	client := &http.Client{}
	resp, err := client.Do(request)

	if err != nil || resp.StatusCode != 200 {
		return err
	}

	defer resp.Body.Close()

	fileExtension := resp.Header["File-Extension"][0]
	fileName := strings.Split(fileID, ".")[0] + "." + fileExtension

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = os.WriteFile(fileName, bodyBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func UploadFile(url string, paramName string, filePath string) (APIResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return APIResponse{}, err
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(filePath))
	if err != nil {
		return APIResponse{}, err
	}
	_, err = io.Copy(part, file)
	err = writer.Close()
	if err != nil {
		return APIResponse{}, err
	}
	request, err := http.NewRequest("POST", url, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return APIResponse{}, err
	}
	defer resp.Body.Close()

	respData, err := generateResponseObject(resp)
	if err != nil {
		return respData, err
	}
	return respData, err
}

func UploadMultipleFiles(url string, paramName string, filePaths []string) (APIResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return APIResponse{}, err
		}
		defer file.Close()
		part, err := writer.CreateFormFile(paramName, filepath.Base(filePath))
		if err != nil {
			return APIResponse{}, err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return APIResponse{}, err
		}
	}
	err := writer.Close()
	if err != nil {
		return APIResponse{}, err
	}
	request, err := http.NewRequest("POST", url, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return APIResponse{}, err
	}
	defer resp.Body.Close()

	respData, err := generateResponseObject(resp)
	if err != nil {
		return respData, err
	}
	return respData, err
}
