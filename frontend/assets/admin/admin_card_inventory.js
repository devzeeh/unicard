document.addEventListener("DOMContentLoaded", function () {
    const adminUsername = window.location.pathname.split('/')[2];
    let allCards = [];

    // Fetch initial data
    fetch(`/v1/admin/${adminUsername}/card-inventory-data`)
        .then(response => response.json())
        .then(data => {
            if (data.success || data.stats) {
                // Populate Statistics
                document.getElementById('totalCards').textContent = data.stats.total.toLocaleString();
                document.getElementById('activeCards').textContent = data.stats.active.toLocaleString();
                document.getElementById('inactiveCards').textContent = data.stats.inactive.toLocaleString();
                document.getElementById('blockedCards').textContent = data.stats.blocked.toLocaleString();
                document.getElementById('lostCards').textContent = data.stats.lost.toLocaleString();

                allCards = data.cards || [];
                renderTable(allCards);
            } else {
                console.error("Failed to load card inventory data", data.message);
            }
        })
        .catch(error => console.error("Error fetching card inventory:", error));

    // Render table
    function renderTable(cards) {
        const tbody = document.getElementById('cardsTableBody');
        tbody.innerHTML = '';

        if (cards.length > 0) {
            cards.forEach(c => {
                const tr = document.createElement('tr');
                tr.className = 'hover:bg-gray-50 cursor-pointer transition duration-150';

                // Status badge styling
                let statusColor = 'bg-gray-100 text-gray-800';
                const statusLower = c.status.toLowerCase();
                if (statusLower === 'active') {
                    statusColor = 'bg-green-100 text-green-800';
                } else if (statusLower === 'inactive') {
                    statusColor = 'bg-gray-100 text-gray-800';
                } else if (statusLower === 'blocked') {
                    statusColor = 'bg-red-100 text-red-800';
                } else if (statusLower === 'lost') {
                    statusColor = 'bg-orange-100 text-orange-800';
                }

                // Safely handle user ID (could be null from database)
                const userIdDisplay = (c.user_id && c.user_id.Valid) ? c.user_id.String : '<span class="text-gray-400 italic">Unassigned</span>';

                // Row click listener to open modal
                tr.onclick = () => {
                    document.getElementById('modalCardNumber').textContent = c.card_number;
                    document.getElementById('modalCardUID').textContent = c.card_uid;
                    document.getElementById('modalUserID').innerHTML = userIdDisplay;
                    document.getElementById('modalCardType').textContent = c.card_type.replace(/_/g, ' ');
                    document.getElementById('modalBalance').textContent = '₱' + c.initial_amount.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
                    
                    const statusEl = document.getElementById('modalStatus');
                    statusEl.textContent = c.status;
                    statusEl.className = 'capitalize px-2 py-1 text-xs font-medium rounded-full ' + statusColor;
                    
                    document.getElementById('modalExpiryDate').textContent = c.expiry_date.split('T')[0];
                    document.getElementById('modalCreatedAt').textContent = new Date(c.created_at).toLocaleString();
                    
                    document.getElementById('cardDetailsModal').classList.remove('hidden');
                };

                tr.innerHTML = `
                    <td class="p-4 whitespace-nowrap">
                        <div class="text-sm font-medium text-gray-900">${c.card_number}</div>
                        <div class="text-xs text-gray-500">UID: ${c.card_uid}</div>
                    </td>
                    <td class="p-4 whitespace-nowrap text-sm text-gray-900">
                        ${userIdDisplay}
                    </td>
                    <td class="p-4 whitespace-nowrap text-sm text-gray-600 capitalize">
                        ${c.card_type.replace(/_/g, ' ')}
                    </td>
                    <td class="p-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        ₱${c.initial_amount.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                    </td>
                    <td class="p-4 whitespace-nowrap">
                        <span class="px-2 py-1 text-xs font-semibold rounded-full ${statusColor} capitalize">
                            ${c.status}
                        </span>
                    </td>
                    <td class="p-4 whitespace-nowrap text-sm text-gray-500">
                        ${new Date(c.created_at).toLocaleDateString()}
                    </td>
                    <td class="p-4 whitespace-nowrap text-center">
                        ${statusLower !== 'blocked' ? `<button class="block-btn px-3 py-1 bg-red-50 text-red-600 hover:bg-red-100 hover:text-red-700 rounded-md text-xs font-semibold transition" data-card-id="${c.card_number}">Block</button>` : `<span class="text-xs text-gray-400 font-medium">Blocked</span>`}
                    </td>
                `;
                tbody.appendChild(tr);
            });

            // Add listener to block buttons
            document.querySelectorAll('.block-btn').forEach(btn => {
                btn.onclick = (e) => {
                    e.stopPropagation(); // Prevent opening the modal
                    const cardId = e.target.getAttribute('data-card-id');
                    if(confirm("Are you sure you want to block this card? This action cannot be undone immediately.")) {
                        fetch(`/v1/admin/${adminUsername}/cards/${cardId}/block`, { method: 'POST' })
                            .then(res => res.json())
                            .then(data => {
                                if (data.success) {
                                    alert("Card blocked successfully.");
                                    // Update local state and re-render
                                    const cardIndex = allCards.findIndex(c => c.card_number === cardId);
                                    if(cardIndex !== -1) {
                                        allCards[cardIndex].status = "Blocked";
                                        applyFiltersAndRender();
                                    }
                                } else {
                                    alert("Failed to block card: " + data.message);
                                }
                            })
                            .catch(err => {
                                console.error(err);
                                alert("An error occurred while blocking the card.");
                            });
                    }
                };
            });

        } else {
            tbody.innerHTML = `
                <tr>
                    <td colspan="7" class="p-8 text-center">
                        <div class="flex flex-col items-center justify-center text-gray-500">
                            <svg class="w-10 h-10 mb-3 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"></path></svg>
                            <p class="text-base font-medium text-gray-900">No cards found</p>
                            <p class="text-sm text-gray-500 mt-1">There are currently no cards matching your criteria.</p>
                        </div>
                    </td>
                </tr>
            `;
        }

        // Update footer stats
        document.getElementById('pageStart').textContent = cards.length > 0 ? 1 : 0;
        document.getElementById('pageEnd').textContent = cards.length;
        document.getElementById('totalItems').textContent = cards.length;
    }

    // Filter and Sort Logic
    const searchInput = document.getElementById('searchInput');
    const filterStatus = document.getElementById('filterStatus');
    const filterType = document.getElementById('filterType');
    const sortOrder = document.getElementById('sortOrder');

    function applyFiltersAndRender() {
        const term = searchInput ? searchInput.value.toLowerCase().trim() : "";
        const statusVal = filterStatus ? filterStatus.value.toLowerCase() : "all";
        const typeVal = filterType ? filterType.value.toLowerCase() : "all";
        const sortVal = sortOrder ? sortOrder.value : "date_desc";

        // 1. Filter
        let filtered = allCards.filter(c => {
            const matchesSearch = !term || 
                c.card_number.toLowerCase().includes(term) ||
                c.card_uid.toLowerCase().includes(term) ||
                (c.user_id && c.user_id.Valid && c.user_id.String.toLowerCase().includes(term));
            
            const matchesStatus = (statusVal === "all") || (c.status.toLowerCase() === statusVal);
            const matchesType = (typeVal === "all") || (c.card_type.toLowerCase() === typeVal);

            return matchesSearch && matchesStatus && matchesType;
        });

        // 2. Sort
        filtered.sort((a, b) => {
            if (sortVal === 'date_desc') {
                return new Date(b.created_at) - new Date(a.created_at);
            } else if (sortVal === 'date_asc') {
                return new Date(a.created_at) - new Date(b.created_at);
            } else if (sortVal === 'balance_desc') {
                return b.initial_amount - a.initial_amount;
            } else if (sortVal === 'balance_asc') {
                return a.initial_amount - b.initial_amount;
            }
            return 0;
        });

        renderTable(filtered);
    }

    if (searchInput) searchInput.addEventListener('input', applyFiltersAndRender);
    if (filterStatus) filterStatus.addEventListener('change', applyFiltersAndRender);
    if (filterType) filterType.addEventListener('change', applyFiltersAndRender);
    if (sortOrder) sortOrder.addEventListener('change', applyFiltersAndRender);
});
