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
        return;
    }

    // --- Generic Modal Logic ---
    // Function to open a modal
    function openModal(modal, modalContent) {
        modal.classList.remove('hidden');
        // Animate in
        setTimeout(() => {
            modal.classList.add('opacity-100');
            modalContent.classList.add('scale-100', 'opacity-100');
            modalContent.classList.remove('scale-95', 'opacity-0');
        }, 10); // 10ms delay to allow CSS to catch up
    }

    // Function to close a modal
    function closeModal(modal, modalContent) {
        // Animate out
        modalContent.classList.add('scale-95', 'opacity-0');
        modalContent.classList.remove('scale-100', 'opacity-100');
        modal.classList.remove('opacity-100');
        
        // Hide after animation (300ms)
        setTimeout(() => {
            modal.classList.add('hidden');
        }, 300);
    }

    
    // --- Attach Listeners for Report Modal ---
    reportButton.addEventListener('click', (e) => {
        e.preventDefault();
        openModal(reportModal, reportModalContent);
    });

    closeReportModalButton.addEventListener('click', () => closeModal(reportModal, reportModalContent));
    cancelReportModalButton.addEventListener('click', () => closeModal(reportModal, reportModalContent));
    
    reportModal.addEventListener('click', (e) => {
        if (e.target === reportModal) {
            closeModal(reportModal, reportModalContent);
        }
    });

    confirmReportButton.addEventListener('click', () => {
        console.log('Reporting card as lost/stolen...');
        
        // --- THIS IS A FRONT-END-ONLY DEMO ---
        // In a real app, you would send a fetch() request to your Go backend here
        // to update the card_blocks table and the user's status.
        
        // --- Logic Update ---
        // Disable the "Report" button, but leave "Request Replacement" active.
        reportButton.disabled = true;
        reportButton.innerHTML = '<i class="fas fa-check mr-2"></i> Card Blocked';
        reportButton.classList.add('opacity-50', 'cursor-not-allowed');
        
        // Update the status badge
        setCardStatus("Blocked");
        
        closeModal(reportModal, reportModalContent);
    });

    
    // --- Attach Listeners for Replacement Modal ---
    replacementButton.addEventListener('click', (e) => {
        e.preventDefault();
        openModal(replacementModal, replacementModalContent);
    });

    closeReplacementModalButton.addEventListener('click', () => closeModal(replacementModal, replacementModalContent));
    cancelReplacementModalButton.addEventListener('click', () => closeModal(replacementModal, replacementModalContent));
    
    replacementModal.addEventListener('click', (e) => {
        if (e.target === replacementModal) {
            closeModal(replacementModal, replacementModalContent);
        }
    });

    confirmReplacementButton.addEventListener('click', () => {
        console.log('Requesting new card...');
        
        // --- THIS IS A FRONT-END-ONLY DEMO ---
        // In a real app, you would send a fetch() request to your Go backend here
        // to deduct the fee and log the replacement.
        
        // --- Logic Update ---
        // Disable BOTH buttons since a replacement has been requested.
        reportButton.disabled = true;
        reportButton.innerHTML = '<i class="fas fa-check mr-2"></i> Card Blocked';
        reportButton.classList.add('opacity-50', 'cursor-not-allowed');
        
        replacementButton.disabled = true;
        replacementButton.textContent = 'Replacement Requested';
        replacementButton.classList.add('opacity-50', 'cursor-not-allowed');
        
        // Update the status badge
        setCardStatus("Replaced");
        
        closeModal(replacementModal, replacementModalContent);
    });


    // --- Card Status Badge Logic ---
    function setCardStatus(status) {
        // Clear all existing color classes
        cardStatusBadge.classList.remove('bg-green-100', 'text-green-800', 'bg-red-100', 'text-red-800', 'bg-yellow-100', 'text-yellow-800', 'bg-gray-100', 'text-gray-800');

        switch (status) {
            case 'Active':
                cardStatusBadge.classList.add('bg-green-100', 'text-green-800');
                cardStatusBadge.textContent = 'Active';
                break;
            case 'Blocked':
            case 'Stolen':
            case 'Lost':
                cardStatusBadge.classList.add('bg-red-100', 'text-red-800');
                cardStatusBadge.textContent = 'Blocked';
                break;
            case 'Replaced':
            case 'Inactive':
                cardStatusBadge.classList.add('bg-yellow-100', 'text-yellow-800');
                cardStatusBadge.textContent = 'Inactive';
                break;
            default:
                cardStatusBadge.classList.add('bg-gray-100', 'text-gray-800');
                cardStatusBadge.textContent = 'Unknown';
        }
    }

    // --- This function simulates fetching the user's card status from the backend ---
    function fetchCardStatus() {
        // --- FRONTEND-ONLY DEMO ---
        // In a real app, you would make a fetch() call to your Go backend here,
        // get the user's real card status from the 'users' table,
        // and then call setCardStatus() with the result.
        
        // For this demo, we'll just set it to "Active".
        const currentDemoStatus = "Active"; // You can change this to "Blocked" to test
        setCardStatus(currentDemoStatus);

        // Also, update the button states based on the fetched status
        if (currentDemoStatus === "Blocked" || currentDemoStatus === "Replaced" || currentDemoStatus === "Inactive") {
             reportButton.disabled = true;
             reportButton.innerHTML = '<i class="fas fa-check mr-2"></i> Card Blocked';
             reportButton.classList.add('opacity-50', 'cursor-not-allowed');
        }
        if (currentDemoStatus === "Replaced") {
            replacementButton.disabled = true;
            replacementButton.textContent = 'Replacement Requested';
            replacementButton.classList.add('opacity-50', 'cursor-not-allowed');
        }
    }

    // --- Run the functions on page load ---
    fetchCardStatus();

});