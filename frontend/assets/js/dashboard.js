document.addEventListener("DOMContentLoaded", function () {
    const sidebar = document.getElementById('sidebar');
    const sidebarOverlay = document.getElementById('sidebar-overlay');
    const toggleButton = document.getElementById('menu-toggle-button');
    const openIcon = document.getElementById('icon-open');
    const closeIcon = document.getElementById('icon-close');
    const mainContent = document.getElementById('main-content');

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

        // Function to open the modal
        function openLogoutModal() {
            logoutModal.classList.remove('hidden');
            setTimeout(() => {
                logoutModal.classList.add('opacity-100');
                logoutModalContent.classList.add('scale-100', 'opacity-100');
                logoutModalContent.classList.remove('scale-95', 'opacity-0');
            }, 10);
        }

        // Function to close the modal
        function closeLogoutModal() {
            logoutModalContent.classList.add('scale-95', 'opacity-0');
            logoutModalContent.classList.remove('scale-100', 'opacity-100');
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
            window.location.href = "/login";
        });

    } else {
        console.error("Logout modal elements not found. Make sure all IDs are correct.");
    }

    // --- Fetch Dashboard Data ---
    function fetchDashboardData() {
        const pathSegments = window.location.pathname.split('/');
        let userId = null;
        if (pathSegments.length >= 2 && pathSegments[1] !== '') {
            userId = pathSegments[1];
        } else {
            // fallback if it's still somehow in query string
            const urlParams = new URLSearchParams(window.location.search);
            userId = urlParams.get('username');
        }

        let endpoint = "/v1/user/";
        if (userId) {
            endpoint += encodeURIComponent(userId);
        }

        fetch(endpoint)
            .then(response => {
                if (response.status === 401) {
                    window.location.href = "/login";
                    return null;
                }
                return response.json();
            })
            .then(data => {
                if (!data) return;

                const fullNameEl = document.getElementById("user-full-name");
                const initialsEl = document.getElementById("user-initials");
                const balanceEl = document.getElementById("user-balance");
                const loyaltyPointsEl = document.getElementById("user-loyalty-points");
                const accountTypeEl = document.getElementById("user-account-type");
                const transactionsBody = document.getElementById("transactions-table-body");

                const cardNoEl = document.getElementById("user-card-number");
                const cardHolderEl = document.getElementById("user-card-holder");
                const cardExpiryEl = document.getElementById("user-card-expiry");
                const cardStatusEl = document.getElementById("card-status-badge");

                // Profile Page Specific Elements
                const profileViewName = document.getElementById("profile-view-name");
                const profileViewEmail = document.getElementById("profile-view-email");
                const profileViewPhone = document.getElementById("profile-view-phone");
                const profileViewUsername = document.getElementById("profile-view-username");
                const profileEditName = document.getElementById("full_name");
                const profileEditEmail = document.getElementById("email");
                const profileEditPhone = document.getElementById("phone");

                // Card Page Specific Elements
                const cardDetailNumber = document.getElementById("card-detail-number");
                const cardDetailExpiry = document.getElementById("card-detail-expiry");

                if (fullNameEl) fullNameEl.innerText = data.name || "";
                if (initialsEl) initialsEl.innerText = data.initials || "U";
                if (balanceEl) balanceEl.innerText = Number(data.balance).toFixed(2);
                if (loyaltyPointsEl) loyaltyPointsEl.innerText = Number(data.loyalty_points).toFixed(2);
                if (accountTypeEl) accountTypeEl.innerText = data.account_type || "Regular";

                if (profileViewName) profileViewName.innerText = data.name || "";
                if (profileViewEmail) profileViewEmail.innerText = data.email || "";
                if (profileViewPhone) profileViewPhone.innerText = data.phone || "";
                if (profileViewUsername) profileViewUsername.innerText = data.username || "";

                if (profileEditName) profileEditName.value = data.name || "";
                if (profileEditEmail) profileEditEmail.value = data.email || "";
                if (profileEditPhone) profileEditPhone.value = data.phone || "";

                if (cardNoEl || cardDetailNumber) {
                    const rawNum = data.card_number || "••••••••••••••••";
                    const formattedNum = rawNum.replace(/(\d{4})/g, '$1 ').trim();
                    if (cardNoEl) cardNoEl.innerText = formattedNum || "•••• •••• •••• ••••";
                    if (cardDetailNumber) cardDetailNumber.innerText = formattedNum || "•••• •••• •••• ••••";
                }
                if (cardHolderEl) {
                    cardHolderEl.innerText = data.name || "CARDHOLDER NAME";
                }
                if (cardExpiryEl || cardDetailExpiry) {
                    if (cardExpiryEl) cardExpiryEl.innerText = data.card_expiry || "MM/YY";
                    if (cardDetailExpiry) cardDetailExpiry.innerText = data.card_expiry || "MM/YY";
                }
                if (cardStatusEl) {
                    const status = data.card_status || "No Card";
                    cardStatusEl.textContent = status;
                    cardStatusEl.className = "px-2 py-0.5 text-[9px] font-bold uppercase rounded-full shadow-sm";
                    if (status === "Active") {
                        cardStatusEl.classList.add("bg-green-500", "text-white");
                    } else if (status === "Blocked" || status === "Lost" || status === "Expired") {
                        cardStatusEl.classList.add("bg-red-500", "text-white");
                    } else {
                        cardStatusEl.classList.add("bg-yellow-500", "text-white");
                    }
                }

                if (transactionsBody) {
                    transactionsBody.innerHTML = "";
                    if (data.recent_transactions && data.recent_transactions.length > 0) {
                        data.recent_transactions.forEach(tx => {
                            const tr = document.createElement("tr");
                            const isPayment = tx.type === "Payment";
                            const colorClass = isPayment ? "text-red-600" : "text-green-600";
                            const sign = isPayment ? "-" : "+";
                            const amount = Number(tx.amount).toFixed(2);

                            tr.innerHTML = `
                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                    ${tx.date}
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                                    ${tx.description}
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                    ${tx.type}
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap text-sm ${colorClass} text-right font-medium">
                                    ${sign}₱${amount}
                                </td>
                            `;
                            transactionsBody.appendChild(tr);
                        });
                    } else {
                        transactionsBody.innerHTML = `
                            <tr>
                                <td colspan="4" class="px-6 py-10 text-center text-sm text-gray-500">
                                    No recent transactions found.
                                </td>
                            </tr>
                        `;
                    }
                }
            })
            .catch(error => {
                console.error("Error loading dashboard data:", error);
                const transactionsBody = document.getElementById("transactions-table-body");
                if (transactionsBody) {
                    transactionsBody.innerHTML = `
                        <tr>
                            <td colspan="4" class="px-6 py-10 text-center text-sm text-red-500">
                                Failed to load dashboard data.
                            </td>
                        </tr>
                    `;
                }
            });
    }

    // Call fetch on load
    fetchDashboardData();

});