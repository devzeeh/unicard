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

    const uidInput = document.getElementById("cardUID");
    const amountInput = document.getElementById("initialAmount");
    const previewUid = document.getElementById("preview-uid");
    const previewBalance = document.getElementById("preview-balance");

    // ==========================================
    // 1. MQTT WEBSOCKET INTEGRATION (Hardware Auto-fill)
    // ==========================================
    const brokerUrl = 'ws://localhost:9001'; // 192.168.254.104 
    const mqttClient = typeof mqtt !== 'undefined' ? mqtt.connect(brokerUrl) : null;

    if (mqttClient) {
        mqttClient.on('connect', () => {
            console.log('✅ Connected to Mosquitto Broker via WebSockets.');
            mqttClient.subscribe('unicard/admin/register');
        });

        mqttClient.on('message', (topic, message) => {
            if (topic === 'unicard/admin/register') {
                const scannedUID = message.toString().trim().toUpperCase();
                console.log('🔥 Hardware Scan Detected! UID:', scannedUID);

                if (uidInput && previewUid) {
                    uidInput.value = scannedUID;
                    previewUid.textContent = scannedUID;

                    // Visual feedback: Flash the input green
                    uidInput.classList.add('ring-2', 'ring-green-500');
                    setTimeout(() => uidInput.classList.remove('ring-2', 'ring-green-500'), 1500);
                }
            }
        });
    } else {
        console.error("MQTT library not found. Add <script src='https://unpkg.com/mqtt/dist/mqtt.min.js'></script> to your HTML.");
    }

    // ==========================================
    // 2. UI DYNAMICS (Manual Typing)
    // ==========================================
    if (form) {
        if(uidInput && previewUid) {
            uidInput.addEventListener('input', (e) => {
                const val = e.target.value.trim().toUpperCase();
                previewUid.textContent = val || 'A346F101';
            });
        }

        if(amountInput && previewBalance) {
            amountInput.addEventListener('input', (e) => {
                const val = parseFloat(e.target.value);
                if (!isNaN(val)) {
                    previewBalance.textContent = val.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
                } else {
                    previewBalance.textContent = '0.00';
                }
            });
        }

        // ==========================================
        // 3. REAL BACKEND SUBMISSION (Go Database)
        // ==========================================
        form.addEventListener("submit", function (e) {
            e.preventDefault();
            if (errorAlert) errorAlert.classList.add("hidden");
            if (successAlert) successAlert.classList.add("hidden");

            const cardUID = uidInput.value.trim().toUpperCase();
            const initialAmount = amountInput.value;

            // Grab the admin username from the URL path
            const adminUsername = window.location.pathname.split('/')[2];
            
            fetch(`/v1/admin/${adminUsername}/addcardauth`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                // Make sure these keys match exactly what your Go struct expects!
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
                    // Reset the form and the dynamic card preview
                    form.reset();
                    if(previewUid) previewUid.textContent = 'A346F101';
                    if(previewBalance) previewBalance.textContent = '0.00';
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