package domain

import "image"

type ImageResized struct {
	Img    *image.NRGBA
	Name   string
	Format string
}
