package resizer

import (
	"fmt"
	"image"
	"imageResizerX/adapters"
	"imageResizerX/domain"
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

type Storer interface {
	Save(img *domain.ImageResized) error
}

type ImageResizer struct {
	resize func(file multipart.File, width int, heigth int) (*image.NRGBA, error)
	storer Storer
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
		storer: adapters.NewStorageInMemory(),
	}
}

func (r *ImageResizer) ResizeImage(originalImage *Image, width, heigth int) (string, error) {

	img, err := r.resize(originalImage.File, width, heigth)
	if err != nil {
		return "", err
	}

	uniqueName := r.generateUniqueFilename(originalImage.Filename)

	resizedImg := &domain.ImageResized{
		Img:    img,
		Name:   uniqueName,
		Format: originalImage.Format,
	}

	err = r.save(resizedImg)
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

func (r *ImageResizer) save(img *domain.ImageResized) error {
	return r.storer.Save(img)
}
