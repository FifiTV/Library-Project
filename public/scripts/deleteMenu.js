document.addEventListener("DOMContentLoaded", () => {
  const accountMenu = document.querySelector(".account-menu > a");
  const dropdown = document.querySelector(".account-dropdown");

  // Obsługa kliknięcia na "Konto"
  accountMenu.addEventListener("click", (event) => {
    event.preventDefault(); // Zapobiega domyślnemu przeładowaniu strony
    dropdown.classList.toggle("hidden");
  });

  // Zamknięcie menu po kliknięciu poza nim
  document.addEventListener("click", (event) => {
    if (!accountMenu.parentElement.contains(event.target)) {
      dropdown.classList.add("hidden");
    }
  });
});
