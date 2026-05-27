// Modal Control Functions
function openDeactivateModal(num, holder, type) {
    document.getElementById('deactivate-card-number').value = num;
    document.getElementById('deactivate-card-holder').value = holder;
    document.getElementById('deactivate-card-type').value = type;

    document.getElementById('deactivate-modal-num').innerText = num;
    document.getElementById('deactivate-modal-holder').innerText = holder;
    document.getElementById('deactivate-modal-type').innerText = type;

    document.getElementById('deactivate-modal').classList.remove('hidden');
}

function closeDeactivateModal() {
    document.getElementById('deactivate-modal').classList.add('hidden');
}

function openDeleteModal(num) {
    document.getElementById('delete-card-number').value = num;
    document.getElementById('delete-modal-num').innerText = num;
    document.getElementById('delete-modal').classList.remove('hidden');
}

function closeDeleteModal() {
    document.getElementById('delete-modal').classList.add('hidden');
}

function openLogoutModal() {
    document.getElementById('logout-modal').classList.remove('hidden');
}

function closeLogoutModal() {
    document.getElementById('logout-modal').classList.add('hidden');
}

// Pagination and Search Configuration
const rowsPerPage = 4;
let currentPage = 1;
let allCards = [];
let filteredCards = [];

function renderCards() {
    const totalRows = filteredCards.length;
    const totalPages = Math.ceil(totalRows / rowsPerPage);

    if (currentPage > totalPages) {
        currentPage = Math.max(1, totalPages);
    }

    const startIdx = (currentPage - 1) * rowsPerPage;
    const endIdx = Math.min(startIdx + rowsPerPage, totalRows);

    const tbody = document.getElementById('table-body');
    tbody.innerHTML = '';

    const pageCards = filteredCards.slice(startIdx, endIdx);
    if (pageCards.length === 0) {
        tbody.innerHTML = `
                    <tr>
                        <td colspan="7" class="px-6 py-10 text-center text-sm text-gray-500">
                            No cards registered or found.
                        </td>
                    </tr>
                `;
    } else {
        pageCards.forEach(c => {
            const tr = document.createElement('tr');
            tr.className = 'hover:bg-gray-50 transition duration-150';

            let holderHtml = '';
            if (c.card_holder === 'Unlinked') {
                holderHtml = '<span class="text-gray-400 italic">Unlinked</span>';
            } else {
                holderHtml = c.card_holder;
            }

            let statusClass = '';
            if (c.status === 'Active') {
                statusClass = 'bg-green-100 text-green-800';
            } else if (c.status === 'Inactive') {
                statusClass = 'bg-gray-100 text-gray-800';
            } else if (c.status === 'Blocked') {
                statusClass = 'bg-red-100 text-red-800';
            } else if (c.status === 'Lost') {
                statusClass = 'bg-yellow-100 text-yellow-800';
            } else {
                statusClass = 'bg-red-100 text-red-800';
            }

            const canDeactivate = c.status === 'Active';
            const deactivateBtn = canDeactivate
                ? `<button type="button" onclick="openDeactivateModal('${c.card_number}', '${c.card_holder}', '${c.card_type}')" class="text-red-600 hover:text-red-900 font-semibold transition duration-150">Deactivate</button>`
                : `<span class="text-gray-300 cursor-not-allowed">Deactivate</span>`;

            tr.innerHTML = `
                        <td class="px-6 py-4 whitespace-nowrap text-sm font-semibold text-gray-900 tracking-wider">${c.card_number}</td>
                        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-700">${holderHtml}</td>
                        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">${c.card_type}</td>
                        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">${c.expiry_date}</td>
                        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">₱${Number(c.initial_amount).toFixed(2)}</td>
                        <td class="px-6 py-4 whitespace-nowrap text-sm">
                            <span class="px-3 py-1 inline-flex text-xs font-semibold rounded-full ${statusClass}">
                                ${c.status}
                            </span>
                        </td>
                        <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-3">
                            ${deactivateBtn}
                            <button type="button" onclick="openDeleteModal('${c.card_number}')" class="text-red-600 hover:text-red-950 font-semibold transition duration-150">
                                Delete
                            </button>
                        </td>
                    `;
            tbody.appendChild(tr);
        });
    }

    // Update footer metrics text
    document.getElementById('start-index').innerText = totalRows === 0 ? 0 : startIdx + 1;
    document.getElementById('end-index').innerText = endIdx;
    document.getElementById('total-count').innerText = totalRows;

    // Enable/disable page control buttons
    const prevBtn = document.getElementById('prev-btn');
    const nextBtn = document.getElementById('next-btn');
    const prevBtnMobile = document.getElementById('prev-btn-mobile');
    const nextBtnMobile = document.getElementById('next-btn-mobile');

    if (prevBtn) prevBtn.disabled = currentPage === 1;
    if (nextBtn) nextBtn.disabled = currentPage === totalPages || totalPages === 0;
    if (prevBtnMobile) prevBtnMobile.disabled = currentPage === 1;
    if (nextBtnMobile) nextBtnMobile.disabled = currentPage === totalPages || totalPages === 0;
}

// Search Filtering Logic
function filterCards() {
    const query = document.getElementById('search-input').value.toLowerCase().trim();

    if (query === '') {
        filteredCards = [...allCards];
    } else {
        filteredCards = allCards.filter(c => {
            const searchString = `${c.card_number} ${c.card_holder} ${c.card_type} ${c.status}`.toLowerCase();
            return searchString.includes(query);
        });
    }

    renderCards();
}

function fetchDashboardData(showAlert = null) {
    fetch("/v1/admin/dashboard-data")
        .then(res => {
            if (res.status === 401) {
                window.location.href = "/login";
                return null;
            }
            return res.json();
        })
        .then(data => {
            if (!data) return;

            // Update stats
            document.getElementById('stats-total').innerText = data.stats.total || 0;
            document.getElementById('stats-active').innerText = data.stats.active || 0;
            document.getElementById('stats-inactive').innerText = data.stats.inactive || 0;
            document.getElementById('stats-blocked').innerText = data.stats.blocked || 0;
            document.getElementById('stats-lost').innerText = data.stats.lost || 0;

            // Update cards list
            allCards = data.cards || [];

            // Show success/error if passed
            if (showAlert) {
                const errorAlert = document.getElementById("error-alert");
                const successAlert = document.getElementById("success-alert");
                const errorText = document.getElementById("error-text");
                const successText = document.getElementById("success-text");

                if (errorAlert) errorAlert.classList.add("hidden");
                if (successAlert) successAlert.classList.add("hidden");

                if (showAlert.type === 'success') {
                    if (successAlert && successText) {
                        successText.innerText = showAlert.message;
                        successAlert.classList.remove("hidden");
                    }
                } else if (showAlert.type === 'error') {
                    if (errorAlert && errorText) {
                        errorText.innerText = showAlert.message;
                        errorAlert.classList.remove("hidden");
                    }
                }
            }

            // Apply filters & render
            filterCards();
        })
        .catch(err => {
            console.error("Error loading dashboard data:", err);
        });
}

// Set up search event listener
document.getElementById('search-input').addEventListener('input', () => {
    currentPage = 1;
    filterCards();
});

// Set up pagination control event listeners
document.getElementById('prev-btn').addEventListener('click', () => {
    if (currentPage > 1) {
        currentPage--;
        renderCards();
    }
});

document.getElementById('next-btn').addEventListener('click', () => {
    const totalPages = Math.ceil(filteredCards.length / rowsPerPage);
    if (currentPage < totalPages) {
        currentPage++;
        renderCards();
    }
});

document.getElementById('prev-btn-mobile').addEventListener('click', () => {
    if (currentPage > 1) {
        currentPage--;
        renderCards();
    }
});

document.getElementById('next-btn-mobile').addEventListener('click', () => {
    const totalPages = Math.ceil(filteredCards.length / rowsPerPage);
    if (currentPage < totalPages) {
        currentPage++;
        renderCards();
    }
});

// Setup confirm event listeners
document.getElementById('confirm-deactivate-btn').addEventListener('click', () => {
    const cardNumber = document.getElementById('deactivate-card-number').value;
    const name = document.getElementById('deactivate-card-holder').value;
    const cardType = document.getElementById('deactivate-card-type').value;

    fetch('/v1/admin/deactivatecardauth', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ cardNumber, name, cardType })
    })
        .then(res => res.json())
        .then(data => {
            closeDeactivateModal();
            if (data.success) {
                fetchDashboardData({ type: 'success', message: data.message });
            } else {
                fetchDashboardData({ type: 'error', message: data.message });
            }
        })
        .catch(err => {
            console.error(err);
            closeDeactivateModal();
            fetchDashboardData({ type: 'error', message: 'Failed to deactivate card due to a network error.' });
        });
});

document.getElementById('confirm-delete-btn').addEventListener('click', () => {
    const cardNumber = document.getElementById('delete-card-number').value;

    fetch('/v1/admin/deletecardauth', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ cardNumber })
    })
        .then(res => res.json())
        .then(data => {
            closeDeleteModal();
            if (data.success) {
                fetchDashboardData({ type: 'success', message: data.message });
            } else {
                fetchDashboardData({ type: 'error', message: data.message });
            }
        })
        .catch(err => {
            console.error(err);
            closeDeleteModal();
            fetchDashboardData({ type: 'error', message: 'Failed to delete card due to a network error.' });
        });
});

function fetchAdminProfile() {
    fetch("/v1/admin/me")
        .then(res => {
            if (res.status === 401) {
                window.location.href = "/login";
                return null;
            }
            return res.json();
        })
        .then(data => {
            if (!data || !data.username) return;
            const username = data.username;
            const welcomeEl = document.getElementById('admin-welcome-name');
            if (welcomeEl) {
                welcomeEl.innerText = `Welcome, ${username}`;
            }
            const avatarEl = document.getElementById('admin-avatar');
            if (avatarEl) {
                avatarEl.innerText = username.substring(0, 2).toUpperCase();
            }
        })
        .catch(err => console.error("Error fetching admin profile:", err));
}

// Initialize display by fetching JSON
document.addEventListener("DOMContentLoaded", () => {
    fetchAdminProfile();
    fetchDashboardData();
});