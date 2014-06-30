package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"github.com/jmoiron/sqlx"
	"github.com/fritz0705/inventory"
)

type Config struct {
	Database string
	AssetsPath string
}

func handlerFactory(configFile string) (http.Handler, error) {
	config := &Config{
		Database: "inventory.db",
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

	var err error

	handler := inventory.NewApplication()
	handler.DB, err = sqlx.Open("sqlite3", config.Database)
	handler.DB.Exec("PRAGMA foreign_keys = on")

	return handler, err
}

func main() {
	var (
		flListen         = flag.String("listen", "localhost:8901", "server listen address")
		flConfig         = flag.String("config", "", "path to configuration file")
		flSsl            = flag.Bool("ssl", false, "enable ssl, requires certificate and key files")
		flCertFile       = flag.String("cert", "", "ssl certificate")
		flKeyFile        = flag.String("ssl-key", "", "ssl private key")
		flReadTimeout    = flag.Duration("read-timeout", 10*time.Second, "read timeout")
		flWriteTimeout   = flag.Duration("write-timeout", 10*time.Second, "write timeout")
		flMaxHeaderBytes = flag.Int("buffer", 1<<20, "maximum header bytes")
	)

	flag.Parse()

	handler, err := handlerFactory(*flConfig)
	if err != nil {
		log.Fatalf("An error occured while initializing the application: %s", err)
	}
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
}
