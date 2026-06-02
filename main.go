package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/gnitoahc/md-html/converter"
	"github.com/gnitoahc/md-html/tmpl"
	"golang.org/x/term"
)

var (
	version    bool
	outputfile string
	watch      bool
	port       string
)

// watchmd watches for changes in the markdown file
func watchmd(inputfile string, update chan<- string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(inputfile)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) {
				update <- "update"
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

// getContent reads the content of a file
func getContent(filename string) ([]byte, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// sseHub broadcasts reload signals to connected browser clients.
type sseHub struct {
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

func newSSEHub() *sseHub {
	return &sseHub{clients: make(map[chan struct{}]struct{})}
}

func (h *sseHub) subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *sseHub) unsubscribe(ch chan struct{}) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
}

func (h *sseHub) broadcast() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// write writes the content of a markdown file to an HTML file
func write(inputfile, outputfile string, liveReload bool) {
	md, err := getContent(inputfile)
	html, err := converter.Md2HTML(md)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	outputTempl := tmpl.GetTemplate()
	tmpl := template.Must(template.New("output").Parse(outputTempl))

	f, err := os.Create(outputfile)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	err = tmpl.Execute(f, map[string]any{
		"Content":    template.HTML(html),
		"LiveReload": liveReload,
	})
	if err != nil {
		log.Fatal(err.Error())
		return
	}
}

func init() {
	flag.BoolVar(&version, "v", false, "Version of the program")
	flag.BoolVar(&watch, "w", false, "Watch for changes in the markdown file")
	flag.StringVar(&outputfile, "o", "", "Name of the output file")
	flag.StringVar(&port, "p", "8080", "Port to serve the HTML file")

	flag.Usage = func() {
		fmt.Println("Usage: md-html [options] <filename>")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	inputfile := flag.Arg(0)

	if version {
		fmt.Println("md-html", Version)
		return
	}

	// Get output filename
	if outputfile == "" {
		slices := strings.Split(inputfile, ".")
		outputfile = strings.Join(slices[:len(slices)-1], "") + ".html"
	}

	// Initial write
	write(inputfile, outputfile, watch)

	if !watch {
		fmt.Println("Output file: ", outputfile)
		return
	}

	hub := newSSEHub()
	update := make(chan string)
	go watchmd(inputfile, update)

	go func() {
		for {
			<-update
			write(inputfile, outputfile, true)
			hub.broadcast()
			fmt.Print("\r[md-html] Page reloaded.\r\n")
		}
	}()

	fs := http.FileServer(http.Dir(".")) // Serve files from the current directory
	http.Handle("/", fs)

	http.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := hub.subscribe()
		defer hub.unsubscribe(ch)

		for {
			select {
			case <-ch:
				fmt.Fprintf(w, "data: reload\n\n")
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			case <-r.Context().Done():
				return
			}
		}
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	oldState, rawErr := term.MakeRaw(int(os.Stdin.Fd()))
	if rawErr == nil {
		go func() {
			buf := make([]byte, 1)
			for {
				n, err := os.Stdin.Read(buf)
				if err != nil || n == 0 {
					return
				}
				switch buf[0] {
				case 'r', 'R':
					select {
					case update <- "reload":
					default:
					}
				case 'q', 'Q', 3: // 3 = Ctrl+C
					quit <- syscall.SIGTERM
					return
				}
			}
		}()
		fmt.Printf("\rServing at http://localhost:%s/%s\r\nPress 'r' to reload, 'q' to quit.\r\n", port, outputfile)
	} else {
		fmt.Printf("Visit the following URL: http://localhost:%s/%s, or simply refresh the HTML manually.\n", port, outputfile)
	}

	go func() {
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}()

	<-quit

	if rawErr == nil {
		term.Restore(int(os.Stdin.Fd()), oldState)
	}

	fmt.Println("\nShutting down, removing", outputfile)
	if err := os.Remove(outputfile); err != nil {
		log.Println("Failed to remove output file:", err)
	}
}
