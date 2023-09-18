package ports

import (
	"context"
	"errors"
	"imageResizerX/logs"
	"imageResizerX/resizer"
	"log"
	"net/http"
	"path/filepath"
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
		websocketOptions: &websocket.AcceptOptions{OriginPatterns: []string{"localhost:3000"}},
	}
}

func (a *httpApp) UploadHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	a.runner.RunTask(func() {
		message := resizer.Message{Action: "processing_failed", DownloadUrl: ""}

		out, err := a.imageResize.ResizeImage(resizer.OriginalFile{File: file, Name: header.Filename}, 300, 200, resizer.JPEG)

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
	filename := strings.TrimPrefix(r.URL.Path, "/download/")
	http.ServeFile(w, r, filepath.Join("uploads", filename))
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
