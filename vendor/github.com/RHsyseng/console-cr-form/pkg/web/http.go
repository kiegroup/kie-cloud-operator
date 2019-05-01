package web

//go:generate go run .packr/packr.go

import (
	"encoding/json"
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"text/template"
)

type GoTemplate struct {
	Schema string
	Form   string
}

func RunWebServer(config Configuration) error {
	//Redirect requests from known locations to the embedded content from ./frontend
	box := packr.New("frontend", "../../frontend/dist")
	http.Handle("/bundle.js", http.FileServer(box))
	http.Handle("/fonts/", http.FileServer(box))
	http.Handle("/favicon.ico", http.FileServer(box))
	http.Handle("/health", checkHealth(box))
	logrus.SetLevel(logrus.DebugLevel)

	returnIndex := func(writer http.ResponseWriter, reader *http.Request){
		templateString, err := box.FindString("index.html")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		templates := template.Must(template.New("template").Parse(templateString))

		formBytes, err := json.Marshal(config.Form())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		schemaBytes, err := json.Marshal(config.Schema())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		goTemplate := GoTemplate{
			Form:   string(formBytes),
			Schema: string(schemaBytes),
		}
		if err := templates.ExecuteTemplate(writer, "template", goTemplate); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}

	processYaml := func(writer http.ResponseWriter, reader *http.Request) {
		body, err := ioutil.ReadAll(reader.Body)
		if err != nil {
			logrus.Errorf("Error reading message %v", err)
			http.Error(writer, "Error reading request", http.StatusInternalServerError)
		} else {
			request := string(body)
			logrus.Debugf("Request is %v", request)
			config.Apply(request)
			writer.WriteHeader(http.StatusOK)
		}
	}

	//For anything else:
	http.HandleFunc("/", func(writer http.ResponseWriter, reader *http.Request) {
		if reader.Method == "POST" {
			//Receive and handle posted yaml separately
			processYaml(writer, reader)
		} else {
			//For index.html and root GET requests, send back the processed index.html
			returnIndex(writer, reader)
		}
	})

	//Start the web server, set the port to listen to 8080. Without a path it assumes localhost
	listenAddr := fmt.Sprintf("%s:%d", config.Host(), config.Port())
	logrus.Info("Will listen on ", listenAddr)
	err := http.ListenAndServe(listenAddr, nil)
	return err
}

func checkHealth(box *packr.Box) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, reader *http.Request) {
		responseStatus := http.StatusNoContent
		required := []string{"index.html", "bundle.js"}
		for _, content := range required {
			if !box.Has(content) {
				logrus.Warnf("Packr box missing %s", content)
				responseStatus = http.StatusFailedDependency
			}
		}
		writer.WriteHeader(responseStatus)
	})
}
