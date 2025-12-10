/**
 * Concept selector and minor UX search improvements.
 */

import TomSelect from "tom-select/base";
import conceptes from "./conceptes.json" with { type: "json" };

const removeCatalanAccents = function (str) {
  return str
    .replace(/à/g, "a")
    .replace(/[èé]/g, "e")
    .replace(/[íï]/g, "i")
    .replace(/[òó]/g, "o")
    .replace(/[úü]/g, "u");
};

const tomSelect = new TomSelect("#cerca-concepte", {
  options: conceptes,
  create: false,
  refreshThrottle: 0,
  openOnFocus: false,
  placeholder: "Introduïu un concepte",
  onChange(value) {
    if (value) {
      window.location.href = "/concepte/" + encodeURIComponent(value);
    }
  },
  render: {
    option(data) {
      return "<div>" + data.text + "</div>";
    },
    item(data) {
      return "<div>" + data.text + "</div>";
    },
    // eslint-disable-next-line camelcase
    no_results() {
      return "<div class='no-results'>No s'ha trobat cap resultat</div>";
    },
  },
  searchField: [],
  sortField: [],
  onType(query) {
    this.clearOptions();
    if (query.length) {
      const normalizedQuery = removeCatalanAccents(query.toLocaleLowerCase());

      // Separate options into two categories:
      // 1. Those that start with the normalized query.
      // 2. Those that contain the normalized query (but do not start with
      // it).
      const matchedOptions = [];
      const otherOptions = [];

      conceptes.forEach((option) => {
        const normalizedText = removeCatalanAccents(
          option.value.toLocaleLowerCase(),
        );
        if (normalizedText.startsWith(normalizedQuery)) {
          matchedOptions.push(option);
        } else if (normalizedText.includes(normalizedQuery)) {
          otherOptions.push(option);
        }
      });

      // Combine the two lists, prioritizing options that start with the
      // query.
      const sortedOptions = matchedOptions.concat(otherOptions);

      // Add sorted options back to the dropdown.
      sortedOptions.forEach((option) => {
        this.addOption(option);
      });

      this.refreshOptions();
    }
  },
});

// Ensure the following is executed with browser back/forward navigation.
window.addEventListener("pageshow", () => {
  const isMobile = /Android|iPad|iPhone/i.test(navigator.userAgent);

  // Ensure browser does not try to remember last form value, as it doesn't
  // help.
  const frase = document.getElementById("frase");
  frase.value = new URLSearchParams(location.search).get("frase") || "";

  // If the browser came from searching a concept, clear the concept selector
  // and focus it.
  if (document.getElementById("cerca-concepte").value) {
    tomSelect.clear(true);

    if (!isMobile) {
      // Focus the concept search input on desktop only. On mobile, this can
      // cause annoying issues with the keyboard (tested in Chrome on
      // Android).
      tomSelect.focus();
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
