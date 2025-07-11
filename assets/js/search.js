/**
 * JavaScript code for the concept selector on the search page.
 *
 * It also handles minor UX search improvements, such as selecting the search
 * input.
 */

(() => {
  const removeCatalanAccents = function (str) {
    return str
      .replace(/à/g, "a")
      .replace(/[èé]/g, "e")
      .replace(/[íï]/g, "i")
      .replace(/[òó]/g, "o")
      .replace(/[úü]/g, "u");
  };

  const tom_select = new TomSelect("#cerca-concepte", {
    options: conceptes,
    create: false,
    refreshThrottle: 0,
    openOnFocus: false,
    placeholder: "Introduïu un concepte",
    onChange: function (value) {
      if (value) {
        window.location.href = "/concepte/" + encodeURIComponent(value);
      }
    },
    render: {
      option: function (data, escape) {
        return "<div>" + data.text + "</div>";
      },
      item: function (data, escape) {
        return "<div>" + data.text + "</div>";
      },
      no_results: function (data, escape) {
        return "<div class='no-results'>No s'ha trobat cap resultat</div>";
      },
    },
    searchField: [],
    sortField: [],
    onType: function (query) {
      const self = this;
      self.clearOptions();
      if (query.length) {
        const normalizedQuery = removeCatalanAccents(query.toLocaleLowerCase());

        // Separate options into two categories:
        // 1. Those that start with the normalized query.
        // 2. Those that contain the normalized query (but do not start with
        // it).
        const matchedOptions = [];
        const otherOptions = [];

        conceptes.forEach(function (option) {
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
        sortedOptions.forEach(function (option) {
          self.addOption(option);
        });

        self.refreshOptions();
      }
    },
  });

  // Ensure the following is executed with browser back/forward navigation.
  window.addEventListener("pageshow", function () {
    const isMobile = /Android|iPad|iPhone/i.test(navigator.userAgent);

    // Ensure browser does not try to remember last form value, as it doesn't
    // help.
    const frase = document.getElementById("frase");
    frase.value = new URLSearchParams(location.search).get("frase") || "";

    // If the browser came from searching a concept, clear the concept selector
    // and focus it.
    if (document.getElementById("cerca-concepte").value) {
      tom_select.clear(true);

      if (!isMobile) {
        // Focus the concept search input on desktop only. On mobile, this can
        // cause annoying issues with the keyboard (tested in Chrome on
        // Android).
        tom_select.focus();
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
})();
