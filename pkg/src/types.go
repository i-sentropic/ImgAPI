package src

import "github.com/sunshineplan/imgconv"

// custom type for union of imgconv option types
type ImgMod interface {
	*imgconv.WatermarkOption | *imgconv.ResizeOption
}

//custom struct to unmarshal post request data for multiple images
type ServeImageRequest struct {
	Payload []ImageProperties `json:"payload"`
}
type ImageProperties struct {
	FileExtension    string `json:"fileExtension"`
	FileID           string `json:"fileID"`
	FileSize         int    `json:"fileSize"`
	OriginalFileName string `json:"originalFileName"`
}

//custom struct to unmarshal post request data for multiple image downloads from external url
type FetchImageRequest struct {
	Payload []FetchRequestData `json:"payload"`
}

type FetchRequestData struct {
	OriginalFileName string `json:"originalFileName"`
	Url              string `json:"url"`
}

//custom struct to unmarshal post request data for multiple image deletion
type DeleteImageRequest struct {
	Payload []DeleteRequestData `json:"payload"`
}

type DeleteRequestData struct {
	FileID string `json:"fileID"`
}

//custom struct to unmarshal post request data for multiple image transformation
type TransformImageRequest struct {
	Payload []TransformRequestData
}

type TransformRequestData struct {
	FileID                  string                      `json:"fileID"`
	TransformationOperation TransformationOperationData `json:"transformationOperation"`
	Overwrite               bool                        `json:"overwrite"`
}

type TransformationOperationData struct {
	Operation string      `json:"operation"`
	Parameter interface{} `json:"parameter"`
}
