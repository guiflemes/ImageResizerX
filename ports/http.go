package ports

import (
	"context"
	"errors"
	"imageResizerX/logs"
	"imageResizerX/middleware"
	"imageResizerX/resizer"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"nhooyr.io/websocket"

	"github.com/CloudyKit/jet/v6"
)

type Runner interface {
	RunTask(task func())
}

type WebsocketHandler interface {
	Handle(ctx context.Context, conn resizer.WebsocketConn) error
	Brodcast(msg resizer.Message)
}

type httpApp struct {
	runner           Runner
	imageResize      *resizer.ImageResizer
	websocketHandler WebsocketHandler
	websocketOptions *websocket.AcceptOptions
}

func NewHttpApp() *httpApp {
	return &httpApp{
		runner:           resizer.NewImagePool(5),
		imageResize:      resizer.NewImageResizer(),
		websocketHandler: resizer.DefaultwebsocketClient(),
		websocketOptions: &websocket.AcceptOptions{OriginPatterns: []string{"127.0.0.0"}},
	}
}

func (a *httpApp) UploadHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	imageFmt := r.Context().Value(middleware.ImgFmt).(string)

	a.runner.RunTask(func() {
		message := resizer.Message{Action: "processing_failed", DownloadUrl: ""}

		out, err := a.imageResize.ResizeImage(
			&resizer.Image{File: file, Filename: header.Filename, Format: imageFmt},
			300,
			200)

		if err == nil {
			message.Action = "processing_complete"
			message.DownloadUrl = "/api/v1/download/" + out
		}

		a.websocketHandler.Brodcast(message)

	})

}

func (a *httpApp) WebsocketHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := websocket.Accept(w, r, a.websocketOptions)

	if err != nil {
		logs.Logger.Error("Error accepting WebSocket connection", zap.Error(err))
		return
	}

	defer conn.Close(websocket.StatusInternalError, "")
	ctx := r.Context()

	err = a.websocketHandler.Handle(ctx, conn)

	if errors.Is(err, context.Canceled) {
		return
	}

	if websocket.CloseStatus(err) == websocket.StatusAbnormalClosure || websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}

	if err != nil {
		logs.Logger.Error("Error accepting WebSocket connection", zap.Error(err))
	}

}

func (a *httpApp) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")

	if len(segments) < 5 {
		logs.Logger.Error("Invalid Url", zap.String("url", r.URL.Path))
		http.Error(w, "Invalid Url", http.StatusBadGateway)
		return
	}

	filename := segments[len(segments)-1]

	if filename == "" {
		logs.Logger.Error("Path param should not be empty")
		http.Error(w, "Invalid Url", http.StatusBadGateway)
		return
	}

	filePath := filepath.Join("uploads", filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logs.Logger.Error("File not found")
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(filename))
	w.Header().Set("Content-Type", "image/png")
	http.ServeFile(w, r, filePath)
}

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil

}
