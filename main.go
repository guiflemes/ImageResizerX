package main

import (
	"fmt"
	"imageResizerX/middleware"
	"imageResizerX/ports"
	"imageResizerX/server"
	"net/http"
)

func main() {
	httpApp := ports.NewHttpApp()
	httpServer := server.NewHttpServer()

	httpServer.Post("/api/v1/upload", middleware.ImageFmtValidatorMiddleware(httpApp.UploadHandler))
	httpServer.Get("/", ports.Home)
	httpServer.Get("/ws", httpApp.WebsocketHandler)
	httpServer.Get("/api/v1/download/", httpApp.DownloadHandler)

	fmt.Println("Server is running on :8080...")
	http.ListenAndServe(":8080", httpServer)
}
