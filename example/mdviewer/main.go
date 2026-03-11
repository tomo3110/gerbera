package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	g "github.com/tomo3110/gerbera"
	gl "github.com/tomo3110/gerbera/live"
)

func main() {
	addr := flag.String("addr", ":8860", "listen address")
	preview := flag.Bool("preview", false, "preview-only mode (requires file argument)")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var filePath string
	if flag.NArg() > 0 {
		fp, err := filepath.Abs(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", fp)
			os.Exit(1)
		}
		filePath = fp
	}

	if *preview && filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: -preview requires a file argument")
		os.Exit(1)
	}

	var opts []gl.Option
	opts = append(opts, gl.WithLang("en"))
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	factory := func(_ context.Context) gl.View {
		return &MarkdownView{
			FilePath: filePath,
			Preview:  *preview,
		}
	}

	http.Handle("/", gl.Handler(factory, opts...))
	log.Printf("mdviewer running on %s", *addr)
	if filePath != "" {
		mode := "editor"
		if *preview {
			mode = "preview"
		}
		log.Printf("  mode: %s, file: %s", mode, filePath)
	}
	log.Fatal(http.ListenAndServe(*addr, g.Serve(http.DefaultServeMux)))
}
