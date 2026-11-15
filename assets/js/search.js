/**
 * Concept selector and minor UX search improvements.
 */

import { createAutocomplete } from "./autocomplete.js";
import rawTitles from "./conceptes.json" with { type: "json" };

const collator = new Intl.Collator("ca-ES");

function normalizeQuery(str) {
  return str
    .toLocaleLowerCase("ca-ES")
    .replace(/à/g, "a")
    .replace(/[èé]/g, "e")
    .replace(/[íï]/g, "i")
    .replace(/[òó]/g, "o")
    .replace(/[úü]/g, "u");
}

function escapeHtml(str) {
  return str
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

/**
 * Builds the suggestion HTML for a raw title, wrapping every substring that
 * matches the normalized query in a highlight span.
 *
 * Matching uses the normalized title, so a query without accents (e.g.
 * "absencia") still highlights the accented title ("absència"). Each
 * normalized character keeps a reference to its raw-title position to map the
 * match back onto the displayed text.
 */
function highlightMatches(raw, normalizedQuery) {
  const chars = Array.from(raw);
  const normalizedChars = [];
  const rawIndexByNormalizedIndex = [];

  for (const [rawIndex, char] of chars.entries()) {
    for (const normalizedChar of normalizeQuery(char)) {
      normalizedChars.push(normalizedChar);
      rawIndexByNormalizedIndex.push(rawIndex);
    }
  }

  const normalizedTitle = normalizedChars.join("");
  const highlighted = new Array(chars.length).fill(false);

  if (normalizedQuery) {
    let searchFrom = 0;
    let matchIndex = normalizedTitle.indexOf(normalizedQuery, searchFrom);

    while (matchIndex !== -1) {
      const start = rawIndexByNormalizedIndex[matchIndex];
      const end =
        rawIndexByNormalizedIndex[matchIndex + normalizedQuery.length - 1];

      for (let i = start; i <= end; i += 1) {
        highlighted[i] = true;
      }

      searchFrom = matchIndex + normalizedQuery.length;
      matchIndex = normalizedTitle.indexOf(normalizedQuery, searchFrom);
    }
  }

  let html = "";
  let run = "";
  let runHighlighted = false;

  for (const [index, char] of chars.entries()) {
    const charHtml = /\d/.test(char)
      ? `<sup>${escapeHtml(char)}</sup>`
      : escapeHtml(char);

    if (index > 0 && highlighted[index] !== runHighlighted) {
      html += runHighlighted ? `<span class="highlight">${run}</span>` : run;
      run = "";
    }

    run += charHtml;
    runHighlighted = highlighted[index];
  }

  html += runHighlighted ? `<span class="highlight">${run}</span>` : run;

  return html;
}

const conceptes = rawTitles
  .map((title) => ({
    value: title.toLocaleLowerCase("ca-ES").replaceAll(" ", "_"),
    text: title.replace(/(\d)/g, "<sup>$1</sup>"),
    raw: title,
    title: title.toLocaleLowerCase("ca-ES"),
    search: normalizeQuery(title),
  }))
  .sort((a, b) => collator.compare(a.value, b.value));

const conceptInput = document.getElementById("cerca-concepte");
const isMobile = /Android|iPad|iPhone/i.test(navigator.userAgent);

const autocompleteInstance = createAutocomplete({
  inputElement: conceptInput,
  containerElement: conceptInput.closest(".concept-search"),
  data: conceptes,
  displayKey: "text",
  searchKey: "search",
  titleKey: "title",
  renderItem: (entry, searchTerm, searchTermNormalized) =>
    highlightMatches(entry.raw, searchTermNormalized),
  onSelect: (entry) => {
    window.location.href = "/concepte/" + encodeURIComponent(entry.value);
  },
  enableTextSelect: isMobile,
});

// Ensure the following is executed with browser back/forward navigation.
window.addEventListener("pageshow", () => {
  // Ensure browser does not try to remember last form value, as it doesn't
  // help.
  const frase = document.getElementById("frase");
  frase.value = new URLSearchParams(location.search).get("frase") || "";

  const hadConcept = conceptInput.value !== "";
  conceptInput.value = "";
  autocompleteInstance.close();

  // If the browser came from searching a concept, focus the concept selector.
  if (hadConcept) {
    if (!isMobile) {
      // Focus the concept search input on desktop only. On mobile, this can
      // cause annoying issues with the keyboard (tested in Chrome on
      // Android).
      conceptInput.focus();
    }

    return;
  }

  // By default, default to focusing the phrase search input.
  if (!isMobile) {
    // On desktop, select the searched value, so it can be replaced by simply
    // typing. This also focuses the input when the field is empty.
    frase.select();
  }
});
