package adapters

import (
	"bufio"
	"errors"
	"image"
	"imageResizerX/domain"
	"imageResizerX/logs"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/disintegration/imaging"
	"go.uber.org/zap"
)

type FileManager interface {
	Open(filename string) error
	Write(data []byte) (n int, err error)
	Close() error
}

type fileManager struct {
	sync.Mutex
	file   *os.File
	writer *bufio.Writer
}

func NewFileManager() *fileManager {
	return &fileManager{
		file:   nil,
		writer: nil,
	}
}

func (fm *fileManager) Open(filename string) error {
	fm.Lock()
	defer fm.Unlock()

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		return err
	}

	fm.file = file
	fm.writer = bufio.NewWriter(file)

	return nil
}

func (fm *fileManager) Write(data []byte) (n int, err error) {
	fm.Lock()
	defer fm.Unlock()
	return fm.writer.Write(data)
}

func (fm *fileManager) Close() error {
	fm.Lock()
	defer fm.Unlock()

	if fm.file != nil {
		err := fm.writer.Flush()
		if err != nil {
			return err
		}

		err = fm.file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

type StorageInMemory struct {
	localStorage string
	fileManager  FileManager
	sweepCh      chan struct{}
	encode       func(file io.Writer, img *image.NRGBA, imgFormat string) error
}

func NewStorageInMemory() *StorageInMemory {
	repo := &StorageInMemory{
		localStorage: "uploads",
		fileManager:  NewFileManager(),
		sweepCh:      make(chan struct{}, 1),
		encode: func(file io.Writer, img *image.NRGBA, imgFormat string) error {

			f := map[string]imaging.Format{
				"jpeg": imaging.JPEG,
				"png":  imaging.PNG,
			}[imgFormat]

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

	repo.sweep()
	return repo
}

func (s *StorageInMemory) Save(img *domain.ImageResized) error {
	err := s.fileManager.Open(filepath.Join(s.localStorage, img.Name))
	defer s.fileManager.Close()
	if err != nil {
		logs.Logger.Error("Failed to performe output file creation",
			zap.Error(err),
		)
		return err
	}

	select {
	case s.sweepCh <- struct{}{}:
	default:
	}

	return s.encode(s.fileManager, img.Img, img.Format)
}

func (s *StorageInMemory) Retrieve(filename string) (*domain.MemoryImg, error) {
	filePath := filepath.Join(s.localStorage, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logs.Logger.Error("File not found")
		return nil, errors.New("file not found")
	}

	return &domain.MemoryImg{FilePath: filePath}, nil
}

func (s *StorageInMemory) clean() {
	dir, err := os.Open(s.localStorage)

	if err != nil {
		logs.Logger.Error("Dir not found")
		return
	}

	files, err := dir.ReadDir(0)

	if err != nil {
		logs.Logger.Error(err.Error())
		return
	}

	for i := range files {
		file := files[i]
		img := &domain.MemoryImg{FilePath: file.Name()}

		if img.IsValid() {
			continue
		}
		err := os.Remove(s.localStorage + "/" + img.FilePath)

		if err != nil {
			logs.Logger.Error(err.Error())
		}

	}

}

func (s *StorageInMemory) sweep() {
	go func() {
		for {
			<-s.sweepCh
			s.clean()
		}

	}()
}
