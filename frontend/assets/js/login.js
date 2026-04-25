document.addEventListener("DOMContentLoaded", function () {
    const loginForm = document.getElementById("login-form");
    const errorMessage = document.getElementById("errorMessage");
    
    if (loginForm) {
        loginForm.addEventListener("submit", function (e) {
            e.preventDefault();
            
            // Hide error message initially
            if (errorMessage) errorMessage.classList.add("hidden");
            
            const idValue = document.getElementById("identifier").value;
            const pass = document.getElementById("password").value;

            fetch("v1/loginauth", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ identifier: idValue, password: pass })
            })
                .then(r => r.json())
                .then(data => {
                    if (data.success) {
                        window.location.href = "/dashboard";
                    } else {
                        if (errorMessage) {
                            errorMessage.classList.remove("hidden");
                            errorMessage.innerText = data.message || "An error occurred";
                        }
                    }
                })
                .catch(err => {
                    console.error(err);
                    if (errorMessage) {
                        errorMessage.classList.remove("hidden");
                        errorMessage.innerText = "Network error. Please try again.";
                    }
                });
        });
    }
});