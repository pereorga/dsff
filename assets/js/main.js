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
