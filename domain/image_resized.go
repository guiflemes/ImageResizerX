package domain

import (
	"image"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ImageResized struct {
	Img    *image.NRGBA
	Name   string
	Format string
}

type MemoryImg struct {
	FilePath string
}

func (m *MemoryImg) IsValid() bool {
	lifeTime := int64(60 * 5)
	return m.CreatedAtUnix()+lifeTime >= time.Now().Unix()
}

func (m *MemoryImg) CreatedAtUnix() int64 {
	rgx := regexp.MustCompile(`_(.*?)\.`)

	match := rgx.FindStringSubmatch(m.FilePath)
	unixDateAsString := match[1]

	createdAt, _ := strconv.ParseInt(unixDateAsString, 10, 64)
	return createdAt
}

func (m *MemoryImg) Format() string {
	parts := strings.Split(m.FilePath, ".")
	return parts[len(parts)-1]
}
