package web

//go:generate go run -mod=vendor .packr/packr.go

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"

	"github.com/gobuffalo/packr/v2"
	"github.com/sirupsen/logrus"
)

type GoTemplate struct {
	ApiVersion string
	Kind       string
	Schema     string
	Form       string
}

func RunWebServer(config Configuration) error {
	//Redirect requests from known locations to the embedded content from ./frontend
	box := packr.New("frontend", "../../frontend/dist")
	http.Handle("/bundle.js", http.FileServer(box))
	http.Handle("/fonts/", http.FileServer(box))
	http.Handle("/favicon.ico", http.FileServer(box))
	http.Handle("/health", checkHealth(box))

	returnIndex := func(writer http.ResponseWriter, reader *http.Request) {
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
			ApiVersion: config.ApiVersion(),
			Kind:       config.Kind(),
			Form:       string(formBytes),
			Schema:     string(schemaBytes),
		}
		if err := templates.ExecuteTemplate(writer, "template", goTemplate); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}

	sanitizeError := func(err error) string {
		errorMsg := strings.Replace(fmt.Sprint(err), "\"", "\\\"", -1)
		errorMsg = strings.Replace(errorMsg, "\n", " ", -1)
		return strings.Replace(errorMsg, "\t", "", -1)
	}

	processYaml := func(writer http.ResponseWriter, reader *http.Request) {
		body, err := ioutil.ReadAll(reader.Body)
		if err != nil {
			logrus.Errorf("Error reading message %v", err)
			http.Error(writer, "Error reading request", http.StatusInternalServerError)
		} else {
			request := string(body)
			//TODO temporary fixes to yaml, which should not be a problem to begin with:
			if strings.Contains(request, "\\n") {
				logrus.Infof("Request was %s", request)
				request = strings.Replace(request, "\\n", "\n", -1)
				logrus.Infof("Request is now %s", request)
			}
			if strings.HasPrefix(request, "\"") {
				logrus.Infof("Request was %s", request)
				request = strings.TrimLeft(request, "\"")
				logrus.Infof("Request is now %s", request)
			}
			if strings.HasSuffix(request, "\"") {
				logrus.Infof("Request was %s", request)
				request = strings.TrimRight(request, "\"")
				logrus.Infof("Request is now %s", request)
			}
			err := config.CallBack(request)
			writer.Header().Set("Content-Type", "application/json")
			if err != nil {
				logrus.Info("Unable to process the request: ", err)
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte(fmt.Sprintf("{\"result\": \"error\", \"message\": \"%v\"}", sanitizeError(err))))
			} else {
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte("{\"result\": \"success\"}"))
			}
		}
	}

	//For anything else:
	http.HandleFunc("/", func(writer http.ResponseWriter, reader *http.Request) {
		//For index.html and root GET requests, send back the processed index.html
		returnIndex(writer, reader)
	})

	http.HandleFunc("/api", func(writer http.ResponseWriter, reader *http.Request) {
		if reader.Method == "POST" {
			processYaml(writer, reader)
		} else {
			http.Error(writer, "Unable to handle request", http.StatusNotFound)
		}
	})

	http.HandleFunc("/api/form", func(writer http.ResponseWriter, reader *http.Request) {
		js, err := json.Marshal(config.Form())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		writer.Write(js)
	})

	http.HandleFunc("/api/schema", func(writer http.ResponseWriter, reader *http.Request) {
		js, err := json.Marshal(config.Schema())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		writer.Write(js)
	})

	http.HandleFunc("/api/spec", func(writer http.ResponseWriter, reader *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(writer, "{\"kind\": \"%v\", \"apiVersion\": \"%v\"}", config.Kind(), config.ApiVersion())
	})

	http.HandleFunc("/dev/js-version", func(writer http.ResponseWriter, reader *http.Request) {
		jsBuildHashString, err := box.Find("build-hash.json")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		writer.Write(jsBuildHashString)
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
