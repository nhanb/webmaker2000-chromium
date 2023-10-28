package main

import (
	"embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"sync"
)

//go:embed frontend
var frontend embed.FS

func main() {

	// Snatch a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	frontendUrl := fmt.Sprintf("http://localhost:%d/frontend/", port)
	websocketUrl := fmt.Sprintf("http://localhost:%d/websocket", port)
	fmt.Println("Serving:")
	fmt.Printf("- %s\n", frontendUrl)
	fmt.Printf("- %s\n", websocketUrl)

	// Start http server
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Server starting")
		srv := &http.Server{}
		http.Handle("/", http.FileServer(http.FS(frontend)))
		http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("TODO"))
		})
		if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
		fmt.Println("Server closed")
	}()

	// Start GUI
	cmdArgs := []string{
		fmt.Sprintf("--app=%s", frontendUrl),
		// Chrome needs both --class & --user-dir to correctly
		// set taskbar icon:
		"--class=webmaker2000",
		"--user-dir=/tmp/webmaker2000",
	}
	fmt.Println("Browser starting")
	cmd := exec.Command("chromium", cmdArgs...)
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println("Browser closed")

	wg.Wait()
	fmt.Println("All closed")
}
