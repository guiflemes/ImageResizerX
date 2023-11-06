package ports

import (
	"context"
	"errors"
	"imageResizerX/adapters"
	"imageResizerX/logs"
	"imageResizerX/middleware"
	"imageResizerX/resizer"
	"log"
	"net/http"
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

type ImageServer func(w http.ResponseWriter, r *http.Request, filename string)

type httpApp struct {
	runner           Runner
	imageResize      *resizer.ImageResizer
	websocketHandler WebsocketHandler
	websocketOptions *websocket.AcceptOptions
	imageServer      ImageServer
}

func NewHttpApp() *httpApp {
	localDiskRepo := adapters.NewStorageInMemory()

	return &httpApp{
		runner:           resizer.NewImagePool(5),
		imageResize:      resizer.NewImageResizer(localDiskRepo),
		websocketHandler: resizer.DefaultwebsocketClient(),
		websocketOptions: &websocket.AcceptOptions{OriginPatterns: []string{"127.0.0.0"}},
		imageServer: func(w http.ResponseWriter, r *http.Request, filename string) {
			img, err := localDiskRepo.Retrieve(filename)

			if err != nil {
				http.Error(w, "File not found", http.StatusNotFound)
			}

			w.Header().Set("Content-Type", "image/"+img.Format())
			w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(filename))
			http.ServeFile(w, r, img.FilePath)
		},
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
		logs.Logger.Info("Closing WebSocket connection due to abnormal closure or going away", zap.Error(err))
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

	a.imageServer(w, r, filename)
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
