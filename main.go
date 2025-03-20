package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gnitoahc/md-html/converter"
	"github.com/gnitoahc/md-html/tmpl"
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

// write writes the content of a markdown file to an HTML file
func write(inputfile, outputfile string) {
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

	err = tmpl.Execute(f, map[string]interface{}{
		"Content": template.HTML(html),
	})
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	return
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
	write(inputfile, outputfile)

	if !watch {
		fmt.Println("Output file: ", outputfile)
		return
	}

	update := make(chan string)
	go watchmd(inputfile, update)

	go func() {
		for {
			<-update
			write(inputfile, outputfile)
		}
	}()

	fs := http.FileServer(http.Dir(".")) // Serve files from the current directory
	http.Handle("/", fs)

	fmt.Printf("Visit the following URL: http://localhost:8080/%s, or simply refresh the HTML manually.", outputfile)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
