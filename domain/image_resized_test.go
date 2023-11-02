package domain

import (
	"fmt"
	"testing"

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
			fmt.Println("R", result)
			assert.Equal(result, scenerio.expectResult)
		})
	}

}
