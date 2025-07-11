/**
 * This toggles the visibility of the navigation menu on smaller screens.
 *
 * This script is inlined in the main HTML template to avoid an additional HTTP
 * request, but is bundled in search.min.js for the homepage and search pages,
 * as and additional JS file is needed there anyway.
 */

(() => {
  // Toggle menu on click.
  const navbarToggler = document.querySelector(".navbar-toggler");
  const navbarCollapse = document.querySelector("#navbar-collapse");
  navbarToggler.addEventListener("click", function () {
    if (navbarCollapse.classList.contains("collapse")) {
      navbarCollapse.classList.remove("collapse");
      navbarCollapse.classList.add("show");
    } else {
      navbarCollapse.classList.add("collapse");
      navbarCollapse.classList.remove("show");
    }
  });
})();
