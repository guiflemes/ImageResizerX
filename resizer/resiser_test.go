package resizer

import (
	"errors"
	"fmt"
	"image"
	"imageResizerX/domain"
	"mime/multipart"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type StoreStub struct {
	data     map[string]*domain.ImageResized
	dataLock sync.RWMutex
}

func NewStoreStub() *StoreStub {
	return &StoreStub{
		data: make(map[string]*domain.ImageResized),
	}
}

func (s *StoreStub) Save(img *domain.ImageResized) error {

	if strings.Contains(img.Name, "error") {
		return errors.New("error saving on db")
	}

	s.dataLock.Lock()
	s.data[img.Name] = img
	s.dataLock.Unlock()
	return nil
}

func (s *StoreStub) Get(name string) *domain.ImageResized {
	s.dataLock.RLock()
	defer s.dataLock.RUnlock()
	img, ok := s.data[name]

	if !ok {
		return nil
	}

	return img
}

type FileStub struct{ mock.Mock }

func (f *FileStub) Read(p []byte) (n int, err error) {
	args := f.Called()
	result := args.Get(0)
	return result.(int), args.Error(1)
}

func (f *FileStub) ReadAt(p []byte, off int64) (n int, err error) {
	args := f.Called()
	result := args.Get(0)
	return result.(int), args.Error(1)
}

func (f *FileStub) Seek(offset int64, whence int) (int64, error) {
	args := f.Called()
	result := args.Get(0)
	return result.(int64), args.Error(1)
}

func (f *FileStub) Close() error {
	args := f.Called()
	return args.Error(0)
}

func TestResizeImage(t *testing.T) {
	assert := assert.New(t)
	storer := NewStoreStub()

	resizer := &ImageResizer{
		resize: func(file multipart.File, width int, heigth int) (*image.NRGBA, error) {
			return &image.NRGBA{}, nil
		},
		storer: storer,
	}

	for _, scenerio := range []*Image{
		{
			File:     &FileStub{},
			Filename: "test1.png",
			Format:   "png",
		},
		{
			File:     &FileStub{},
			Filename: "test2.jpg",
			Format:   "jpg",
		},
		{
			File:     &FileStub{},
			Filename: "error1.jpg",
			Format:   "jpg",
		},
		{
			File:     &FileStub{},
			Filename: "error2.jpg",
			Format:   "jpg",
		},
	} {
		t.Run(scenerio.Filename, func(t *testing.T) {
			uniqueName, err := resizer.ResizeImage(scenerio, 200, 300)
			img := storer.Get(uniqueName)

			if err != nil {
				assert.Nil(img)
				return
			}

			assert.Contains(uniqueName, strings.Replace(scenerio.Filename, fmt.Sprintf(".%s", scenerio.Format), "", -1))
			assert.NotNil(img)
		})
	}

}
