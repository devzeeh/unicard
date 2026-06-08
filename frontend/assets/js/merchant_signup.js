document.addEventListener("DOMContentLoaded", () => {
    const signupForm = document.getElementById("merchantSignupForm");
    const formAlert = document.getElementById("formAlert");
    const submitBtn = document.getElementById("submitBtn");

    if (!signupForm) return;

    signupForm.addEventListener("submit", async (e) => {
        e.preventDefault();
        hideAlert();

        const businessName = document.getElementById("businessName").value.trim();
        const businessType = document.getElementById("businessType").value;
        const businessAddress = document.getElementById("businessAddress").value.trim();
        const ownerName = document.getElementById("ownerName").value.trim();
        const businessPhone = document.getElementById("businessPhone").value.trim();
        const businessEmail = document.getElementById("businessEmail").value.trim();
        const password = document.getElementById("password").value;
        const confirmPassword = document.getElementById("confirmPassword").value;

        // Basic validation
        if (password !== confirmPassword) {
            showAlert("Passwords do not match.", "error");
            return;
        }

        if (password.length < 6) {
            showAlert("Password must be at least 6 characters long.", "error");
            return;
        }

        // Prepare payload
        const payload = {
            businessName,
            businessType,
            businessAddress,
            ownerName,
            businessPhone,
            businessEmail,
            password
        };

        // Disable button to prevent double submission
        setLoading(true);

        try {
            const response = await fetch("/v1/merchant-signup", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(payload),
            });

            const data = await response.json();

            if (response.ok) {
                showAlert("Application submitted successfully! Please wait for admin approval before logging in.", "success");
                signupForm.reset();
                // Optionally redirect to login page after a delay
                setTimeout(() => {
                    window.location.href = "/login";
                }, 3000);
            } else {
                showAlert(data.message || "Failed to submit application. Please try again.", "error");
            }
        } catch (error) {
            console.error("Signup error:", error);
            showAlert("A network error occurred. Please try again later.", "error");
        } finally {
            setLoading(false);
        }
    });

    function showAlert(message, type) {
        formAlert.textContent = message;
        formAlert.classList.remove("hidden", "bg-red-100", "text-red-700", "border-red-400", "bg-green-100", "text-green-700", "border-green-400", "border");
        
        if (type === "error") {
            formAlert.classList.add("bg-red-100", "text-red-700", "border-red-400", "border");
        } else {
            formAlert.classList.add("bg-green-100", "text-green-700", "border-green-400", "border");
        }
    }

    function hideAlert() {
        formAlert.classList.add("hidden");
        formAlert.textContent = "";
    }

    function setLoading(isLoading) {
        if (isLoading) {
            submitBtn.disabled = true;
            submitBtn.innerHTML = `
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white inline-block" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Submitting...
            `;
            submitBtn.classList.add("opacity-75", "cursor-not-allowed");
        } else {
            submitBtn.disabled = false;
            submitBtn.innerHTML = "Submit Application";
            submitBtn.classList.remove("opacity-75", "cursor-not-allowed");
        }
    }
});
