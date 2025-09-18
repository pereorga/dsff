/**
 * This toggles the visibility of the navigation menu on smaller screens.
 *
 * The code of this script is duplicated: it is inlined inside the main template in the Go implementation.
 * This file is now only used in the legacy PHP implementation.
 */

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
