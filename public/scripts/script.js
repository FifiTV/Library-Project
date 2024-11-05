document.addEventListener("DOMContentLoaded", function () {
    const contactLink = document.getElementById("contact-link");
    const contactPopup = document.getElementById("contact-popup");
    const closeBtn = document.querySelector(".close-btn");
  
    // Otwieranie pop-upu po kliknięciu linku
    contactLink.addEventListener("click", function (event) {
      event.preventDefault();
      contactPopup.style.display = "flex";
    });
  
    // Zamykanie pop-upu po kliknięciu "X"
    closeBtn.addEventListener("click", function () {
      contactPopup.style.display = "none";
    });
  
    // Zamykanie pop-upu po kliknięciu w tło
    window.addEventListener("click", function (event) {
      if (event.target === contactPopup) {
        contactPopup.style.display = "none";
      }
    });
  });
  