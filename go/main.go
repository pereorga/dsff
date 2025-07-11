package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
)

const (
	ServerAddress        = ":8090"
	BaseCanonicalUrl     = "https://dsff.uab.cat"
	DefaultPageSize      = 10
	SearchModeConte      = "Conté"
	SearchModeComencaPer = "Comença per"
	SearchModeAcabaEn    = "Acaba en"
	SearchModeCoincident = "Coincident"
)

var BuildDate string

var (
	NotFoundTemplate *template.Template
	MainTemplate     *template.Template
)

//go:embed templates/*
var TemplateFS embed.FS

var (
	// Array of all dictionary entries
	AllEntries []Entry
	// Map of all phrases. Used to quickly check if a phrase exists
	PhrasesMap map[string]bool
	// Map of initial letters and their concepts. Used in letter pages
	ConceptsByFirstLetter map[string][]string
)

func main() {
	err := loadDataFromFile("data.json.gz")
	if err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}

	log.Printf("Loaded %d entries, covering %d initial letters.\n",
		len(AllEntries), len(ConceptsByFirstLetter))

	MainTemplate = template.Must(template.New("main.html").ParseFS(TemplateFS, "templates/main.html"))
	NotFoundTemplate = template.Must(template.New("404.html").ParseFS(TemplateFS, "templates/404.html"))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", searchHandler)
	mux.HandleFunc("GET /lletra/{letter}", letterHandler)
	mux.HandleFunc("GET /concepte/{concept}", conceptHandler)
	mux.HandleFunc("GET /abreviatures", basicPageHandler("Abreviatures"))
	mux.HandleFunc("GET /coneix", basicPageHandler("Coneix el diccionari"))
	mux.HandleFunc("GET /credits", basicPageHandler("Crèdits"))
	mux.HandleFunc("GET /presentacio", basicPageHandler("Presentació"))

	// Handle static files individually to avoid the default directory file
	// listing.
	mux.Handle("GET /main.min.css", http.FileServer(http.Dir("public/css/")))
	mux.Handle("GET /search.min.js", http.FileServer(http.Dir("public/js/")))
	mux.Handle("GET /by-nc-sa.svg", http.FileServer(http.Dir("public/img/")))
	mux.Handle("GET /uab.svg", http.FileServer(http.Dir("public/img/")))
	mux.Handle("GET /favicon.ico", http.FileServer(http.Dir("public/")))
	mux.Handle("GET /opensearch.xml", http.FileServer(http.Dir("public/")))
	mux.Handle("GET /robots.txt", http.FileServer(http.Dir("public/")))

	// The first website used to have the homepage duplicated at /cerca,
	// redirect it to the homepage.
	mux.HandleFunc("GET /cerca", func(w http.ResponseWriter, r *http.Request) {
		redirectURL := "/"
		if r.URL.RawQuery != "" {
			redirectURL = "/?" + r.URL.RawQuery
		}
		http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
	})

	log.Println("Server started at", ServerAddress)
	log.Fatal(http.ListenAndServe(ServerAddress, mux))
}
