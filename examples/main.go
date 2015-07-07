package main

import (
	"net/http"
	"runtime"
	"text/template"
	"time"

	"github.com/joeybloggs/assets"
)

var indexTemplate *template.Template

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	time.Local = time.UTC // DO NOT REMOVE - NEEDED TO KEEP TIME IN UTC

	indexTemplate = template.Must(template.ParseFiles("views/index.html"))

	assetsConfig := &assets.Config{
		RunMode: assets.DevelopmentMode,
	}

	assets.Init(assetsConfig)

	http.Handle("/", http.HandlerFunc(render))
	http.ListenAndServe(":3001", nil)
}

func render(w http.ResponseWriter, r *http.Request) {
	indexTemplate.ExecuteTemplate(w, "master", nil)
}
