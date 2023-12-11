package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nullseed/logruseq"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func main() {
	f, err := os.Open("config.yml")
	if err != nil {
		fmt.Println(err)
	}
	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	log.AddHook(logruseq.NewSeqHook(cfg.Seq.Url, logruseq.OptionAPIKey(cfg.Seq.ApiKey)))
	log.Info("Starting Websocket ...")

	flag.Parse()
	hub := newHub()
	go hub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	addr := flag.String("addr", strings.Join([]string{cfg.Server.Host, cfg.Server.Port}, ":"), "http service address")
	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	err = server.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		log.Error("server closed")
	} else if err != nil {
		log.WithField("error", err).Error("error starting server")
		os.Exit(1)
	}
}
