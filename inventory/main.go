package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path/filepath"
	"time"

	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"github.com/fritz0705/inventory"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	Database        string
	BasePath        string
	TemplatesPath   string
	AssetsPath      string
	AttachmentsPath string
	SessionKey      []byte
}

func handlerFactory(configFile string, basePath string) (http.Handler, error) {
	config := &Config{
		Database: "inventory.db",
		BasePath: basePath,
	}

	if configFile != "" {
		f, err := os.Open(configFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		decoder := json.NewDecoder(f)
		err = decoder.Decode(config)
		if err != nil {
			return nil, err
		}
	}

	if config.BasePath != "" {
		if config.TemplatesPath == "" {
			config.TemplatesPath = filepath.Join(config.BasePath, "templates")
		}
		if config.AssetsPath == "" {
			config.AssetsPath = filepath.Join(config.BasePath, "assets")
		}
		if config.AttachmentsPath == "" {
			config.AttachmentsPath = filepath.Join(config.BasePath, "attachments")
		}
	}

	if config.SessionKey != nil {
		inventory.SessionKey = config.SessionKey
	}

	var err error

	handler := inventory.NewApplication()
	handler.DB, err = sqlx.Open("sqlite3", config.Database)
	handler.DB.Exec("PRAGMA foreign_keys = on")
	if config.TemplatesPath != "" {
		handler.TemplatesPath = config.TemplatesPath
	}
	if config.AssetsPath != "" {
		handler.AssetsPath = config.AssetsPath
	}
	if config.AttachmentsPath != "" {
		handler.AttachmentStore = &inventory.FileAttachmentStore{config.AttachmentsPath}
	}

	handler.Init()

	return handler, err
}

func main() {
	var (
		flListen         = flag.String("listen", "localhost:8901", "server listen address")
		flMode           = flag.String("mode", "http", "server mode (http or fcgi)")
		flProtocol       = flag.String("protocol", "tcp", "listener protocol, only valid for fcgi mode")
		flConfig         = flag.String("config", "", "path to configuration file")
		flSsl            = flag.Bool("ssl", false, "enable ssl, requires certificate and key files")
		flCertFile       = flag.String("cert", "", "ssl certificate")
		flKeyFile        = flag.String("ssl-key", "", "ssl private key")
		flReadTimeout    = flag.Duration("read-timeout", 10*time.Second, "read timeout")
		flWriteTimeout   = flag.Duration("write-timeout", 10*time.Second, "write timeout")
		flMaxHeaderBytes = flag.Int("buffer", 1<<20, "maximum header bytes")
		flBase           = flag.String("base", "", "path to base directory")
	)

	flag.Parse()

	handler, err := handlerFactory(*flConfig, *flBase)
	if err != nil {
		log.Fatalf("An error occured while initializing the application: %s", err)
	}

	switch *flMode {
	case "http":
		s := &http.Server{
			Addr:           *flListen,
			Handler:        handler,
			ReadTimeout:    *flReadTimeout,
			WriteTimeout:   *flWriteTimeout,
			MaxHeaderBytes: *flMaxHeaderBytes,
		}

		if *flSsl {
			if *flCertFile == "" || *flKeyFile == "" {
				log.Fatalf("Requires SSL certificate and key files for SSL mode")
			}

			err = s.ListenAndServeTLS(*flCertFile, *flKeyFile)
		} else {
			err = s.ListenAndServe()
		}

		if err != nil {
			log.Fatal(err)
		}
	case "fcgi":
		listener, err := net.Listen(*flProtocol, *flListen)
		if err != nil {
			log.Fatal(err)
		}
		err = fcgi.Serve(listener, handler)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("Invalid server mode: %v (Valid modes are 'http' and 'fcgi')", *flMode)
	}
}
