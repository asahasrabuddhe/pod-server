package main

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/urfave/negroni"
	"go.ajitem.com/realip"
	"go.ajitem.com/zapdriver"
	"go.uber.org/zap"
)

type LoggingHandler struct {
	handler http.Handler
	logger  *zap.Logger
}

func NewLoggingHandler(handler http.Handler, logger *zap.Logger) http.Handler {
	return &LoggingHandler{
		handler: handler,
		logger:  logger,
	}
}

func (l *LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	lrw := negroni.NewResponseWriter(w)
	l.handler.ServeHTTP(lrw, r)

	payload := zapdriver.NewHTTP(r, nil)
	payload.Status = lrw.Status()
	payload.ResponseSize = strconv.Itoa(lrw.Size())

	ip, err := IPAddress()
	if err == nil {
		payload.ServerIP = ip.String()
	}

	payload.RemoteIP = realip.FromRequest(r)

	payload.Latency = time.Since(now).String()

	l.logger.Info("", zapdriver.HTTP(payload))
}

func IPAddress() (net.IP, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, address := range addresses {
		if ipNet, ok := address.(*net.IPNet); ok && ipNet.IP.IsGlobalUnicast() {
			if ipNet.IP.To4() != nil && ipNet.IP.To16() != nil {
				return ipNet.IP, nil
			}
		}
	}

	return nil, errors.New("ip not assigned")
}
