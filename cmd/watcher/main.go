package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/mattermost/mattermost-plugin-nagios/internal/watcher"
)

var (
	dir   = flag.String("dir", "/usr/local/nagios/etc/", "Nagios configuration files directory")
	url   = flag.String("url", "", "Mattermost Server address")
	token = flag.String("token", "", "Nagios plugin token")
)

func main() {
	flag.Parse()

	baseDir := *dir

	if !filepath.IsAbs(baseDir) {
		log.Fatal("dir argument must be an absolute path, like /usr/local/nagios/etc/")
	}

	allowedExtensions := []string{".cfg"}

	files, directories, err := watcher.GetAllInDirectory(baseDir, allowedExtensions)
	if err != nil {
		log.Fatalf("GetAllInDirectory: %v", err)
	}

	differential, err := watcher.NewDifferential(allowedExtensions, files, http.DefaultClient, *url, *token)
	if err != nil {
		log.Fatalf("NewDifferential: %v", err)
	}

	log.Printf("Initialized Differential watcher with %d files and %d directories", len(files), len(directories))

	done := make(chan struct{})

	go func() {
		defer close(done)

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		<-interrupt

		// We received an interrupt signal, shut down.
		log.Println("Bye")
	}()

	if err := watcher.WatchDirectories(directories, differential, done); err != nil {
		log.Panic(err)
	}
}
