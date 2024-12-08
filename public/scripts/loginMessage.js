document.addEventListener("DOMContentLoaded", function () {
    const messageDiv = document.getElementById("login-message");
    const closeButton = document.getElementById("close-message");

    // Function to trigger the fade-out effect
    const fadeOut = () => {
        if (messageDiv) {
            messageDiv.classList.add("opacity-0");
            setTimeout(() => messageDiv.remove(), 500);
        }
    };

    // Auto-dismiss after 10 seconds
    if (messageDiv) {
        setTimeout(fadeOut, 10000);
    }

    // Manual dismiss
    if (closeButton) {
        closeButton.addEventListener("click", fadeOut);
    }
});