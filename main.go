package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"go.imnhan.com/webmaker2000/database"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

//go:embed frontend
var frontend embed.FS

const FileExt = "wm2k"
const FileDefaultName = "Site1." + FileExt

var AppState = struct {
	FilePath string `json:"filePath"`
	Route    string `json:"route"`
}{}

type Request struct {
	Proc string `json:"proc"`
}

type Response struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func main() {
	if len(os.Args) > 2 {
		fmt.Println("Usage: webmaker2000 Site1.wm2k")
		os.Exit(1)
	}

	if len(os.Args) == 2 {
		AppState.FilePath = os.Args[1]
		database.ConnectDb(AppState.FilePath)
	}

	// Snatch a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	frontendUrl := fmt.Sprintf("http://localhost:%d/frontend/", port)
	websocketUrl := fmt.Sprintf("ws://localhost:%d/websocket", port)
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

		// Basically a way to pass server-side constants to client:
		http.HandleFunc("/frontend/constants.js", func(w http.ResponseWriter, r *http.Request) {
			constants := []byte(fmt.Sprintf(`
const constants = {
    WEBSOCKET_URL: "%s",
};
export default constants;
`,
				websocketUrl,
			))
			w.Header().Add("Content-Type", "text/javascript")
			w.WriteHeader(200)
			w.Write(constants)
		})

		// RPC transport between client & server.
		// When either client or server exits, the other side notices that the
		// websocket is closed, then exits itself.
		http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
			c, err := websocket.Accept(w, r, nil)
			if err != nil {
				panic(err)
			}
			defer c.CloseNow()

			ctx := context.Background()

			var req Request
			for {
				err = wsjson.Read(ctx, c, &req)
				if websocket.CloseStatus(err) == websocket.StatusGoingAway {
					fmt.Println("Websocket closed by client. Closing server")
					srv.Shutdown(context.TODO())
					return
				}
				if err != nil {
					panic(fmt.Errorf("decode json: %w", err))
				}

				fmt.Printf("Received: %v\n", req)

				switch req.Proc {
				case "getstate":
					state, err := json.Marshal(Response{
						Type: "state",
						Data: AppState,
					})
					if err != nil {
						panic(state)
					}
					c.Write(ctx, websocket.MessageText, state)
				default:
					fmt.Println("Unexpected request:", req)
				}
			}
		})
		if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
		fmt.Println("Server closed")
	}()

	// Start GUI browser
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Browser starting")

		// The chrome/chromium command may or may not block: if there's already
		// a running Chrome process, running this command will not block,
		// otherwise it will. Therefore, we cannot use cmd to determine if
		// the browser is open or closed.
		// (we'll check whether the websocket is closed instead)
		cmd := exec.Command(
			"chromium",
			fmt.Sprintf("--app=%s", frontendUrl),
			// Chrome needs both --class & --user-dir to correctly
			// set taskbar icon:
			"--class=webmaker2000",
			"--user-dir=/tmp/webmaker2000",
		)
		err = cmd.Run()
		if err != nil {
			panic(err)
		}
	}()

	wg.Wait()
	fmt.Println("All closed")
}
