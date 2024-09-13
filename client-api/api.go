package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	UploadFile("http://localhost:8080/upload", "image", "badger-004.jpg")
	UploadMultipleFiles("http://localhost:8080/upload", "image", []string{"badger-001.jpg", "badger-002.jpg", "badger-003.jpg"})
	// GetFile("http://localhost:8080/download/", "66e2ddc8f754d67716625e31")
	// GenericPost("http://localhost:8080/fetch", "fetchImageRequest.json")
	//GenericPost("http://localhost:8080/transform", "transformImageRequest.json")
	//GenericPost("http://localhost:8080/delete", "deleteImageRequest.json")
	//testConv()

}

func GenericGet(url string, fileRef string, fileExt string) error {
	resp, err := http.Get(url + fileRef + fileExt)
	if err != nil {
		return err
	}
	fmt.Println(resp)
	return nil
}

func GenericPost(url string, fileRef string) error {
	jsonbytes, err := os.ReadFile(fileRef)
	if err != nil {
		return err
	}

	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonbytes))
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(request)

	if err != nil || resp.StatusCode != 200 {
		return err
	}

	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Println("response Body:", string(bodyBytes))
	return nil
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
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	//fmt.Println("response Body:", string(bodyBytes))
	return nil
}

func UploadFile(url string, paramName string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(filePath))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	err = writer.Close()
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", url, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// add better handling to the resp handler
	fmt.Println(resp.Header, resp.StatusCode)
	bodyBytes, err := io.ReadAll(resp.Body)
	fmt.Println(string(bodyBytes))
	return nil
}

func UploadMultipleFiles(url string, paramName string, filePaths []string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		part, err := writer.CreateFormFile(paramName, filepath.Base(filePath))
		if err != nil {
			return err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return err
		}
	}
	err := writer.Close()
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", url, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	fmt.Println(string(bodyBytes))
	return nil
}

// func reg() {
// 	inputTexts := []string{".jpg", "te.jpgds", "test.TIF"}

// 	for _, inputText := range inputTexts {

// 		match, err := regexp.MatchString(".(gif|jpe?g|tiff?|png|bmp)$", inputText)
// 		if err == nil {
// 			fmt.Println("Match:", match)
// 		} else {
// 			fmt.Println("Error:", err)
// 		}
// 	}

// 	cheese := "mystring"
// 	beans := strings.Split(cheese, ".")
// 	fmt.Println(beans)

// }

// func convert() {
// 	bodyBytes, _ := os.ReadFile("badger-001.jpg")
// 	r := bytes.NewReader(bodyBytes)
// 	img, err := imgconv.Decode(r)
// 	if err != nil {
// 		fmt.Print("error")
// 	}

// 	w := bytes.NewBuffer(nil)

// 	err = imgconv.Write(w, img, &imgconv.FormatOption{Format: imgconv.TIFF})
// 	if err != nil {
// 		fmt.Print("error")
// 	}

// 	fileType, _ := filetype.Match(w.Bytes())
// 	fmt.Println(fileType)

// }
