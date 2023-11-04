package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreatedAtUnix(t *testing.T) {
	assert := assert.New(t)

	type testCase struct {
		img          *MemoryImg
		expectResult int64
	}

	for _, scenerio := range []testCase{
		{
			img:          &MemoryImg{FilePath: "testimage1_1698718519.png"},
			expectResult: 1698718519,
		},
		{
			img:          &MemoryImg{FilePath: "test__image_-2_1998718513.png"},
			expectResult: 1998718513,
		},
		{
			img:          &MemoryImg{FilePath: "testimage_3_1998718513.png"},
			expectResult: 1998718513,
		},
	} {
		t.Run(scenerio.img.FilePath, func(t *testing.T) {
			result := scenerio.img.CreatedAtUnix()
			assert.Equal(result, scenerio.expectResult)
		})
	}

}

func TestIsValid(t *testing.T) {
	assert := assert.New(t)

	type testCase struct {
		img          *MemoryImg
		expectResult bool
	}

	for _, scenario := range []testCase{
		{
			img:          &MemoryImg{FilePath: "testimage1_1698718519.png"},
			expectResult: false,
		},
		{
			img:          &MemoryImg{FilePath: fmt.Sprintf("testimage2_%v.png", time.Now().Unix())},
			expectResult: true,
		},
	} {
		t.Run(scenario.img.FilePath, func(t *testing.T) {
			result := scenario.img.IsValid()
			assert.Equal(result, scenario.expectResult)
		})
	}
}

func TestFormat(t *testing.T) {
	assert := assert.New(t)

	type testCase struct {
		img          *MemoryImg
		expectResult string
	}

	for _, scenario := range []testCase{
		{
			img:          &MemoryImg{FilePath: "testimage1_1698718519.png"},
			expectResult: "png",
		},
		{
			img:          &MemoryImg{FilePath: "testimage1_1698718519.jpeg"},
			expectResult: "jpeg",
		},
		{
			img:          &MemoryImg{FilePath: "testimage1_1698718519.jpg"},
			expectResult: "jpg",
		},
	} {
		t.Run(scenario.img.FilePath, func(t *testing.T) {
			result := scenario.img.Format()
			assert.Equal(result, scenario.expectResult)
		})
	}
}
