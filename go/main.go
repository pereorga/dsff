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

// BuildDate is set at compile time to indicate when the binary was built.
var BuildDate string

var (
	// NotFoundTemplate is the parsed template for 404 error pages.
	NotFoundTemplate *template.Template
	// MainTemplate is the parsed template for the main page layout.
	MainTemplate *template.Template
)

//go:embed templates/*
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

	// Parse the HTML templates from the embedded filesystem.
	MainTemplate = template.Must(template.New("main.html").ParseFS(TemplateFS, "templates/main.html"))
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
	// TODO:
	//  - Set long cache headers for static assets (JS, CSS, images). But then:
	//  - Append cache-busting query strings or version hashes to CSS/JS URLs.
	// Also consider:
	//  - Enable compression (gzip and brotli) — this may be better handled at a
	//    higher layer, such as TLS termination or the reverse proxy. Mainly
	//    useful for the CSS and JS files, which are the only responses likely to
	//    exceed 100 KB.
	mux.Handle("GET /main.min.css", http.FileServer(http.Dir("public/css/")))
	mux.Handle("GET /search.min.js", http.FileServer(http.Dir("public/js/")))
	mux.Handle("GET /by-nc-sa.svg", http.FileServer(http.Dir("public/img/")))
	mux.Handle("GET /uab.svg", http.FileServer(http.Dir("public/img/")))
	mux.Handle("GET /favicon.ico", http.FileServer(http.Dir("public/")))
	mux.Handle("GET /opensearch.xml", http.FileServer(http.Dir("public/")))
	mux.Handle("GET /robots.txt", http.FileServer(http.Dir("public/")))

	// Handle legacy /cerca URL by redirecting to the homepage.
	// This ensures that old bookmarks and search engine links continue to work.
	mux.HandleFunc("GET /cerca", func(w http.ResponseWriter, r *http.Request) {
		redirectURL := "/"
		if r.URL.RawQuery != "" {
			redirectURL = "/?" + r.URL.RawQuery
		}
		http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
	})

	// Start the HTTP server.
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
