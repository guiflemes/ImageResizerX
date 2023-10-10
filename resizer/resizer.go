package resizer

import (
	"fmt"
	"image"
	"imageResizerX/logs"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"go.uber.org/zap"
)

type Image struct {
	File     multipart.File
	Filename string
	Format   string
}

type ImageResizer struct {
	resize func(file multipart.File, width int, heigth int) (*image.NRGBA, error)
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
	}
}

func (r *ImageResizer) ResizeImage(originalImage *Image, width, heigth int) (string, error) {

	resizedImg, err := r.resize(originalImage.File, width, heigth)
	if err != nil {
		return "", err
	}

	uniqueName := r.generateUniqueFilename(originalImage.Filename)

	err = r.save(resizedImg, uniqueName, originalImage.Format)
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

func (r *ImageResizer) save(img *image.NRGBA, filename string, format string) error {
	return nil

	// err := r.fileManager.Open(filepath.Join("uploads", filename))
	// defer r.fileManager.Close()

	// if err != nil {
	// 	logs.Logger.Error("Failed to performe output file creation",
	// 		zap.Error(err),
	// 	)
	// 	return err
	// }

	// return r.encode(r.fileManager, img, format)

}
