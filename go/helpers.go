package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// precompressedFileHandler serves pre-compressed .br or .gz files when the client accepts those encodings.
// This is more efficient than runtime compression, especially for static files.
func precompressedFileHandler(originalPath, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Vary", "Accept-Encoding")
		acceptEncoding := r.Header.Get("Accept-Encoding")

		// Prefer Brotli if supported
		if strings.Contains(acceptEncoding, "br") {
			brotliPath := originalPath + ".br"
			_, err := os.Stat(brotliPath)
			if err == nil {
				w.Header().Set("Content-Encoding", "br")
				http.ServeFile(w, r, brotliPath)
				return
			}
		}

		// Fall back to gzip if supported
		if strings.Contains(acceptEncoding, "gzip") {
			gzipPath := originalPath + ".gz"
			_, err := os.Stat(gzipPath)
			if err == nil {
				w.Header().Set("Content-Encoding", "gzip")
				http.ServeFile(w, r, gzipPath)
				return
			}
		}

		// Fall back to serving the original uncompressed file
		http.ServeFile(w, r, originalPath)
	}
}

// getServerAddress returns the server address from the PORT env variable.
func getServerAddress() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	return ":" + port
}

// getAllAbbreviations returns a map of all abbreviations and their corresponding full text.
// This map is used to expand abbreviations found in the dictionary data.
// Note: Some abbreviations might be substrings of longer words, which could lead to
// false positives. However, this is currently only used in fields where this is not an issue.
func getAllAbbreviations() map[string]string {
	// This could cause false positives in replaceAbbreviations() if the
	// abbreviation is a substring of a longer word. For example, sentences
	// ending with words ending with "ant", "fam", "met", or the word "pop".
	// But should not be a problem at the moment because it is only used in
	// MarcatgeDialectal field.
	return map[string]string{
		"ant.":          "antonímia",
		"aprox.":        "aproximadament",
		"Bal.":          "Balears i baleàric",
		"Barc.":         "Barcelona",
		"Camp de Tarr.": "Camp de Tarragona",
		"Cast.":         "Castelló",
		"Cat.":          "Catalunya",
		"Eiv.":          "Eivissa",
		"Emp.":          "Empordà",
		"esp.":          "especialment",
		"euf.":          "eufemisme",
		"fam.":          "familiar, familiarment",
		"fig.":          "valor figurat del concepte",
		"Gir.":          "Girona",
		"interj.":       "interjecció",
		"inv.":          "inversió",
		"iròn.":         "[emprat] irònicament",
		"Mall.":         "Mallorca i mallorquí",
		"Men.":          "Menorca i menorquí",
		"met.":          "metàfora, metafòric",
		"Occ.":          "català (nord)occidental",
		"Or.":           "català oriental (català central)",
		"p.ext.":        "per extensió",
		"per ex.":       "per exemple",
		"Pir-or.":       "pirinenc-oriental",
		"pop.":          "[cançó] popular",
		"Ross.":         "Rosselló",
		"Tarr.":         "Tarragona",
		"v.f.":          "variant formal",
		"Val.":          "València i valencià",
		"vg.":           "vegeu",
	}
}

// getAllSources returns a map of all source abbreviations and their full text.
// This map is used to expand source citations found in the dictionary data.
func getAllSources() map[string]string {
	return map[string]string{
		"*":     "no prové de cap obra lexicogràfica",
		"A-M":   "Alcover, A. M. - F. de B. Moll, Diccionari Català-Valencià-Balear",
		"B":     "Balbastre, J., Nou Recull de Modismes i Frases Fetes. Català-castellà / castellà-català",
		"DIEC1": "Institut d'Estudis Catalans, Diccionari de la Llengua Catalana",
		"EC":    "Enciclopèdia Catalana, Diccionaris",
		"ECe":   "Enciclopèdia Catalana i Universitat Politècnica de Catalunya, Diccionari d'Economia i Gestió",
		"F":     "Fabra, P., Diccionari General de la Llengua Catalana",
		"Fr":    "Franquesa, M., Diccionari de Sinònims",
		"GEC":   "Gran Enciclopèdia Catalana",
		"P":     "Peris, A., Diccionari de Locucions i Frases Llatines",
		"PDL":   "Institut d'Estudis Catalans, Portal de Dades Lingüístiques",
		"R-M":   "Raspall, J. - J. Martí, Diccionari de Locucions i de Frases Fetes",
		"R":     "Riera Jaume, A., Així Xerram a Mallorca",
		"SP":    "Perramón, S., Proverbis, Dites i Frases Fetes de la Llengua Catalana",
		"T":     "Termcat",
	}
}

// getObservationSources returns a map of source abbreviations used specifically
// within the "Observacions" field and their corresponding full text.
func getObservationSources() map[string]string {
	return map[string]string{
		"DIEC1": "Institut d'Estudis Catalans, Diccionari de la Llengua Catalana",
	}
}

// getCategory returns the HTML representation of a grammatical category.
// It takes a category key (e.g., "sv") and returns an HTML string with an
// <abbr> tag that provides the full category name on hover.
//
// Postconditions:
//   - Returns formatted HTML <abbr> tag for recognized categories
//   - Returns original categoryKey for unrecognized categories
func getCategory(categoryKey string) string {
	categories := map[string]string{
		"o":      "O",
		"sa":     "SA",
		"sadv":   "SAdv",
		"sconj":  "SConj",
		"scoord": "SCoord",
		"sd":     "SD",
		"sn":     "SN",
		"sp":     "SP",
		"sq":     "SQ",
		"sv":     "SV",
	}
	categoriesAbbr := map[string]string{
		"o":      "oració",
		"sa":     "sintagma adjectival",
		"sadv":   "sintagma adverbial",
		"sconj":  "sintagma conjuntiu",
		"scoord": "sintagma coordinat",
		"sd":     "sintagma determinant",
		"sn":     "sintagma nominal",
		"sp":     "sintagma preposicional",
		"sq":     "sintagma quantificador",
		"sv":     "sintagma verbal",
	}

	category := categories[categoryKey]
	categoryTitle := categoriesAbbr[categoryKey]

	if category == "" || categoryTitle == "" {
		return categoryKey
	}

	return fmt.Sprintf("<em><abbr title=\"%s\">%s</abbr></em>", categoryTitle, category)
}

// loadDataFromFile loads and processes the dictionary data from a gzipped JSON file.
// It populates the global variables AllEntries, PhrasesMap, and ConceptsByFirstLetter,
// which are used throughout the application. This function is called once at startup.
func loadDataFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open data file %s: %w", filePath, err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	err = json.NewDecoder(gzipReader).Decode(&AllEntries)
	if err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	PhrasesMap = make(map[string]bool, len(AllEntries))
	ConceptsByFirstLetter = make(map[string][]string)

	// Populate data structures for efficient lookups.
	for _, entry := range AllEntries {
		PhrasesMap[removeParenthesesContent(entry.Title)] = true

		// Group concepts by their first letter for alphabetical browsing.
		firstRune := []rune(entry.Concepte)[0]
		key := strings.ToUpper(toLowercaseNoAccents(string(firstRune)))

		// Add the concept to the list for its corresponding letter, avoiding duplicates.
		if !slices.Contains(ConceptsByFirstLetter[key], entry.Concepte) {
			ConceptsByFirstLetter[key] = append(ConceptsByFirstLetter[key], entry.Concepte)
		}
	}

	// Sort the concepts within each letter group alphabetically.
	collator := collate.New(language.Catalan)
	for _, conceptList := range ConceptsByFirstLetter {
		slices.SortFunc(conceptList, collator.CompareString)
	}

	return nil
}

// getCanonicalURL returns the canonical URL for a given request.
// This is used to generate <link rel="canonical"> tags, which helps prevent
// search engines from indexing duplicate content from development or staging environments.
func getCanonicalURL(r *http.Request) string {
	canonical := BaseCanonicalURL + r.URL.EscapedPath()

	// For search results (on the root path), include the mode and frase query parameters.
	if r.URL.Path == "/" || r.URL.Path == "" {
		params := url.Values{}
		mode := r.URL.Query().Get("mode")
		if mode != "" {
			params.Set("mode", mode)
		}
		frase := r.URL.Query().Get("frase")
		if frase != "" {
			params.Set("frase", frase)
		}

		if len(params) > 0 {
			canonical += "?" + params.Encode()
		}
	}

	return canonical
}

// createAbbrReplacer creates a strings.Replacer to replace abbreviations with <abbr> tags.
func createAbbrReplacer(abbrMap map[string]string) *strings.Replacer {
	var replacements []string
	for key, value := range abbrMap {
		replacements = append(replacements, key, fmt.Sprintf("<abbr title=\"%s\">%s</abbr>", value, key))
	}
	return strings.NewReplacer(replacements...)
}

// createAbbrReplacerInParentheses creates a strings.Replacer for abbreviations enclosed in parentheses.
func createAbbrReplacerInParentheses(abbrMap map[string]string) *strings.Replacer {
	var replacements []string
	for key, value := range abbrMap {
		pattern := "(" + key + ")"
		replacement := fmt.Sprintf("(<abbr title=\"%s\">%s</abbr>)", value, key)
		replacements = append(replacements, pattern, replacement)
	}
	return strings.NewReplacer(replacements...)
}

// replaceAbbreviationsParentheses replaces abbreviations that are enclosed in parentheses.
// For example, it transforms "(v.f.)" into "(<abbr title=\"...\">v.f.</abbr>)".
func replaceAbbreviationsParentheses(text string) string {
	return createAbbrReplacerInParentheses(getAllAbbreviations()).Replace(text)
}

// replaceAbbreviations replaces abbreviations that are not necessarily in parentheses.
// This function is used when more selective replacement is not possible, but it carries
// a higher risk of making unintended replacements.
func replaceAbbreviations(text string) string {
	return createAbbrReplacer(getAllAbbreviations()).Replace(text)
}

// replaceSourceAbbreviationsParentheses replaces source abbreviations that are enclosed in parentheses.
// For example, it transforms "(DIEC1)" into "(<abbr title=\"...\">DIEC1</abbr>)".
func replaceSourceAbbreviationsParentheses(text string) string {
	return createAbbrReplacerInParentheses(getAllSources()).Replace(text)
}

// replaceObservationsSourceAbbreviations replaces source abbreviations for the "Observacions" field.
// This is similar to replaceAbbreviations but uses a specific set of sources.
func replaceObservationsSourceAbbreviations(text string) string {
	return createAbbrReplacer(getObservationSources()).Replace(text)
}

// getSources formats a comma-separated string of source abbreviations into an HTML string.
// Each source is wrapped in an <abbr> tag with its full name as the title.
// The entire string is enclosed in parentheses.
func getSources(sources string) string {
	// Remove parentheses
	cleanedSources := strings.ReplaceAll(sources, "(", "")
	cleanedSources = strings.ReplaceAll(cleanedSources, ")", "")
	cleanedSources = strings.TrimSpace(cleanedSources)

	if cleanedSources == "" {
		return ""
	}

	allSources := getAllSources()

	sourcesList := strings.Split(cleanedSources, ",")
	var formattedSources []string

	for _, source := range sourcesList {
		source = strings.TrimSpace(source)
		fullForm, exists := allSources[source]
		if exists {
			formattedSources = append(formattedSources,
				fmt.Sprintf("<abbr title=\"%s\">%s</abbr>", fullForm, source),
			)
		} else {
			// Not found in the map, just keep the raw text
			formattedSources = append(formattedSources, source)
		}
	}

	joinedSources := strings.Join(formattedSources, ",&nbsp;")
	return fmt.Sprintf("(%s)", joinedSources)
}

// getPhrase formats a single phrase for display, rendering it in bold.
func getPhrase(phrase string) string {
	return renderBoldPhrases(phrase, true)
}

// getNewIncorporationPhrase formats a new phrase, adding a marker and rendering it in bold.
func getNewIncorporationPhrase(phrase string) string {
	return "■ " + getPhrase(phrase)
}

// phraseExists checks if a given phrase exists in the dictionary.
// It uses the PhrasesMap for efficient lookup.
func phraseExists(phrase string) bool {
	return PhrasesMap[removeParenthesesContent(phrase)]
}

// smartSplit splits a string by a separator, but ignores separators that are inside parentheses.
// This is useful for splitting lists of phrases where some phrases may contain commas.
//
// Postconditions:
//   - Returns empty slice if input is empty
//   - Returns slice with at least one element for non-empty input
//   - Parentheses nesting is correctly handled
//   - Original separators inside parentheses are restored
func smartSplit(input, separator string) []string {
	const placeholderUnusedChar = "|"

	var processedBuilder strings.Builder
	processedBuilder.Grow(len(input))
	var parenthesesDepth int

	for _, char := range input {
		if char == '(' {
			parenthesesDepth++
		} else if char == ')' {
			parenthesesDepth--
		} else if string(char) == separator && parenthesesDepth > 0 {
			processedBuilder.WriteString(placeholderUnusedChar)
			continue
		}
		processedBuilder.WriteRune(char)
	}

	parts := strings.Split(processedBuilder.String(), separator)
	restoredParts := make([]string, len(parts))

	for i, part := range parts {
		restoredParts[i] = strings.ReplaceAll(strings.TrimSpace(part), placeholderUnusedChar, separator)
	}

	return restoredParts
}

// Phrases that should not be split or linked
var PhrasesWhitelist = []string{
	"Jesús, Maria i Josep (v.f.)",
	"en Pere, en Pau i en Berenguera (v.f.)",
	"córrer la Seca, la Meca i la vall d'Andorra (v.f.)",
}

// renderBoldPhrases renders one or more phrases in bold.
// If createLink is true, it also wraps each phrase in an anchor tag that links to a search for that phrase.
// It handles single phrases, as well as lists of phrases separated by commas or semicolons.
func renderBoldPhrases(input string, createLink bool) string {
	const placeholderUnusedChar = "|"

	if input == "" {
		return ""
	}

	// By default, assume input can be multiple phrases separated by a comma
	separator := ","
	var isSinglePhrase bool

	if phraseExists(input) || slices.Contains(PhrasesWhitelist, input) {
		// If the provided input exists as a phrase, don't try to split it.
		// Use a placeholder that won't be in the input, so the sentence is not
		// split but still gets processed correctly.
		isSinglePhrase = true
		separator = placeholderUnusedChar
	} else if strings.Contains(input, ";") {
		// ";" is used as a separator in the CMS when at least 1 phrase
		// contains commas.
		separator = ";"
	}

	phraseList := smartSplit(input, separator)
	for i, phrase := range phraseList {
		isFormalVariant := strings.Contains(phrase, " (v.f.)")
		shouldCreateLink := createLink && !isFormalVariant && phraseExists(phrase)

		phraseHTML := fmt.Sprintf("<strong>%s</strong>", phrase)
		if shouldCreateLink {
			searchPath := "/?mode=Conté&frase=" + url.QueryEscape(removeParenthesesContent(phrase))
			phraseHTML = fmt.Sprintf("<a href=\"%s\" rel=\"nofollow\">%s</a>", searchPath, phraseHTML)
		}

		// Make parentheses non-bold. This should not leave
		// inconsistent/unclosed tags, as long as the parentheses are correctly
		// placed.
		phraseHTML = strings.ReplaceAll(phraseHTML, "(", "</strong>(")
		phraseHTML = strings.ReplaceAll(phraseHTML, ")", ")<strong>")

		// Remove superfluous tags that might have been created
		phraseHTML = strings.ReplaceAll(phraseHTML, "<strong> </strong>", " ")
		phraseHTML = strings.ReplaceAll(phraseHTML, "<strong></strong>", "")

		if isSinglePhrase {
			return phraseHTML
		}

		phraseList[i] = phraseHTML
	}

	return strings.Join(phraseList, separator+" ")
}

// renderConceptsByLetter renders a list of concepts as an HTML unordered list.
// Each concept is a link to its corresponding concept page. This is used on the letter pages.
func renderConceptsByLetter(concepts []string) string {
	var html strings.Builder
	html.WriteString(`<ul class="list-unstyled">`)
	for _, concept := range concepts {
		fmt.Fprintf(&html, `<li class="mb-3"><a class="concepte" href="/concepte/%s">%s</a></li>`,
			getConceptSlug(concept),
			getConceptTitleHTML(concept),
		)
	}
	html.WriteString(`</ul>`)
	return html.String()
}

// getAccepcio formats the "accepció" (meaning) text for display.
// If the text starts with a numbered item (e.g., "1."), it bolds the number.
// It also replaces any abbreviations with their full-text versions.
func getAccepcio(accepcioText string) string {
	formattedText := accepcioText

	spaceIndex := strings.Index(accepcioText, " ")
	if spaceIndex != -1 {
		firstWord := accepcioText[:spaceIndex]
		if isNumberedItem(firstWord) {
			remainingText := accepcioText[spaceIndex:]
			formattedText = fmt.Sprintf("<strong>%s</strong>%s", firstWord, remainingText)
		}
	}

	return fmt.Sprintf(`<div class="accepcio">%s</div>`, replaceAbbreviations(formattedText))
}

// isNumberedItem checks if a word is a numbered item, such as "1.".
// This is used to identify numbered meanings in the Entry.AccepcioConcepte field.
func isNumberedItem(word string) bool {
	if !strings.HasSuffix(word, ".") {
		return false
	}

	numberPart := strings.TrimSuffix(word, ".")
	_, err := strconv.Atoi(numberPart)
	return err == nil
}

// renderEntriesForConceptPage renders entries for a concept page, grouping them by "accepció".
func renderEntriesForConceptPage(entries []Entry) string {
	var htmlOutput strings.Builder
	var lastAccepcio string

	for _, entry := range entries {
		if entry.AccepcioConcepte != "" && entry.AccepcioConcepte != lastAccepcio {
			if lastAccepcio != "" {
				htmlOutput.WriteString(`<hr>`)
			}
			htmlOutput.WriteString(getAccepcio(entry.AccepcioConcepte))
			lastAccepcio = entry.AccepcioConcepte
		}
		htmlOutput.WriteString(`<article class="entry frase">`)
		htmlOutput.WriteString(renderSingleEntry(entry))
		htmlOutput.WriteString(`</article>`)
	}

	return htmlOutput.String()
}

// renderEntriesForSearch renders entries for a search results page, including the concept title for each.
func renderEntriesForSearch(entries []Entry) string {
	var htmlOutput strings.Builder

	for _, entry := range entries {
		htmlOutput.WriteString(`<article class="entry frase">`)
		fmt.Fprintf(&htmlOutput, `<h2 class="concepte"><a href="/concepte/%s">%s</a></h2>`,
			getConceptSlug(entry.Concepte),
			getConceptTitleHTML(entry.Concepte),
		)
		htmlOutput.WriteString(renderSingleEntry(entry))
		htmlOutput.WriteString(`</article>`)
	}

	return htmlOutput.String()
}

// renderSingleEntry renders the HTML for a single dictionary entry.
func renderSingleEntry(entry Entry) string {
	var htmlOutput strings.Builder

	if entry.AntonimConcepte {
		htmlOutput.WriteString(`<div><abbr title="valor antònim del concepte">ANT</abbr></div>`)
	}

	var phraseHTML string
	if entry.NovaIncorporacio {
		phraseHTML = getNewIncorporationPhrase(entry.Title)
	} else {
		phraseHTML = getPhrase(entry.Title)
	}

	fmt.Fprintf(&htmlOutput, `<p>%s %s, %s %s</p>`,
		phraseHTML,
		getCategory(entry.Categoria),
		entry.Definicio,
		getSources(entry.FontDefinicio),
	)

	if entry.Exemples != "" {
		fmt.Fprintf(&htmlOutput, "<p>%s %s</p>",
			replaceAbbreviationsParentheses(entry.Exemples),
			getSources(entry.FontExemples),
		)
	}
	if entry.Sinonims != "" {
		fmt.Fprintf(&htmlOutput, `<p><span class="simbol">→</span>%s</p>`,
			replaceAbbreviationsParentheses(renderBoldPhrases(entry.Sinonims, true)),
		)
	}
	if entry.AltresRelacions != "" {
		fmt.Fprintf(&htmlOutput, `<p><span class="simbol">▷</span>%s</p>`,
			replaceAbbreviationsParentheses(renderBoldPhrases(entry.AltresRelacions, true)),
		)
	}
	if entry.VariantsDialectals != "" {
		fmt.Fprintf(&htmlOutput, `<p><span class="simbol simbol-punt">•</span>%s</p>`,
			replaceAbbreviations(renderBoldPhrases(entry.VariantsDialectals, false)),
		)
	}
	if entry.MarcatgeDialectal != "" {
		fmt.Fprintf(&htmlOutput, `<p>[%s]</p>`, replaceSourceAbbreviationsParentheses(replaceAbbreviations(entry.MarcatgeDialectal)))
	}
	if entry.Observacions != "" {
		fmt.Fprintf(&htmlOutput, `<p>[%s]</p>`, replaceObservationsSourceAbbreviations(entry.Observacions))
	}

	return htmlOutput.String()
}

// getConceptTitleHTML formats a concept title for HTML display by converting numbers to superscripts.
// For example, "Concepte1" becomes "Concepte<sup>1</sup>".
func getConceptTitleHTML(concept string) string {
	return regexp.MustCompile(`(\d)`).ReplaceAllString(concept, "<sup>$1</sup>")
}

// getConceptTitle formats a concept title for display in page titles.
// It converts the title to lowercase and adds a space before any numbers.
func getConceptTitle(concept string) string {
	return strings.ToLower(regexp.MustCompile(`(\d)`).ReplaceAllString(concept, " $1"))
}

// getConceptSlug creates a URL-friendly slug from a concept title.
// It converts the title to lowercase and replaces spaces with underscores.
func getConceptSlug(concept string) string {
	slug := strings.ToLower(concept)
	slug = strings.Join(strings.Fields(slug), "_")
	return slug
}

// removeParenthesesContent removes content inside parentheses and brackets from a string.
// This is used to normalize phrases for searching and comparison.
func removeParenthesesContent(input string) string {
	content := input

	parenRegex := regexp.MustCompile(`\([^()]*\)`)
	for parenRegex.MatchString(content) {
		content = parenRegex.ReplaceAllString(content, "")
	}

	bracketRegex := regexp.MustCompile(`\[[^\[\]]*\]`)
	for bracketRegex.MatchString(content) {
		content = bracketRegex.ReplaceAllString(content, "")
	}

	content = strings.Join(strings.Fields(content), " ")
	content = strings.ReplaceAll(content, " , ", ", ")

	return strings.TrimSpace(content)
}

// toLowercaseNoAccents converts a string to lowercase and removes common Catalan accents.
// This is used for case-insensitive and accent-insensitive string comparisons.
func toLowercaseNoAccents(input string) string {
	removeAccentsReplacer := strings.NewReplacer(
		"à", "a", "è", "e", "é", "e", "í", "i", "ï", "i",
		"ò", "o", "ó", "o", "ú", "u", "ü", "u",
	)
	return removeAccentsReplacer.Replace(strings.ToLower(input))
}

// normalizeForSearch prepares a string for use as a search query.
// It removes parentheses, normalizes some characters (e.g., "’" to "'"),
// converts to lowercase, and removes accents.
func normalizeForSearch(input string) string {
	// TODO: ideally, we would also normalize Unicode here and in the database
	// export (NFC). But this has not been necessary so far.
	normalizeSearchReplacer := strings.NewReplacer(
		// Perform some UTF-8 normalizations
		"’", "'",
		"...", "…",
		// Remove some characters
		"(", "",
		")", "",
	)
	query := normalizeSearchReplacer.Replace(input)

	// Convert multiple spaces to single space
	query = strings.Join(strings.Fields(query), " ")

	// Trim, lowercase, and remove accents to match PHP export
	query = strings.Trim(query, "-, ")
	query = toLowercaseNoAccents(query)

	return query
}

// getEntries retrieves a paginated list of dictionary entries that match a search query.
// It supports different search modes (contains, starts with, ends with, exact match)
// and sorts the results alphabetically.
//
// Preconditions:
//   - normalizedQuery must be non-empty
//   - page must be >= 1
//   - pageSize must be >= 1
//
// Postconditions:
//   - Returns entries slice with length <= pageSize
//   - Returns total count of matching entries
//   - Results are sorted according to search mode and Catalan collation rules
//   - For default search mode, exact matches appear first
func getEntries(normalizedQuery, searchMode string, page, pageSize int) ([]Entry, int) {
	regex := regexp.MustCompile(fmt.Sprintf(`(^|[^\p{L}\p{M}])%s([^\p{L}\p{M}]|$)`, regexp.QuoteMeta(normalizedQuery)))

	var results []Entry
	for _, entry := range AllEntries {
		var match bool
		switch searchMode {
		// Search in normalized phrases (both without parentheses content and
		// without parentheses).
		case SearchModeComencaPer:
			match = strings.HasPrefix(entry.TitleNormalizedWpc, normalizedQuery) || strings.HasPrefix(entry.TitleNormalizedWp, normalizedQuery)
		case SearchModeAcabaEn:
			match = strings.HasSuffix(entry.TitleNormalizedWpc, normalizedQuery) || strings.HasSuffix(entry.TitleNormalizedWp, normalizedQuery)
		case SearchModeCoincident:
			match = entry.TitleNormalizedWpc == normalizedQuery || entry.TitleNormalizedWp == normalizedQuery
		default: // "Conté"
			match = regex.MatchString(entry.TitleNormalizedWpc) || (entry.TitleNormalizedWpc != entry.TitleNormalizedWp && regex.MatchString(entry.TitleNormalizedWp))
		}

		if match {
			results = append(results, entry)
		}
	}

	// Sort results by phrase
	collator := collate.New(language.Catalan)
	slices.SortFunc(results, func(a, b Entry) int {
		// For default search mode, show exact matches at the top
		if searchMode == "" || searchMode == SearchModeConte {
			// Check if either entry is an exact match
			aExact := a.TitleNormalizedWpc == normalizedQuery || a.TitleNormalizedWp == normalizedQuery
			bExact := b.TitleNormalizedWpc == normalizedQuery || b.TitleNormalizedWp == normalizedQuery

			// If one is exact and the other isn't, prioritize the exact match
			if aExact && !bExact {
				return -1
			}
			if !aExact && bExact {
				return 1
			}
		}

		// Sort alphabetically by normalized title.
		// If the normalized titles are the same without parentheses content,
		// consider the parentheses content.
		if a.TitleNormalizedWpc == b.TitleNormalizedWpc {
			return collator.CompareString(a.TitleNormalizedWp, b.TitleNormalizedWp)
		}

		// Sort alphabetically (without parentheses content)
		return collator.CompareString(a.TitleNormalizedWpc, b.TitleNormalizedWpc)
	})

	resultsCount := len(results)
	if resultsCount == 0 {
		return nil, resultsCount
	}

	// Slice for pagination
	start := (page - 1) * pageSize
	if start >= resultsCount {
		// Page is out of range
		return nil, resultsCount
	}

	end := min(start+pageSize, resultsCount)

	return results[start:end], resultsCount
}

// getEntriesByConceptSlug retrieves all dictionary entries for a given concept slug.
// The slug is converted back to the original concept format for matching.
//
// Postconditions:
//   - Returns all entries matching the concept (case-insensitive)
//   - Returns empty slice if no matches found
//   - Slug format: underscores converted to spaces for matching
func getEntriesByConceptSlug(conceptSlug string) []Entry {
	var records []Entry

	// Normalize the incoming slug back (space separated)
	conceptToMatch := strings.ReplaceAll(conceptSlug, "_", " ")

	for _, entry := range AllEntries {
		if strings.EqualFold(entry.Concepte, conceptToMatch) {
			records = append(records, entry)
		}
	}
	return records
}
