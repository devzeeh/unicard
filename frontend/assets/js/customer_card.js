document.addEventListener("DOMContentLoaded", function () {
    
    // --- Report Card Modal Elements ---
    const reportButton = document.getElementById('report-card-button');
    const reportModal = document.getElementById('report-card-modal');
    const reportModalContent = document.getElementById('report-card-modal-content');
    const closeReportModalButton = document.getElementById('report-modal-close-button');
    const cancelReportModalButton = document.getElementById('report-modal-cancel-button');
    const confirmReportButton = document.getElementById('report-modal-confirm-button');

    // --- Replacement Modal Elements ---
    const replacementButton = document.getElementById('request-replacement-button');
    const replacementModal = document.getElementById('replacement-modal');
    const replacementModalContent = document.getElementById('replacement-modal-content');
    const closeReplacementModalButton = document.getElementById('replacement-modal-close-button');
    const cancelReplacementModalButton = document.getElementById('replacement-modal-cancel-button');
    const confirmReplacementButton = document.getElementById('replacement-modal-confirm-button');

    // --- Status Badge Element ---
    const cardStatusBadge = document.getElementById('card-status-badge');


    // --- Robustness Check ---
    const reportElementsExist = reportButton && reportModal && reportModalContent && 
                                closeReportModalButton && cancelReportModalButton && confirmReportButton;
    
    const replacementElementsExist = replacementButton && replacementModal && replacementModalContent &&
                                     closeReplacementModalButton && cancelReplacementModalButton && confirmReplacementButton;
    
    if (!reportElementsExist) {
        console.error("My Card script error: Report Card modal elements were not found. Check your IDs.");
    }
    
    if (!replacementElementsExist) {
        console.error("My Card script error: Replacement Modal elements were not found. Check your IDs.");
    }
     
    if (!cardStatusBadge) {
         console.error("My Card script error: Card Status Badge element not found. Check your ID.");
    }
    
    // Stop the script if critical elements are missing
    if (!reportElementsExist || !replacementElementsExist || !cardStatusBadge) {
        // We still want to run the toggle lock logic, so don't return here entirely if we're just ignoring old missing elements
        console.warn("Some card management elements are missing, but we'll continue for lock toggle.");
    }

    // --- Helper to get User ID ---
    const pathSegments = window.location.pathname.split('/');
    let userId = null;
    if (pathSegments.length >= 3 && pathSegments[1] === 'u') {
        userId = pathSegments[2];
    } else {
        const params = new URLSearchParams(window.location.search);
        userId = params.get('username');
    }

    function updateCardStatusAPI(newStatus, onSuccess) {
        if (!userId) return;
        fetch(`/v1/user/${encodeURIComponent(userId)}/card/status`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status: newStatus })
        })
        .then(res => {
            if (res.ok) onSuccess();
            else alert("Failed to update card status. Please try again.");
        })
        .catch(err => {
            console.error(err);
            alert("Error updating card status.");
        });
    }

    function requestReplacementAPI(onSuccess) {
        if (!userId) return;
        fetch(`/v1/user/${encodeURIComponent(userId)}/card/replace`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' }
        })
        .then(res => res.json().then(data => ({ status: res.status, ok: res.ok, data })))
        .then(res => {
            if (res.ok && res.data.success) {
                onSuccess();
            } else {
                alert(res.data.message || "Failed to request replacement. Please try again.");
            }
        })
        .catch(err => {
            console.error(err);
            alert("Error requesting replacement.");
        });
    }

    // --- Card Toggle & Copy Logic ---
    const toggleCardBtn = document.getElementById("toggle-card-btn");
    const copyCardBtn = document.getElementById("copy-card-btn");
    const userCardSpan = document.getElementById("user-card-number");
    const eyeOpen = document.getElementById("eye-icon-open");
    const eyeClosed = document.getElementById("eye-icon-closed");
    
    if (toggleCardBtn && userCardSpan) {
        toggleCardBtn.addEventListener('click', () => {
            const isHidden = userCardSpan.getAttribute("data-hidden") === "true";
            const fullNum = userCardSpan.getAttribute("data-full");
            const maskedNum = userCardSpan.getAttribute("data-masked");

            if (isHidden) {
                // Show full card number
                userCardSpan.innerText = fullNum;
                userCardSpan.setAttribute("data-hidden", "false");
                eyeOpen.classList.remove("hidden");
                eyeClosed.classList.add("hidden");
            } else {
                // Hide card number
                userCardSpan.innerText = maskedNum;
                userCardSpan.setAttribute("data-hidden", "true");
                eyeOpen.classList.add("hidden");
                eyeClosed.classList.remove("hidden");
            }
        });
    }

    if (copyCardBtn && userCardSpan) {
        copyCardBtn.addEventListener('click', () => {
            const fullNum = userCardSpan.getAttribute("data-full");
            if (fullNum && fullNum !== "•••• •••• •••• ••••") {
                // Copy strictly numbers, without spaces if preferred, or copy formatted.
                // Let's copy formatted as usually expected by users for cards.
                navigator.clipboard.writeText(fullNum.replace(/\s+/g, '')).then(() => {
                    const originalHTML = copyCardBtn.innerHTML;
                    // Show a checkmark temporarily
                    copyCardBtn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 text-green-400"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>`;
                    setTimeout(() => {
                        copyCardBtn.innerHTML = originalHTML;
                    }, 2000);
                }).catch(err => {
                    console.error("Failed to copy: ", err);
                    alert("Failed to copy card number");
                });
            }
        });
    }

    // --- Lock Card Toggle Logic ---
    const toggle = document.getElementById("lock-card-toggle");
    const knob = document.getElementById("lock-card-knob");
    const overlay = document.getElementById("card-lock-overlay");
    let isLocked = false;

    if (toggle) {
        toggle.addEventListener("click", () => {
            const newStatus = !isLocked ? 'blocked' : 'active';
            
            updateCardStatusAPI(newStatus, () => {
                isLocked = !isLocked;
                if (isLocked) {
                    // Switch on
                    toggle.classList.remove("bg-gray-200");
                    toggle.classList.add("bg-blue-600");
                    if (knob) {
                        knob.classList.remove("translate-x-1");
                        knob.classList.add("translate-x-6");
                    }
                    
                    // Show overlay
                    if (overlay) {
                        overlay.classList.remove("hidden");
                        setTimeout(() => {
                            overlay.classList.remove("opacity-0");
                            overlay.classList.add("opacity-100");
                        }, 10);
                    }
                    setCardStatus("Blocked");
                } else {
                    // Switch off
                    toggle.classList.remove("bg-blue-600");
                    toggle.classList.add("bg-gray-200");
                    if (knob) {
                        knob.classList.remove("translate-x-6");
                        knob.classList.add("translate-x-1");
                    }
                    
                    // Hide overlay
                    if (overlay) {
                        overlay.classList.remove("opacity-100");
                        overlay.classList.add("opacity-0");
                        setTimeout(() => {
                            overlay.classList.add("hidden");
                        }, 300);
                    }
                    setCardStatus("Active");
                }
            });
        });
    }

    // --- Generic Modal Logic ---
    function openModal(modal, modalContent) {
        modal.classList.remove('hidden');
        setTimeout(() => {
            modal.classList.add('opacity-100');
            modalContent.classList.add('scale-100', 'opacity-100');
            modalContent.classList.remove('scale-95', 'opacity-0');
        }, 10);
    }

    function closeModal(modal, modalContent) {
        modalContent.classList.add('scale-95', 'opacity-0');
        modalContent.classList.remove('scale-100', 'opacity-100');
        modal.classList.remove('opacity-100');
        setTimeout(() => {
            modal.classList.add('hidden');
        }, 300);
    }
    
    // --- Attach Listeners for Report Modal ---
    if (reportButton) {
        reportButton.addEventListener('click', (e) => {
            e.preventDefault();
            if (!reportButton.disabled) openModal(reportModal, reportModalContent);
        });
    }
    if (closeReportModalButton) closeReportModalButton.addEventListener('click', () => closeModal(reportModal, reportModalContent));
    if (cancelReportModalButton) cancelReportModalButton.addEventListener('click', () => closeModal(reportModal, reportModalContent));
    
    if (reportModal) {
        reportModal.addEventListener('click', (e) => {
            if (e.target === reportModal) closeModal(reportModal, reportModalContent);
        });
    }

    if (confirmReportButton) {
        confirmReportButton.addEventListener('click', () => {
            updateCardStatusAPI('lost', () => {
                reportButton.disabled = true;
                reportButton.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="mr-2 w-5 h-5"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg> Card Blocked';
                reportButton.classList.add('opacity-50', 'cursor-not-allowed');
                
                setCardStatus("Lost");
                closeModal(reportModal, reportModalContent);
            });
        });
    }

    // --- Attach Listeners for Replacement Modal ---
    if (replacementButton) {
        replacementButton.addEventListener('click', (e) => {
            e.preventDefault();
            if (!replacementButton.disabled) openModal(replacementModal, replacementModalContent);
        });
    }

    if (closeReplacementModalButton) closeReplacementModalButton.addEventListener('click', () => closeModal(replacementModal, replacementModalContent));
    if (cancelReplacementModalButton) cancelReplacementModalButton.addEventListener('click', () => closeModal(replacementModal, replacementModalContent));
    
    if (replacementModal) {
        replacementModal.addEventListener('click', (e) => {
            if (e.target === replacementModal) closeModal(replacementModal, replacementModalContent);
        });
    }

    if (confirmReplacementButton) {
        confirmReplacementButton.addEventListener('click', () => {
            requestReplacementAPI(() => {
                reportButton.disabled = true;
                reportButton.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="mr-2 w-5 h-5"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg> Card Blocked';
                reportButton.classList.add('opacity-50', 'cursor-not-allowed');
                
                replacementButton.disabled = true;
                replacementButton.textContent = 'Replacement Requested';
                replacementButton.classList.add('opacity-50', 'cursor-not-allowed');
                
                setCardStatus("Blocked");
                closeModal(replacementModal, replacementModalContent);
            });
        });
    }

    // --- Card Status Badge Logic ---
    function setCardStatus(status) {
        if (!cardStatusBadge) return;
        cardStatusBadge.classList.remove('bg-green-100', 'text-green-800', 'bg-red-100', 'text-red-800', 'bg-yellow-100', 'text-yellow-800', 'bg-gray-100', 'text-gray-800', 'bg-green-500', 'bg-red-500', 'bg-yellow-500', 'text-white');
        cardStatusBadge.className = "px-2 py-0.5 text-[9px] font-bold uppercase rounded-full shadow-sm";

        const lowerStatus = status.toLowerCase();
        if (lowerStatus === 'active') {
            cardStatusBadge.classList.add('bg-green-500', 'text-white');
            cardStatusBadge.textContent = 'Active';
        } else if (lowerStatus === 'blocked' || lowerStatus === 'lost' || lowerStatus === 'stolen') {
            cardStatusBadge.classList.add('bg-red-500', 'text-white');
            cardStatusBadge.textContent = status; // Keep original casing
        } else {
            cardStatusBadge.classList.add('bg-yellow-500', 'text-white');
            cardStatusBadge.textContent = status;
        }
    }

    // --- Initial fetch to set the toggle state properly ---
    function fetchCardStatus() {
        if (!userId) return;
        fetch(`/v1/user/${encodeURIComponent(userId)}`)
            .then(res => res.json())
            .then(data => {
                if (!data) return;
                const status = data.card_status || "Active";
                
                // If it's already blocked, sync the lock toggle visually
                if (status.toLowerCase() === "blocked") {
                    isLocked = true;
                    if (toggle) {
                        toggle.classList.remove("bg-gray-200");
                        toggle.classList.add("bg-blue-600");
                        if (knob) {
                            knob.classList.remove("translate-x-1");
                            knob.classList.add("translate-x-6");
                        }
                    }
                    if (overlay) {
                        overlay.classList.remove("hidden", "opacity-0");
                        overlay.classList.add("opacity-100");
                    }
                }

                // If it's lost, disable buttons
                if (status.toLowerCase() === "lost") {
                    if (reportButton) {
                        reportButton.disabled = true;
                        reportButton.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="mr-2 w-5 h-5"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg> Card Blocked';
                        reportButton.classList.add('opacity-50', 'cursor-not-allowed');
                    }
                    if (replacementButton) {
                        replacementButton.disabled = true;
                        replacementButton.textContent = 'Replacement Requested';
                        replacementButton.classList.add('opacity-50', 'cursor-not-allowed');
                    }
                }
            })
            .catch(err => console.error("Error fetching initial status", err));
    }

    fetchCardStatus();

});