package main

import (
	// "encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

const DEFAULT_ADDR = "127.0.0.1:8000"

// ==========================================================================
//  Routes
// ==========================================================================

func main() {
	var site WebSite
	router := NewRouter()

	// Register routes
	// router.RegisterView("/", site.HomeView)
	router.RegisterStreamView("/", site.HomeView)
	router.RegisterView("/about/", site.AboutView)
	router.RegisterView("/slow/", site.SlowView)
	router.RegisterHandle("/assets/", http.StripPrefix("/assets/",
		http.FileServer(http.Dir("./assets"))),
	)

	runServer(router)
}

// ==========================================================================
//  Views
// ==========================================================================

type View func(*http.Request) string
type StreamView func(http.ResponseWriter, *http.Request)
type WebSite struct{}

func (site WebSite) HomeView(w http.ResponseWriter, r *http.Request) {
	const tpl = `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>{{.Title}}</title>
	<link type="text/css" rel="stylesheet" href="/assets/style.css">
</head>
<body>
	<h1>{{.Title}}</h1>
	<p>
		{{.Text}}
	</p>
	{{.ScriptSafe}}
	{{/* .Script */}}
</body>
</html>`

	t, err := template.New("home").Parse(tpl)

	data := struct {
		Title      string
		Text       string
		ScriptSafe string
		// Script template.HTML
	}{
		Title:      "Hello Home!",
		Text:       "Something, something, something, darkside...",
		ScriptSafe: "<script>alert('Hello from the whiteside!')</script>",
		// Script: template.HTML("<script>alert('Hello from the darkside!')</script>"),
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}
}

func (site WebSite) AboutView(r *http.Request) string {
	return "Hello About!"
}

func (site WebSite) SlowView(r *http.Request) string {
	// Simulate a long operation
	time.Sleep(950 * time.Millisecond)

	return "Hello Slow!"
}

// ==========================================================================
//  Router
// ==========================================================================

type Router struct {
	mux *http.ServeMux
}

func (router Router) GetHandler() http.Handler {
	return router.mux
}

func (router Router) RegisterView(urlPath string, view View) {
	router.mux.HandleFunc(urlPath, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != urlPath {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, view(r))
	})
}

func (router Router) RegisterStreamView(urlPath string, view StreamView) {
	router.mux.HandleFunc(urlPath, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != urlPath {
			http.NotFound(w, r)
			return
		}
		view(w, r)
	})
}

func (router Router) RegisterHandle(urlPath string, handler http.Handler) {
	router.mux.Handle(urlPath, handler)
}

// func (router Router) makeHandler(urlPath string, view View) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if r.URL.Path != urlPath {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		fmt.Fprint(w, view(r))
// 	}
// }

func NewRouter() Router {
	return Router{mux: http.NewServeMux()}
}

// ==========================================================================
//  Utils
// ==========================================================================

type ResponseWriterWrapper struct {
	http.ResponseWriter
	status int
}

func (r *ResponseWriterWrapper) Write(p []byte) (int, error) {
	return r.ResponseWriter.Write(p)
}

func (r *ResponseWriterWrapper) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func LoggerHandler(f http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writerWrapper := &ResponseWriterWrapper{
			ResponseWriter: w,
		}

		url := r.URL.String()

		f.ServeHTTP(writerWrapper, r)

		status := writerWrapper.status
		if status == 0 {
			status = http.StatusOK
		}

		log.Println(r.Method, url, status)
	}
}

func runServer(router Router) {
	addr := DEFAULT_ADDR

	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	server := &http.Server{
		Addr:    addr,
		Handler: LoggerHandler(router.GetHandler()),
		// ReadTimeout:    10 * time.Second,
		// WriteTimeout:   10 * time.Second,
		// MaxHeaderBytes: 1 << 20,
	}

	log.Println("Starting webserver on", addr, "...")
	log.Fatal(server.ListenAndServe())
}
