function openLogoutModal() {
    document.getElementById('logout-modal').classList.remove('hidden');
}

function closeLogoutModal() {
    document.getElementById('logout-modal').classList.add('hidden');
}

document.addEventListener("DOMContentLoaded", function () {
    const form = document.getElementById("add-card-form");
    const errorAlert = document.getElementById("error-alert");
    const successAlert = document.getElementById("success-alert");
    const errorText = document.getElementById("error-text");
    const successText = document.getElementById("success-text");

    if (form) {
        form.addEventListener("submit", function (e) {
            e.preventDefault();
            if (errorAlert) errorAlert.classList.add("hidden");
            if (successAlert) successAlert.classList.add("hidden");

            const cardUID = document.getElementById("cardUID").value;
            const initialAmount = document.getElementById("initialAmount").value;

            const adminUsername = window.location.pathname.split('/')[2];
            fetch(`/v1/admin/${adminUsername}/addcardauth`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    card_uid: cardUID,
                    initial_amount: parseFloat(initialAmount)
                })
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