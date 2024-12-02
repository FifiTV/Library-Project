document.addEventListener("DOMContentLoaded", () => {
  const notificationMenu = document.getElementById("notification-menu");
  const notificationList = document.getElementById("notification-list");

  // Przełącz widoczność listy powiadomień po kliknięciu
  notificationMenu.addEventListener("click", (event) => {
    event.preventDefault();
    notificationList.classList.toggle("hidden");
  });

  // Zamknij listę, jeśli kliknięto poza nią
  document.addEventListener("click", (event) => {
    if (
      !notificationList.contains(event.target) &&
      !notificationMenu.contains(event.target)
    ) {
      notificationList.classList.add("hidden");
    }
  });
});
