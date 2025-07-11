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

// Returns an HTTP handler function for rendering basic static pages.
//
// The title parameter is used both as the page title and to determine which
// page-specific boolean flag to set in PageData (IsCreditsPage, etc.). This
// boolean flag controls which HTML block is rendered in the template.
func basicPageHandler(title string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pageData := PageData{
			Title:        title,
			CanonicalUrl: getCanonicalUrl(r),
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
		}
		MainTemplate.Execute(w, pageData)
	}
}

// Processes both the homepage and search queries.
func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		serveNotFound(w)
		return
	}

	// Add build date header to homepage
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
		CanonicalUrl: getCanonicalUrl(r),
	}

	normalizedQuery := normalizeForSearch(query)
	if normalizedQuery != "" {
		entries, total := getEntries(normalizedQuery, searchMode, pageNumber, DefaultPageSize)
		pageData.PhrasesHtml = template.HTML(renderPhrases(entries, false))
		pageData.TotalPages = (total + DefaultPageSize - 1) / DefaultPageSize
		if pageNumber > 1 {
			pageData.PreviousPage = pageNumber - 1
		}
		if pageNumber < pageData.TotalPages {
			pageData.NextPage = pageNumber + 1
		}
	}

	MainTemplate.Execute(w, pageData)
}

// Processes letter pages (/lletra/{A-Z}).
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
		LetterHtml:   template.HTML(renderConceptsByLetter(ConceptsByFirstLetter[letter])),
		CanonicalUrl: getCanonicalUrl(r),
	}

	MainTemplate.Execute(w, pageData)
}

// Processes concept pages (/concepte/{conceptSlug}).
func conceptHandler(w http.ResponseWriter, r *http.Request) {
	entries := getEntriesByConceptSlug(r.PathValue("concept"))
	if len(entries) == 0 {
		serveNotFound(w)
		return
	}

	// Sort entries for this concept by accepció, antònim, and phrase
	collator := collate.New(language.Catalan)
	slices.SortFunc(entries, func(a, b Entry) int {
		// 1) Compare by the numbered acception from the concept
		comparison := collator.CompareString(a.AccepcioConcepte, b.AccepcioConcepte)
		if comparison != 0 {
			return comparison
		}

		// 2) Put antonyms at the end
		if a.AntonimConcepte != b.AntonimConcepte {
			if a.AntonimConcepte {
				return 1
			}
			return -1
		}

		// 3) Compare by phrase without parentheses content
		return collator.CompareString(a.TitleNormalizedWpc, b.TitleNormalizedWpc)
	})

	pageData := PageData{
		Title:         getConceptTitle(entries[0].Concepte),
		IsConceptPage: true,
		Concept:       template.HTML(getConceptTitleHtml(entries[0].Concepte)),
		PhrasesHtml:   template.HTML(renderPhrases(entries, true)),
		CanonicalUrl:  getCanonicalUrl(r),
	}

	MainTemplate.Execute(w, pageData)
}

// Serves a 404 Not Found page.
func serveNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	NotFoundTemplate.Execute(w, nil)
}
