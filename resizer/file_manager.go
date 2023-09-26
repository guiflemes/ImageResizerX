package resizer

import (
	"bufio"
	"os"
	"sync"
)

type FileManager struct {
	sync.Mutex
	file   *os.File
	writer *bufio.Writer
}

func NewFileManager() *FileManager {
	return &FileManager{
		file:   nil,
		writer: nil,
	}
}

func (fm *FileManager) Open(filename string) error {
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

func (fm *FileManager) Write(data []byte) error {
	fm.Lock()
	defer fm.Unlock()

	_, err := fm.writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (fm *FileManager) Close() error {
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
