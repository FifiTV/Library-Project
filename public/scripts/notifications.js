document.addEventListener("DOMContentLoaded", () => {
  const bellButton = document.getElementById("notification-bell");
  const notificationList = document.getElementById("notification-list");

  // Ukryj listę powiadomień na starcie
  notificationList.classList.add("hidden");

  // Przełącz widoczność listy po kliknięciu
  bellButton.addEventListener("click", () => {
    notificationList.classList.toggle("hidden");
  });

  // Zamknij listę, jeśli użytkownik kliknie poza nią
  document.addEventListener("click", (event) => {
    if (
      !notificationList.contains(event.target) &&
      !bellButton.contains(event.target)
    ) {
      notificationList.classList.add("hidden");
    }
  });
});
