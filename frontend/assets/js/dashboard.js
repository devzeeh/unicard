document.addEventListener("DOMContentLoaded", function () {
    const sidebar = document.getElementById('sidebar');
    const sidebarOverlay = document.getElementById('sidebar-overlay');
    const toggleButton = document.getElementById('menu-toggle-button');
    const openIcon = document.getElementById('icon-open');
    const closeIcon = document.getElementById('icon-close');
    const mainContent = document.getElementById('main-content');

    const validTransactionTypes = new Set(['payment', 'top-up', 'refund', 'charge', 'deduction']);
    const shouldShowAmount = (type) => typeof type === 'string' && validTransactionTypes.has(type.toLowerCase());

    // --- Profile Dropdown Elements ---
    const profileButton = document.getElementById('profile-avatar-button');
    const profileMenu = document.getElementById('profile-dropdown-menu');
    const profileLogoutButton = document.getElementById('profile-logout-button'); // ADDED

    // --- Logout Modal Elements ---
    const logoutButton = document.getElementById('logout-button');
    const logoutModal = document.getElementById('logout-modal');
    const logoutModalContent = document.getElementById('logout-modal-content');
    const closeModalButton = document.getElementById('modal-close-button');
    const cancelModalButton = document.getElementById('modal-cancel-button');
    const confirmLogoutButton = document.getElementById('modal-confirm-logout-button');


    // --- Sidebar Logic ---
    if (sidebar && sidebarOverlay && toggleButton && openIcon && closeIcon && mainContent) {

        function toggleSidebar() {
            sidebar.classList.toggle('-translate-x-full');
            mainContent.classList.toggle('md:pl-72');
            openIcon.classList.toggle('hidden');
            closeIcon.classList.toggle('hidden');
            if (window.innerWidth < 768) {
                sidebarOverlay.classList.toggle('hidden');
            }
        }

        toggleButton.addEventListener('click', function (e) {
            e.stopPropagation();
            toggleSidebar();
        });

        sidebarOverlay.addEventListener('click', function () {
            toggleSidebar();
        });

        // --- Auto-close sidebar on nav link click (for mobile) ---
        const navLinks = sidebar.querySelectorAll('nav a');
        navLinks.forEach(link => {
            link.addEventListener('click', () => {
                if (window.innerWidth < 768 && !closeIcon.classList.contains('hidden')) {
                    toggleSidebar();
                }
            });
        });

    } else {
        console.error("Sidebar elements not found. Make sure all IDs are correct.");
    }

    // --- Profile Dropdown Logic ---
    if (profileButton && profileMenu) {

        profileButton.addEventListener('click', function (event) {
            event.stopPropagation();
            profileMenu.classList.toggle('hidden');
        });

        document.addEventListener('click', function (event) {
            if (!profileMenu.classList.contains('hidden') &&
                !profileButton.contains(event.target) &&
                !profileMenu.contains(event.target)) {
                profileMenu.classList.add('hidden');
            }
        });

    } else {
        console.error("Profile dropdown elements not found. Make sure all IDs are correct.");
    }

    // --- Logout Modal Logic ---
    // Check for all required modal elements
    const modalElementsExist = logoutModal && logoutModalContent && closeModalButton && cancelModalButton && confirmLogoutButton;

    if (modalElementsExist) {

        function openLogoutModal() {
            logoutModal.classList.remove('hidden');
            setTimeout(() => {
                logoutModal.classList.remove('opacity-0');
                logoutModal.classList.add('opacity-100');
                logoutModalContent.classList.add('scale-100', 'opacity-100');
                logoutModalContent.classList.remove('scale-95', 'opacity-0');
            }, 10);
        }

        // Function to close the modal
        function closeLogoutModal() {
            logoutModalContent.classList.add('scale-95', 'opacity-0');
            logoutModalContent.classList.remove('scale-100', 'opacity-100');
            logoutModal.classList.add('opacity-0');
            logoutModal.classList.remove('opacity-100');

            setTimeout(() => {
                logoutModal.classList.add('hidden');
            }, 300);
        }

        // --- UPDATED: Attach to all logout buttons ---

        // 1. Sidebar Logout Button
        if (logoutButton) {
            logoutButton.addEventListener('click', (e) => {
                e.preventDefault();
                openLogoutModal();
            });
        }

        // 2. Profile Dropdown Logout Button
        if (profileLogoutButton) {
            profileLogoutButton.addEventListener('click', (e) => {
                e.preventDefault();
                profileMenu.classList.add('hidden'); // Close dropdown
                openLogoutModal();
            });
        }
        // --- END OF UPDATE ---

        // Close modal buttons
        closeModalButton.addEventListener('click', closeLogoutModal);
        cancelModalButton.addEventListener('click', closeLogoutModal);

        // Also close if clicking on the background overlay
        logoutModal.addEventListener('click', (e) => {
            if (e.target === logoutModal) {
                closeLogoutModal();
            }
        });

        // Confirm logout and redirect
        confirmLogoutButton.addEventListener('click', () => {
            console.log('Logging out...');
            window.location.href = "/logout";
        });

    } else {
        console.error("Logout modal elements not found. Make sure all IDs are correct.");
    }

    // --- Fetch Dashboard Data (Removed for SSR) ---

    // --- Modal Logic ---
    const txnModal = document.getElementById("txnModal");
    const txnModalContent = document.getElementById("txnModalContent");
    const closeTxnModalBtn = document.getElementById("closeTxnModalBtn");
    const closeTxnModalBottomBtn = document.getElementById("closeTxnModalBottomBtn");

    if (txnModal && closeTxnModalBtn) {
        closeTxnModalBtn.onclick = closeTxnModal;
        closeTxnModalBottomBtn.onclick = closeTxnModal;
        txnModal.onclick = function (e) {
            if (e.target === txnModal) closeTxnModal();
        };
    }

    function openTxnModal(tx) {
        if (!txnModal) return;

        document.getElementById("modalTxnId").textContent = tx.transaction_id || 'N/A';
        document.getElementById("modalTxnDate").textContent = `${tx.date} at ${tx.time}`;
        document.getElementById("modalTxnType").textContent = tx.type ? tx.type.charAt(0).toUpperCase() + tx.type.slice(1) : "N/A";

        document.getElementById("modalTxnMerchant").textContent = tx.merchant_name || 'N/A';
        const terminalIdEl = document.getElementById("modalTxnTerminalId");
        const terminalLabelEl = document.getElementById("modalTxnTerminalLabel");
        if (tx.terminal_id && tx.terminal_id !== 'N/A' && tx.terminal_id.trim() !== '') {
            terminalIdEl.textContent = tx.terminal_id;
            if (terminalLabelEl) terminalLabelEl.classList.remove("hidden");
        } else {
            terminalIdEl.textContent = '';
            if (terminalLabelEl) terminalLabelEl.classList.add("hidden");
        }

        document.getElementById("modalTxnFee").textContent = `₱${Number(tx.service_fee || 0).toFixed(2)}`;

        if (tx.points_earned && tx.points_earned > 0) {
            document.getElementById("modalTxnPoints").textContent = `+${tx.points_earned}`;
        } else {
            document.getElementById("modalTxnPoints").textContent = "0";
        }

        document.getElementById("modalTxnDesc").textContent = tx.description || 'N/A';

        const merchantRow = document.getElementById("modalTxnMerchantRow");
        if (tx.type && tx.type.toLowerCase() === "topup") {
            merchantRow.classList.remove("flex");
            merchantRow.classList.add("hidden");
        } else {
            merchantRow.classList.remove("hidden");
            merchantRow.classList.add("flex");
        }

        const isPayment = tx.type && tx.type.toLowerCase() === "payment";
        const sign = isPayment ? "-" : "+";
        const colorClass = isPayment ? "text-red-600" : "text-green-600";

        const amtEl = document.getElementById("modalTxnTotal");
        amtEl.textContent = `${sign}₱${Number(tx.amount).toFixed(2)}`;
        amtEl.className = `font-bold text-lg ${colorClass}`;

        txnModal.classList.remove('hidden');
        setTimeout(() => {
            txnModal.classList.add('opacity-100');
            txnModalContent.classList.add('scale-100', 'opacity-100');
            txnModalContent.classList.remove('scale-95', 'opacity-0');
        }, 10);
    }

    function closeTxnModal() {
        txnModalContent.classList.add('scale-95', 'opacity-0');
        txnModalContent.classList.remove('scale-100', 'opacity-100');
        txnModal.classList.remove('opacity-100');
        setTimeout(() => {
            txnModal.classList.add('hidden');
        }, 300);
    }
});