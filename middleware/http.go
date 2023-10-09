package middleware

import (
	"context"
	"net/http"
)

var validImageInputs = []string{
	"image/png",
	"image/jpeg",
}

func matchImageFmt(format string) bool {
	for _, f := range validImageInputs {
		if format == f {
			return true
		}
	}
	return false
}

type ImageFmt string

func ImageFmtValidatorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to read uploaded file.", http.StatusInternalServerError)
			return
		}

		defer file.Close()

		buffer := make([]byte, 512)
		_, err = file.Read(buffer)

		if err != nil {
			http.Error(w, "Failed to read file content.", http.StatusInternalServerError)
		}

		contentType := http.DetectContentType(buffer)

		if !matchImageFmt(contentType) {
			http.Error(w, "Invalid image format. Only images are allowed.", http.StatusBadRequest)
		}

		ctx := context.WithValue(r.Context(), ImageFmt(contentType), contentType)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
