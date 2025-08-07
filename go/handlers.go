package main

import (
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strconv"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// basicPageHandler returns an HTTP handler function for rendering basic static pages.
// It takes a title, which is used for both the page title and to set a corresponding
// boolean flag in the PageData struct. This flag determines which content block is
// rendered within the main template.
func basicPageHandler(title string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pageData := PageData{
			Title:        title,
			CanonicalURL: getCanonicalURL(r),
		}
		switch title {
		case "Crèdits":
			pageData.IsCreditsPage = true
		case "Coneix el diccionari":
			pageData.IsConeixPage = true
		case "Abreviatures":
			pageData.IsAbreviaturesPage = true
		case "Presentació":
			pageData.IsPresentacioPage = true
		default:
			// No-op
		}

		err := MainTemplate.Execute(w, pageData)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// searchHandler handles requests for the homepage and search queries.
// It processes the search query, search mode, and pagination from the URL parameters,
// retrieves the corresponding dictionary entries, and renders the results using the main template.
// If no query is provided, it displays the homepage.
//
// Additionally:
//   - Serves a 404 page for non-root paths
//   - Renders search results with proper pagination and sorting
//   - Page numbers are normalized (invalid values default to 1)
func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		serveNotFound(w)
		return
	}

	// Add build date header to the homepage for debugging and tracking purposes.
	if BuildDate != "" {
		w.Header().Set("X-Build-Date", BuildDate)
	}

	query := r.URL.Query().Get("frase")
	searchMode := r.URL.Query().Get("mode")
	pageNumberParam := r.URL.Query().Get("pagina")

	pageNumber := 1
	parsedPageNumber, err := strconv.Atoi(pageNumberParam)
	if err == nil && parsedPageNumber > 0 {
		pageNumber = parsedPageNumber
	}

	title := "Diccionari de Sinònims de Frases Fetes"
	if query != "" {
		title = fmt.Sprintf("Cerca «%s»", query)
	}

	pageData := PageData{
		IsHomepage:   true,
		SearchQuery:  query,
		SearchMode:   searchMode,
		SearchModes:  []string{SearchModeConte, SearchModeComencaPer, SearchModeAcabaEn, SearchModeCoincident},
		Title:        title,
		CurrentPage:  pageNumber,
		CanonicalURL: getCanonicalURL(r),
	}

	normalizedQuery := normalizeForSearch(query)
	if normalizedQuery != "" {
		entries, total := getEntries(normalizedQuery, searchMode, pageNumber, DefaultPageSize)
		pageData.PhrasesHTML = template.HTML(renderEntriesForSearch(entries))
		pageData.TotalPages = (total + DefaultPageSize - 1) / DefaultPageSize
		if pageNumber > 1 {
			pageData.PreviousPage = pageNumber - 1
		}
		if pageNumber < pageData.TotalPages {
			pageData.NextPage = pageNumber + 1
		}
	}

	err = MainTemplate.Execute(w, pageData)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// letterHandler handles requests for browsing dictionary entries by the first letter of a concept.
// It expects a URL path in the format /lletra/{letter}, where {letter} is a single uppercase letter (A-Z).
// If the letter is valid and has associated concepts, it renders a page with a list of those concepts.
//
// Additionally:
//   - Serves a 404 page for invalid letters or letters with no concepts
//   - Sorts concepts using the Catalan locale
func letterHandler(w http.ResponseWriter, r *http.Request) {
	letter := r.PathValue("letter")

	if len(letter) != 1 || letter[0] < 'A' || letter[0] > 'Z' {
		serveNotFound(w)
		return
	}

	if len(ConceptsByFirstLetter[letter]) == 0 {
		serveNotFound(w)
		return
	}

	pageData := PageData{
		Title:        fmt.Sprintf("Lletra %s", letter),
		IsLetterPage: true,
		Letter:       letter,
		LetterHTML:   template.HTML(renderConceptsByLetter(ConceptsByFirstLetter[letter])),
		CanonicalURL: getCanonicalURL(r),
	}

	err := MainTemplate.Execute(w, pageData)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// conceptHandler handles requests for displaying all phrases related to a specific concept.
// It expects a URL path in the format /concepte/{conceptSlug}, where {conceptSlug} is the
// URL-friendly version of the concept name. It retrieves all entries for that concept and
// renders them on a dedicated concept page.
//
// Additionally:
//   - Serves a 404 page if no entries found for the concept
//   - Sorts entries by accepció, antònim, and phrase
func conceptHandler(w http.ResponseWriter, r *http.Request) {
	entries := getEntriesByConceptSlug(r.PathValue("concept"))
	if len(entries) == 0 {
		serveNotFound(w)
		return
	}

	// Sort entries for this concept by accepció, antònim, and phrase.
	// This ensures a consistent and logical order for display.
	collator := collate.New(language.Catalan)
	slices.SortFunc(entries, func(a, b Entry) int {
		// 1) Compare by the numbered meaning from the concept.
		comparison := collator.CompareString(a.AccepcioConcepte, b.AccepcioConcepte)
		if comparison != 0 {
			return comparison
		}

		// 2) Put antonyms at the end.
		if a.AntonimConcepte != b.AntonimConcepte {
			if a.AntonimConcepte {
				return 1
			}
			return -1
		}

		// 3) Compare by phrase without parentheses content.
		return collator.CompareString(a.TitleNormalizedWpc, b.TitleNormalizedWpc)
	})

	pageData := PageData{
		Title:         getConceptTitle(entries[0].Concepte),
		IsConceptPage: true,
		Concept:       template.HTML(getConceptTitleHTML(entries[0].Concepte)),
		PhrasesHTML:   template.HTML(renderEntriesForConceptPage(entries)),
		CanonicalURL:  getCanonicalURL(r),
	}

	err := MainTemplate.Execute(w, pageData)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// serveNotFound renders a standard 404 Not Found error page.
func serveNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)

	err := NotFoundTemplate.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
