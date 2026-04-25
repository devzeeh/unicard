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
            mainContent.classList.toggle('md:pl-64');
            openIcon.classList.toggle('hidden');
            closeIcon.classList.toggle('hidden');
            if (window.innerWidth < 768) {
                sidebarOverlay.classList.toggle('hidden');
            }
        }

        toggleButton.addEventListener('click', function(e) {
            e.stopPropagation();
            toggleSidebar();
        });

        sidebarOverlay.addEventListener('click', function() {
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
        
        profileButton.addEventListener('click', function(event) {
            event.stopPropagation();
            profileMenu.classList.toggle('hidden');
        });

        document.addEventListener('click', function(event) {
            if (!profileMenu.classList.contains('hidden') && 
                !profileButton.contains(event.target) && 
                !profileMenu.contains(event.target)) 
            {
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
            window.location.href = "login.html";
        });

    } else {
        console.error("Logout modal elements not found. Make sure all IDs are correct.");
    }

});