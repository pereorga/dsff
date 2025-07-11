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

// Returns a map of all abbreviations and their full text.
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

// Returns a map of all sources (fonts) and their full text.
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

// Returns a map of all sources (fonts) and their full text used in Observacions field.
func getObservationSources() map[string]string {
	return map[string]string{
		"DIEC1": "Institut d'Estudis Catalans, Diccionari de la Llengua Catalana",
	}
}

// Returns the HTML for a given category key in Drupal.
func getCategory(categoryKey string) string {
	// Category mappings
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

// Loads and processes the data from the gzipped JSON file.
// This populates the global variables AllEntries, PhrasesMap, and
// ConceptsByFirstLetter.
func loadDataFromFile(filePath string) error {
	const letterCount = 23

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
	ConceptsByFirstLetter = make(map[string][]string, letterCount)

	// Populate data structures
	for _, entry := range AllEntries {
		PhrasesMap[removeParenthesesContent(entry.Title)] = true

		// Group by first letter
		firstRune := []rune(entry.Concepte)[0]
		key := strings.ToUpper(toLowercaseNoAccents(string(firstRune)))

		// Check if concept already exists
		if !slices.Contains(ConceptsByFirstLetter[key], entry.Concepte) {
			ConceptsByFirstLetter[key] = append(ConceptsByFirstLetter[key], entry.Concepte)
		}
	}

	// Sort each letter's concepts
	collator := collate.New(language.Catalan)
	for _, conceptList := range ConceptsByFirstLetter {
		slices.SortFunc(conceptList, collator.CompareString)
	}

	return nil
}

// Returns the canonical URL (pointing to production) for a given request.
// This is to avoid crawlers indexing the development/beta versions, and to
// try to make dsff.uab.es (which is also live!) disappear from Google search
// results.
func getCanonicalUrl(r *http.Request) string {
	return BaseCanonicalUrl + r.URL.RequestURI()
}

// Creates a strings.Replacer to replace abbreviations with abbr tags.
func createAbbrReplacer(abbrMap map[string]string, inParentheses bool) *strings.Replacer {
	var replacements []string
	for key, value := range abbrMap {
		pattern := key
		replacement := fmt.Sprintf("<abbr title=\"%s\">%s</abbr>", value, key)
		if inParentheses {
			pattern = "(" + key + ")"
			replacement = "(" + replacement + ")"
		}
		replacements = append(replacements, pattern, replacement)
	}
	return strings.NewReplacer(replacements...)
}

// Replaces abbreviations that are inside parentheses.
func replaceAbbreviationsParentheses(text string) string {
	// (v.f.) -> (<abbr title="...">v.f.</abbr>)
	return createAbbrReplacer(getAllAbbreviations(), true).Replace(text)
}

// Replaces abbreviations.
// To be used only when the more selective replaceAbbreviationsParentheses()
// is not suitable, as there is more risk for unwanted replacements.
func replaceAbbreviations(text string) string {
	// v.f. -> <abbr title="...">v.f.</abbr>
	return createAbbrReplacer(getAllAbbreviations(), false).Replace(text)
}

// Replaces source abbreviations, in parentheses.
func replaceSourceAbbreviationsParentheses(text string) string {
	// (DIEC1) -> (<abbr title="...">DIEC1</abbr>)
	return createAbbrReplacer(getAllSources(), true).Replace(text)
}

// Replaces source abbreviations for Observacions field.
func replaceObservationsSourceAbbreviations(text string) string {
	// DIEC1 -> <abbr title="...">DIEC1</abbr>
	return createAbbrReplacer(getObservationSources(), false).Replace(text)
}

// Formats the sources (fonts) string with abbr tags.
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

// Formats a single phrase, adding a marker for new incorporations.
func getPhrase(phrase string, isNewIncorporation bool) string {
	if isNewIncorporation {
		return "■ " + renderBoldPhrases(phrase, true)
	}
	return renderBoldPhrases(phrase, true)
}

// Checks if a phrase exists.
func phraseExists(phrase string) bool {
	return PhrasesMap[removeParenthesesContent(phrase)]
}

// Splits a string by a separator, but ignoring separators inside parentheses.
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

// Renders phrases in bold, optionally creating links for them.
func renderBoldPhrases(input string, createLink bool) string {
	const placeholderUnusedChar = "|"

	if input == "" {
		return ""
	}

	// By default, assume input can be multiple phrases separated by a comma
	separator := ","
	var isSinglePhrase bool

	// Phrases that should not be split or linked
	RenderBoldPhrasesWhitelist := []string{
		"Jesús, Maria i Josep (v.f.)",
		"en Pere, en Pau i en Berenguera (v.f.)",
		"córrer la Seca, la Meca i la vall d'Andorra (v.f.)",
	}

	if phraseExists(input) || slices.Contains(RenderBoldPhrasesWhitelist, input) {
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
		shouldCreateLink := createLink && !isFormalVariant

		var html strings.Builder
		if shouldCreateLink {
			if phraseExists(phrase) {
				searchPath := "/?mode=Conté&frase=" + url.QueryEscape(removeParenthesesContent(phrase))
				fmt.Fprintf(&html, "<a href=\"%s\" rel=\"nofollow\"><strong>%s</strong></a>", searchPath, phrase)
			} else {
				// No reference found (this should not happen)
				fmt.Fprintf(&html, "<span class=\"bg-danger text-white py-1\"><strong>%s</strong></span>", phrase)
			}
		} else {
			fmt.Fprintf(&html, "<strong>%s</strong>", phrase)
		}
		htmlString := html.String()

		// Make parentheses non-bold. This should not leave
		// inconsistent/unclosed tags, as long as the parentheses are correctly
		// placed.
		htmlString = strings.ReplaceAll(htmlString, "(", "</strong>(")
		htmlString = strings.ReplaceAll(htmlString, ")", ")<strong>")

		// Remove superfluous tags that might have been created
		htmlString = strings.ReplaceAll(htmlString, "<strong> </strong>", " ")
		htmlString = strings.ReplaceAll(htmlString, "<strong></strong>", "")

		if isSinglePhrase {
			return htmlString
		}

		phraseList[i] = htmlString
	}

	return strings.Join(phraseList, separator+" ")
}

// Renders the provided list of concepts.
// Used in letter pages.
func renderConceptsByLetter(concepts []string) string {
	var html strings.Builder
	html.WriteString(`<ul class="list-unstyled">`)
	for _, concept := range concepts {
		fmt.Fprintf(&html, `<li class="mb-3"><a class="concepte" href="/concepte/%s">%s</a></li>`,
			getConceptSlug(concept),
			getConceptTitleHtml(concept),
		)
	}
	html.WriteString(`</ul>`)
	return html.String()
}

// Formats the "accepció" text.
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

// Checks if a word is a numbered item (e.g., "1.").
// This is used to identify numbered acceptions in Entry.AccepcioConcepte.
func isNumberedItem(word string) bool {
	if !strings.HasSuffix(word, ".") {
		return false
	}

	numberPart := strings.TrimSuffix(word, ".")
	_, err := strconv.Atoi(numberPart)
	return err == nil
}

// Renders a list of entries into HTML.
func renderPhrases(entries []Entry, isConceptPage bool) string {
	var htmlOutput strings.Builder
	var lastAccepcio string

	for _, entry := range entries {
		if isConceptPage {
			if entry.AccepcioConcepte != "" && entry.AccepcioConcepte != lastAccepcio {
				if lastAccepcio != "" {
					htmlOutput.WriteString(`<hr>`)
				}
				htmlOutput.WriteString(getAccepcio(entry.AccepcioConcepte))
				lastAccepcio = entry.AccepcioConcepte
			}
			htmlOutput.WriteString(`<article class="entry frase">`)
		} else {
			htmlOutput.WriteString(`<article class="entry frase">`)
			fmt.Fprintf(&htmlOutput, `<h2 class="concepte"><a href="/concepte/%s">%s</a></h2>`,
				getConceptSlug(entry.Concepte),
				getConceptTitleHtml(entry.Concepte),
			)
		}

		if entry.AntonimConcepte {
			htmlOutput.WriteString(`<div><abbr title="valor antònim del concepte">ANT</abbr></div>`)
		}

		fmt.Fprintf(&htmlOutput, `<p>%s %s, %s %s</p>`,
			getPhrase(entry.Title, entry.NovaIncorporacio),
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

		htmlOutput.WriteString(`</article>`)
	}

	return htmlOutput.String()
}

// Formats a concept title with HTML for superscripts.
func getConceptTitleHtml(concept string) string {
	return regexp.MustCompile(`(\d)`).ReplaceAllString(concept, "<sup>$1</sup>")
}

// Formats a concept title for display, in lowercase, with a space before
// numbers.
func getConceptTitle(concept string) string {
	return strings.ToLower(regexp.MustCompile(`(\d)`).ReplaceAllString(concept, " $1"))
}

// Creates a URL-friendly slug from a concept title, replacing spaces with
// underscores, and normalizing to lowercase.
func getConceptSlug(concept string) string {
	// Normalize to lowercase and replace spaces with underscores for URL
	slug := strings.ToLower(concept)
	slug = strings.Join(strings.Fields(slug), "_")
	return slug
}

// Removes content inside parentheses and brackets from a string.
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

// Converts a string to lowercase and removes Catalan accents.
func toLowercaseNoAccents(input string) string {
	removeAccentsReplacer := strings.NewReplacer(
		"à", "a", "è", "e", "é", "e", "í", "i", "ï", "i",
		"ò", "o", "ó", "o", "ú", "u", "ü", "u",
	)
	return removeAccentsReplacer.Replace(strings.ToLower(input))
}

// Normalizes a string for searching.
// It removes parentheses characters, converts to lowercase, does some minor
// character replacements, and removes accents.
func normalizeForSearch(input string) string {
	// Apply some normalizations to match current export, removing parentheses.
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

// Searches for entries based on a query and search mode.
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
		// For default search mode, prioritize exact matches first
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

// Retrieves all dictionary entries for a given concept slug.
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
