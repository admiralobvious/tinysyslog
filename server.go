package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/admiralobvious/tinysyslog/config"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mcuadros/go-syslog.v2"
)

// Server holds the config
type Server struct {
	config *config.Config
}

// NewServer creates a Server instance
func NewServer(cnf *config.Config) *Server {
	server := Server{
		config: cnf,
	}
	return &server
}

// Run runs the server
func (s *Server) Run(_ []string) error {
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.RFC5424)
	server.SetHandler(handler)

	switch strings.ToLower(s.config.SocketType) {
	case "tcp":
		if err := server.ListenTCP(s.config.Address); err != nil {
			log.Fatalln(err)
		}
	case "udp":
		if err := server.ListenUDP(s.config.Address); err != nil {
			log.Fatalln(err)
		}
	default:
		if err := server.ListenTCP(s.config.Address); err != nil {
			log.Fatalln(err)
		}
		if err := server.ListenUDP(s.config.Address); err != nil {
			log.Fatalln(err)
		}
	}

	server.Boot()
	log.Infof("TinySyslog listening on %s", s.config.Address)

	sink := SinkFactory(s.config)

	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {
			t := logParts["timestamp"].(time.Time)
			formatted := fmt.Sprintf("%s %s %s[%s]: %s",
				t.Format("Jan _2 15:04:05"),
				logParts["hostname"],
				logParts["app_name"],
				logParts["proc_id"],
				logParts["message"])

			log.Debugln(formatted)
			if err := sink.Write([]byte(formatted + "\n")); err != nil {
				log.Errorln(err)
			}
		}
	}(channel)

	server.Wait()
	return nil
}
