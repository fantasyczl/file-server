package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	// default port
	var port int
	port = *flag.Int("port", 8000, "listen port")

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Listen on %s\n", addr)

	if err := http.ListenAndServe(addr, http.FileServer(http.Dir("./"))); err != nil {
		fmt.Print(err)
	}
}