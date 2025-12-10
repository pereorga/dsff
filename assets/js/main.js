/**
 * Toggles the visibility of the navigation menu on smaller screens.
 *
 * Note: This script is inlined manually in go/templates/main.html.
 */

// Toggle menu on click.
const navbarToggler = document.querySelector(".navbar-toggler");
const navbarCollapse = document.querySelector("#navbar-collapse");
navbarToggler.addEventListener("click", () => {
  if (navbarCollapse.classList.contains("collapse")) {
    navbarCollapse.classList.remove("collapse");
    navbarCollapse.classList.add("show");
  } else {
    navbarCollapse.classList.add("collapse");
    navbarCollapse.classList.remove("show");
  }
});
