document.addEventListener("DOMContentLoaded", () => {
  const notificationMenu = document.getElementById("notification-menu");
  const notificationList = document.getElementById("notification-items");
  const notificationPanel = document.getElementById("notification-list");
  const userId = window.userId || "UNKNOWN";

  console.log("notificationMenu:", notificationMenu);
  console.log("notificationList:", notificationList);
  console.log("notificationPanel:", notificationPanel);

  if (!notificationMenu || !notificationList || !notificationPanel) {
    console.error("Brak wymaganych elementów DOM dla powiadomień.");
    return;
  }

  // Funkcja do pobierania powiadomień
  const fetchNotifications = async () => {
    console.log("Rozpoczęto pobieranie powiadomień...");
    console.log("Aktualne userId:", userId);

    try {
      const response = await fetch(`/notifications?userId=${userId}`);
      console.log("Otrzymano odpowiedź z API:", response);

      if (!response.ok) {
        console.error(
          "Błąd w odpowiedzi API:",
          response.status,
          response.statusText
        );
        return;
      }

      const notifications = await response.json();
      console.log("Powiadomienia:", notifications);

      // Wyczyść listę przed dodaniem nowych elementów
      notificationList.innerHTML = "";

      if (notifications.length === 0) {
        console.log("Brak nowych powiadomień.");
        notificationList.innerHTML = ` 
          <li class="py-2 px-2 text-center text-gray-400">
            Brak nowych powiadomień.
          </li>`;
        return;
      }

      // Iteracja przez powiadomienia i dodanie ich do listy
      notifications.forEach((notification) => {
        console.log("Przetwarzanie powiadomienia:", notification);
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
    console.log("Kliknięto na Powiadomienia");

    if (notificationPanel.classList.contains("hidden")) {
      console.log("Otwieranie panelu powiadomień...");
    } else {
      console.log("Zamykanie panelu powiadomień...");
    }

    notificationPanel.classList.toggle("hidden");
    fetchNotifications(); // Pobierz powiadomienia przy każdym otwarciu
  });

  // Zamknięcie panelu, gdy klikniemy poza nim
  document.addEventListener("click", (event) => {
    if (
      !notificationPanel.contains(event.target) &&
      !notificationMenu.contains(event.target)
    ) {
      console.log("Kliknięto poza panelem powiadomień. Zamykanie...");
      notificationPanel.classList.add("hidden");
    }
  });
});