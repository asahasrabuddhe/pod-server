package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cip8/autoname"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	name := os.Getenv("NAME")
	if name == "" {
		logger.Warn("Name not set using environment variable. Generating a name.")
		name = autoname.Generate()
	}
	logger.Info("server name is " + name)

	handler := NewLoggingHandler(&PodServerHandler{
		name: name,
	}, logger)
	http.Handle("/", handler)

	_ = http.ListenAndServe(":8080", nil)
}

type PodServerHandler struct {
	name string
}

func (p *PodServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host

	response := map[string]string{
		"message":        fmt.Sprintf("%s got a request from %s", p.name, host),
		"request_method": r.Method,
		"request_path":   r.URL.Path,
	}

	env := r.URL.Query().Get("env")
	response["env"] = os.Getenv(env)

	queryMethod := r.URL.Query().Get("query_method")
	queryUrl := r.URL.Query().Get("query_url")
	if queryUrl != "" {
		if queryMethod == "" {
			queryMethod = http.MethodGet
		}
		req, err := http.NewRequest(queryMethod, queryUrl, nil)
		if err == nil {
			res, err := http.DefaultClient.Do(req)
			if err == nil {
				response["query_status"] = fmt.Sprintf("%d - %s", res.StatusCode, res.Status)
				body, _ := io.ReadAll(res.Body)
				response["query_response"] = string(body)
			}
			_ = res.Body.Close()
		}
	}

	encoder := json.NewEncoder(w)
	_ = encoder.Encode(response)
}
