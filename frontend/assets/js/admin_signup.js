document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("adminSignupForm");
    const alertBox = document.getElementById("formAlert");

    form.addEventListener("submit", (e) => {
        e.preventDefault();
        alertBox.classList.add("hidden");
        alertBox.classList.remove("bg-red-100", "text-red-700", "bg-green-100", "text-green-700");

        const name = document.getElementById("name").value;
        const username = document.getElementById("username").value;
        const email = document.getElementById("email").value;
        const password = document.getElementById("password").value;

        fetch("/v1/admin-signup", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name, username, email, password })
        })
        .then(res => res.json())
        .then(data => {
            alertBox.classList.remove("hidden");
            if (data.success) {
                alertBox.classList.add("bg-green-100", "text-green-700");
                alertBox.textContent = data.message;
                setTimeout(() => {
                    window.location.href = "/login";
                }, 2000);
            } else {
                alertBox.classList.add("bg-red-100", "text-red-700");
                alertBox.textContent = data.message || "An error occurred.";
            }
        })
        .catch(err => {
            alertBox.classList.remove("hidden");
            alertBox.classList.add("bg-red-100", "text-red-700");
            alertBox.textContent = "Network error. Please try again.";
            console.error("Error signing up admin:", err);
        });
    });
});
