package main

import (
	"flag"
	"fmt"
	"net/http"
)

const (
	ModeView     = "view"     // default mode, supply by go std lib
	ModeDownload = "download" // every file will be download after clicked by mouse
)

func main() {
	// default port
	var port int
	var mode string
	port = *flag.Int("port", 8000, "listen port")
	mode = *flag.String("mode", ModeDownload, "[view: default mode| download: file download]")

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Listen on %s, mode=%s\n", addr, mode)

	// mode
	root := http.Dir("./")
	var handler http.Handler
	switch mode {
	case ModeView:
		handler = http.FileServer(root)
	case ModeDownload:
		handler = &DownloadHandler{root}
	default:
		fmt.Printf("argument error: invalid mode[%s]\n", mode)
		return
	}

	if err := http.ListenAndServe(addr, handler); err != nil {
		fmt.Print(err)
	}
}
