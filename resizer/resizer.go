package resizer

import (
	"fmt"
	"image"
	"imageResizerX/logs"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"go.uber.org/zap"
)

type Image struct {
	File   multipart.File
	Header *multipart.FileHeader
}

func (i *Image) Format() ImageFmt {
	mime := i.Header.Header.Get("Content-Type")
	imagefmt := strings.TrimSuffix(mime, "image/")

	return map[string]ImageFmt{
		PNG.format:  PNG,
		JPEG.format: JPEG,
	}[imagefmt]
}

func (i *Image) FilenameName() string {
	return i.Header.Filename
}

type ImageFmt struct {
	format string
}

func (i ImageFmt) String() string {
	return i.format
}

var (
	PNG  = ImageFmt{"PNG"}
	JPEG = ImageFmt{"JPEG"}
)

type ImageResizer struct {
	resize func(file multipart.File, width int, heigth int) (*image.NRGBA, error)
	encode func(file *os.File, img *image.NRGBA, fmt ImageFmt) error
}

func NewImageResizer() *ImageResizer {
	return &ImageResizer{
		resize: func(file multipart.File, width, heigth int) (*image.NRGBA, error) {
			img, err := imaging.Decode(file)
			if err != nil {
				logs.Logger.Error("Failed to performe image decode",
					zap.Error(err),
				)
				return nil, err
			}
			return imaging.Resize(img, width, heigth, imaging.Lanczos), nil
		},
		encode: func(file *os.File, img *image.NRGBA, format ImageFmt) error {

			f := map[ImageFmt]imaging.Format{
				JPEG: imaging.JPEG,
				PNG:  imaging.PNG,
			}[format]

			err := imaging.Encode(file, img, f)

			if err != nil {
				logs.Logger.Error("Failed to performe image encode",
					zap.Error(err),
				)
				return err
			}

			return nil
		},
	}
}

func (r *ImageResizer) ResizeImage(originalImage Image, width, heigth int) (string, error) {

	resizedImg, err := r.resize(originalImage.File, width, heigth)
	if err != nil {
		return "", err
	}

	uniqueName := r.generateUniqueFilename(originalImage.FilenameName())

	err = r.save(resizedImg, uniqueName, originalImage.Format())
	if err != nil {
		return "", err
	}

	return uniqueName, nil

}

func (r *ImageResizer) generateUniqueFilename(originalFilename string) string {
	sufix := filepath.Ext(originalFilename)
	base := strings.TrimSuffix(originalFilename, sufix)
	return fmt.Sprintf("%s_%d%s", base, time.Now().Unix(), sufix)
}

func (r *ImageResizer) save(img *image.NRGBA, filename string, format ImageFmt) error {

	outputFile, err := os.Create(filepath.Join("uploads", filename))
	if err != nil {
		logs.Logger.Error("Failed to performe output file creation",
			zap.Error(err),
		)
		return err
	}

	defer outputFile.Close()

	return r.encode(outputFile, img, format)
}
