// Package main implements a web server for the Diccionari de Sinònims de Frases Fetes.
//
// The server is responsible for the following:
//   - Loading dictionary data from a gzipped JSON file.
//   - Parsing HTML templates for rendering web pages.
//   - Handling HTTP requests for search, letter, and concept pages.
//   - Serving static assets such as CSS, JavaScript, and images.
//   - Redirecting legacy URLs to their new counterparts.
package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"time"
)

const (
	BaseCanonicalURL     = "https://dsff.uab.cat"
	DefaultPageSize      = 10
	SearchModeConte      = "Conté"
	SearchModeComencaPer = "Comença per"
	SearchModeAcabaEn    = "Acaba en"
	SearchModeCoincident = "Coincident"
)

// AssetVersion is a content hash of the minified CSS and JavaScript assets.
// It is appended to asset URLs to invalidate browser caches when they change.
var AssetVersion string

var (
	NotFoundTemplate *template.Template
	MainTemplate     *template.Template
)

//go:embed templates/*.html templates/pages/*.html
var TemplateFS embed.FS

var (
	// AllEntries contains all dictionary entries loaded from the data file.
	AllEntries []Entry
	// PhrasesMap maps phrases to their existence for quick lookup.
	PhrasesMap map[string]bool
	// ConceptsByFirstLetter maps initial letters to their associated concepts.
	ConceptsByFirstLetter map[string][]string
)

func main() {
	// Load the dictionary data from the gzipped JSON file.
	// This populates the AllEntries, PhrasesMap, and ConceptsByFirstLetter variables.
	err := loadDataFromFile("data.json.gz")
	if err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}

	log.Printf("Loaded %d entries, covering %d initial letters.\n",
		len(AllEntries), len(ConceptsByFirstLetter))

	// Hash the minified assets so asset URLs change only when their contents do.
	AssetVersion = computeAssetVersion("public/css/main.min.css", "public/js/search.min.js")

	// Parse the HTML templates from the embedded filesystem.
	// main.html is the layout; templates/pages/*.html define static page partials.
	MainTemplate = template.Must(template.New("main.html").ParseFS(
		TemplateFS,
		"templates/main.html",
		"templates/pages/*.html",
	))
	NotFoundTemplate = template.Must(template.New("404.html").ParseFS(TemplateFS, "templates/404.html"))

	// Create a new ServeMux to handle HTTP requests.
	mux := http.NewServeMux()

	// Register handlers for the main application routes.
	mux.HandleFunc("GET /", searchHandler)
	mux.HandleFunc("GET /lletra/{letter}", letterHandler)
	mux.HandleFunc("GET /concepte/{concept}", conceptHandler)
	mux.HandleFunc("GET /abreviatures", basicPageHandler("Abreviatures"))
	mux.HandleFunc("GET /coneix", basicPageHandler("Coneix el diccionari"))
	mux.HandleFunc("GET /credits", basicPageHandler("Crèdits"))
	mux.HandleFunc("GET /presentacio", basicPageHandler("Presentació"))

	// Register handlers for serving static files.
	// These are handled individually to avoid showing the annoying default
	// directory file listing.
	//
	// Include the build date in minified asset URLs to invalidate browser caches
	// when a new binary is deployed.
	mux.HandleFunc("GET /css/main.min.css", precompressedFileHandler("public/css/main.min.css", "text/css"))
	mux.HandleFunc("GET /js/search.min.js", precompressedFileHandler("public/js/search.min.js", "application/javascript"))
	cachedPublicFiles := cachedFileServer("public/")
	publicFiles := http.FileServer(http.Dir("public/"))
	mux.Handle("GET /img/by-nc-sa.svg", cachedPublicFiles)
	mux.Handle("GET /img/uab.svg", cachedPublicFiles)
	mux.Handle("GET /favicon.ico", cachedPublicFiles)
	mux.Handle("GET /opensearch.xml", publicFiles)
	mux.Handle("GET /robots.txt", publicFiles)

	// Handle legacy /cerca URL by redirecting to the homepage.
	// This ensures that old bookmarks and search engine links continue to work.
	mux.HandleFunc("GET /cerca", func(w http.ResponseWriter, r *http.Request) {
		redirectURL := "/"
		if r.URL.RawQuery != "" {
			redirectURL = "/?" + r.URL.RawQuery
		}
		http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
	})

	serverAddress := getServerAddress()
	server := &http.Server{
		Addr:         serverAddress,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Println("Server started at", serverAddress)
	log.Fatal(server.ListenAndServe())
}
