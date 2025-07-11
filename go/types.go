package main

import "html/template"

// Represents a dictionary entry.
// See Drupal export at preprocessNodeJson() in
// web/modules/custom/dsff_custom/src/Commands/DsffCustomDrushCommands.php.
type Entry struct {
	Title              string `json:"title"`                // The phrase used for rendering.
	TitleNormalizedWp  string `json:"title_normalized_wp"`  // The phrase in lowercase, without accents, without parentheses. Used for searching.
	TitleNormalizedWpc string `json:"title_normalized_wpc"` // The phrase in lowercase, without accents, without parentheses and their contents. Used for searching and sorting.
	Concepte           string `json:"concepte"`             // The concept related to the phrase.
	AntonimConcepte    bool   `json:"antonim_concepte"`     // True if the phrase is related to the antonym of the concept instead (usually false).
	AccepcioConcepte   string `json:"accepcio_concepte"`    // Optional: only used for concepts that have multiple meanings, or when the meaning is figurative.
	NovaIncorporacio   bool   `json:"nova_incorporacio"`    // True if the phrase does not exist on any other source (usually false).
	Categoria          string `json:"categoria"`            // The category of the phrase, e.g. "sv" for "Sintagma Verbal".
	Definicio          string `json:"definicio"`            // The definition.
	FontDefinicio      string `json:"font_definicio"`       // Optional: list of sources of the definitions.
	Exemples           string `json:"exemples"`             // Examples of the phrase.
	FontExemples       string `json:"font_exemples"`        // Optional: list of sources of the examples.
	Sinonims           string `json:"sinonims"`             // Optional: list of synonyms.
	AltresRelacions    string `json:"altres_relacions"`     // Optional: list of related phrases.
	VariantsDialectals string `json:"variants_dialectals"`  // Optional: list of dialectal variants.
	MarcatgeDialectal  string `json:"marcatge_dialectal"`   // Optional: dialectal information of the phrase.
	Observacions       string `json:"observacions"`         // Optional: miscellaneous observations.
}

// Represents the data for rendering a page.
// Used in the main template.
type PageData struct {
	Title        string
	CanonicalUrl string

	// Flags to indicate the page being rendered
	IsHomepage         bool
	IsAbreviaturesPage bool
	IsConceptPage      bool
	IsConeixPage       bool
	IsCreditsPage      bool
	IsLetterPage       bool
	IsPresentacioPage  bool

	// Search functionality
	SearchQuery  string
	SearchMode   string
	SearchModes  []string
	CurrentPage  int
	TotalPages   int
	PreviousPage int
	NextPage     int

	// Used in concept pages
	Concept template.HTML // The concept title. May contain HTML, e.g. <sup>1</sup>.

	// Used in letter pages
	Letter     string        // The letter ({A-Z}).
	LetterHtml template.HTML // Body of the letter page.

	// Used in search and concept pages
	PhrasesHtml template.HTML // List of rendered, clickable phrases.
}
