function openLogoutModal() {
    document.getElementById('logout-modal').classList.remove('hidden');
}

function closeLogoutModal() {
    document.getElementById('logout-modal').classList.add('hidden');
}

document.addEventListener("DOMContentLoaded", function () {
    const cardNumInput = document.getElementById("cardNumber");
    const nameInput = document.getElementById("name");
    const cardTypeSelect = document.getElementById("cardType");

    // Submit form via fetch
    const form = document.getElementById("deactivate-card-form");
    const errorAlert = document.getElementById("error-alert");
    const successAlert = document.getElementById("success-alert");
    const errorText = document.getElementById("error-text");
    const successText = document.getElementById("success-text");

    if (form) {
        form.addEventListener("submit", function (e) {
            e.preventDefault();
            if (errorAlert) errorAlert.classList.add("hidden");
            if (successAlert) successAlert.classList.add("hidden");

            const bodyData = {
                cardNumber: cardNumInput.value,
                cardHolder: nameInput.value,
                cardType: cardTypeSelect.value
            };

            fetch("/v1/admin/deactivatecardauth", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(bodyData)
            })
                .then(res => res.json())
                .then(data => {
                    if (data.success) {
                        if (successAlert && successText) {
                            successText.innerText = data.message;
                            successAlert.classList.remove("hidden");
                        }
                        form.reset();
                    } else {
                        if (errorAlert && errorText) {
                            errorText.innerText = data.message;
                            errorAlert.classList.remove("hidden");
                        }
                    }
                })
                .catch(err => {
                    console.error(err);
                    if (errorAlert && errorText) {
                        errorText.innerText = "Network error. Please try again.";
                        errorAlert.classList.remove("hidden");
                    }
                });
        });
    }
});