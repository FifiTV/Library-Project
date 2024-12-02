document.addEventListener("DOMContentLoaded", () => {
  const notificationMenu = document.getElementById("notification-menu");
  const notificationList = document.getElementById("notification-items");
  const notificationPanel = document.getElementById("notification-list");

  // Funkcja do pobierania powiadomień
  const fetchNotifications = async () => {
    const userId = "CURRENT_USER_ID"; // Podmień na aktualne ID użytkownika (np. z sesji)
    try {
      const response = await fetch(`/notifications?userId=${userId}`);
      const notifications = await response.json();

      // Wyczyść listę przed dodaniem nowych elementów
      notificationList.innerHTML = "";

      if (notifications.length === 0) {
        notificationList.innerHTML = `
          <li class="py-2 px-2 text-center text-gray-400" style="background-color: #333;">
            Brak nowych powiadomień.
          </li>`;
        return;
      }

      // Iteracja przez powiadomienia i dodanie ich do listy
      notifications.forEach((notification) => {
        const li = document.createElement("li");
        li.className =
          "py-2 px-2 border-b border-gray-600 hover:bg-gray-600 rounded-lg transition";
        li.innerHTML = `
          <span class="font-semibold">${notification.message}</span>
          <br />
          <span class="text-sm text-gray-400">Książka: ${notification.bookTitle}</span>
        `;
        notificationList.appendChild(li);
      });
    } catch (error) {
      console.error("Błąd podczas pobierania powiadomień:", error);
    }
  };

  // Obsługa kliknięcia na menu powiadomień
  notificationMenu.addEventListener("click", (event) => {
    event.preventDefault();
    notificationPanel.classList.toggle("hidden");
    fetchNotifications(); // Pobierz powiadomienia przy każdym otwarciu
  });

  // Zamknięcie panelu, gdy klikniemy poza nim
  document.addEventListener("click", (event) => {
    if (
      !notificationPanel.contains(event.target) &&
      !notificationMenu.contains(event.target)
    ) {
      notificationPanel.classList.add("hidden");
    }
  });
});
