package src

import (
	"bytes"
	"errors"
	"fmt"
	"image"

	"github.com/h2non/filetype"
	"github.com/sunshineplan/imgconv"
)

func ExecuteTransform(buf *bytes.Buffer, request TransformRequestData, foundFile GridfsFileData) (*bytes.Buffer, []string, error) {
	//check that transform operation exists
	if _, ok := TransformMap[request.TransformationOperation.Operation]; !ok {
		return buf, nil, errors.New("failed to find operation in transform map")
	}

	//execute transformation
	transformHistory := foundFile.Metadata.TransformHistory
	if request.TransformationOperation.Operation == "watermark" {
		buf, err := WatermarkImage(buf, request)
		if err != nil {
			return buf, nil, err
		}
		transformHistory = append(transformHistory, "water mark applied")
	}
	if request.TransformationOperation.Operation == "resize" {
		buf, err := ResizeImage(buf, request)
		if err != nil {
			return buf, nil, err
		}
		transformString := fmt.Sprintf("Image resized @ %v%%", request.TransformationOperation.Parameter.(float64))
		transformHistory = append(transformHistory, transformString)
	}
	return buf, transformHistory, nil
}

func ConvertImage(buf *bytes.Buffer, format *imgconv.FormatOption) *bytes.Buffer {
	r := bytes.NewReader(buf.Bytes())
	img, err := imgconv.Decode(r)
	if err != nil {
		return buf
	}
	newBuf := bytes.NewBuffer(nil)
	err = imgconv.Write(newBuf, img, format)
	if err != nil {
		return buf
	}
	return newBuf
}

func WatermarkImage(buf *bytes.Buffer, request TransformRequestData) (*bytes.Buffer, error) {
	markImage, err := imgconv.Open("files/watermark.jpg")
	if err != nil {
		return buf, errors.New("unable to read in watermark image")
	}
	param, _ := request.TransformationOperation.Parameter.(float64)
	opt := &imgconv.WatermarkOption{
		Mark:    markImage,
		Opacity: uint8(param),
		Offset:  image.Pt(5, 5),
	}
	buf, err = ModifyImage(buf, opt)
	if err != nil {
		return buf, errors.New("unable to apply watermark")
	}
	return buf, nil
}

func ResizeImage(buf *bytes.Buffer, request TransformRequestData) (*bytes.Buffer, error) {
	param, _ := request.TransformationOperation.Parameter.(float64)
	opt := &imgconv.ResizeOption{
		Percent: float64(param),
	}
	buf, err := ModifyImage(buf, opt)
	if err != nil {
		return buf, errors.New("unable to resize")
	}
	return buf, nil
}

func ModifyImage(buf *bytes.Buffer, option interface{}) (*bytes.Buffer, error) {
	r := bytes.NewReader(buf.Bytes())
	img, err := imgconv.Decode(r)
	if err != nil {
		return buf, errors.New("failed to decode image in conversion")
	}

	fileExt, err := filetype.Match(buf.Bytes())
	if err != nil {
		return buf, errors.New("failed to detect file type")
	}

	format, ok := ImgMap[fileExt.Extension]
	if !ok {
		return buf, errors.New("failed to get file type from map")
	}

	switch opt := option.(type) {
	case *imgconv.WatermarkOption:
		newBuf, err := Modify[*imgconv.WatermarkOption](imgconv.Watermark, img, opt, format)
		if err != nil {
			return buf, errors.New("failed to write image in conversion")
		}
		return newBuf, nil
	case *imgconv.ResizeOption:
		newBuf, err := Modify[*imgconv.ResizeOption](imgconv.Resize, img, opt, format)
		if err != nil {
			return buf, errors.New("failed to write image in conversion")
		}
		return newBuf, nil
	default:
		return buf, errors.New("failed to detect image transformation operation")
	}
}

func Modify[T ImgMod](obj func(image.Image, T) image.Image, img image.Image, opt T, format *imgconv.FormatOption) (*bytes.Buffer, error) {
	imgOut := obj(img, opt)
	newBuf := bytes.NewBuffer(nil)
	err := imgconv.Write(newBuf, imgOut, format)
	return newBuf, err
}
