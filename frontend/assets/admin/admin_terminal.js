window.renderAssignedTerminals = function(terminals) {
    if (!terminals || terminals.length === 0) {
        return '<span class="text-gray-400 italic">No terminals</span>';
    }
    return terminals.map(t => `<div class="mb-1"><span class="font-mono text-xs bg-gray-100 px-1 rounded">${t.terminal_sn}</span> - ${t.device_name}</div>`).join('');
};

document.addEventListener("DOMContentLoaded", () => {
    const tableBody = document.getElementById("terminal-table-body");
    if (!tableBody) return; // Exit if not on the terminal page

    const searchInput = document.getElementById("searchInput");
    const startItemSpan = document.getElementById("start-item");
    const endItemSpan = document.getElementById("end-item");
    const totalItemsSpan = document.getElementById("total-items");
    const prevPageBtn = document.getElementById("prev-page");
    const nextPageBtn = document.getElementById("next-page");
    const pageInfoSpan = document.getElementById("page-info");

    let currentPage = 1;
    const limit = 10;
    let currentSearch = "";
    let currentSortOrder = "desc";
    let currentStatus = "";

    function fetchTerminals(page = 1, search = "", sort = currentSortOrder, status = currentStatus) {
        const adminUsername = window.location.pathname.split('/')[2];
        const url = `/v1/admin/${adminUsername}/terminals-data?page=${page}&limit=${limit}&search=${encodeURIComponent(search)}&sort=${sort}&status=${status}`;

        fetch(url)
            .then(res => res.json())
            .then(data => {
                if (data.success) {
                    renderTable(data.data.terminals || []);
                    updatePagination(data.data.totalItems, data.data.page, data.data.limit);
                    
                    const activeCountEl = document.getElementById("activeTerminalCount");
                    const inactiveCountEl = document.getElementById("inactiveTerminalCount");
                    if (activeCountEl) activeCountEl.textContent = data.data.activeCount || 0;
                    if (inactiveCountEl) inactiveCountEl.textContent = data.data.inactiveCount || 0;
                } else {
                    console.error("Failed to fetch terminals:", data.message);
                }
            })
            .catch(err => console.error("Error fetching terminals:", err));
    }

    function renderTable(terminals) {
        tableBody.innerHTML = "";

        if (terminals.length === 0) {
            tableBody.innerHTML = `<tr><td colspan="5" class="px-6 py-8 text-center text-sm text-gray-500">No terminals found matching your criteria.</td></tr>`;
            return;
        }

        terminals.forEach(terminal => {
            const statusClass = terminal.status === 'active' || terminal.status === 'Online'
                ? 'bg-green-100 text-green-800' 
                : terminal.status === 'inactive' || terminal.status === 'Inactive'
                    ? 'bg-yellow-100 text-yellow-800'
                    : 'bg-red-100 text-red-800';

            const statusText = terminal.status.charAt(0).toUpperCase() + terminal.status.slice(1);
            
            // simple icon logic based on device name or type if available, else default to desktop
            const icon = terminal.device_name.toLowerCase().includes('rfid') || terminal.device_name.toLowerCase().includes('turnstile')
                ? 'fa-door-open' : 'fa-desktop';
            
            const iconBg = statusText === 'Online' || statusText === 'Active' ? 'bg-blue-100 text-blue-600' : 'bg-gray-100 text-gray-600';

            const row = document.createElement('tr');
            row.className = 'hover:bg-blue-50 cursor-pointer transition duration-150';
            row.onclick = (e) => {
                if (e.target.closest('button')) return; // Ignore button clicks
                document.getElementById('modalDeviceName').textContent = terminal.device_name;
                document.getElementById('modalTerminalId').textContent = terminal.terminal_id;
                document.getElementById('modalTerminalSn').textContent = terminal.terminal_sn;
                document.getElementById('modalAssignedMerchant').textContent = terminal.assigned_merchant;
                document.getElementById('modalLocationDetails').textContent = terminal.location_details || 'Not Set';
                
                const statusEl = document.getElementById('modalTerminalStatus');
                statusEl.textContent = statusText;
                statusEl.className = `capitalize px-2 py-1 text-xs font-medium rounded-full ${statusClass}`;
                
                document.getElementById('terminalDetailsModal').classList.remove('hidden');
            };

            row.innerHTML = `
                <td class="px-6 py-4 whitespace-nowrap">
                    <div class="text-sm font-medium text-gray-900 font-mono">${terminal.terminal_id}</div>
                    <div class="text-xs text-gray-500 font-mono">SN: ${terminal.terminal_sn}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600 max-w-[200px] truncate" title="${terminal.assigned_merchant}">${terminal.assigned_merchant}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600 max-w-[150px] truncate" title="${terminal.device_name}">${terminal.device_name}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600 max-w-[250px] truncate" title="${terminal.location_details || 'Not Set'}">${terminal.location_details || 'Not Set'}</td>
                <td class="px-6 py-4 whitespace-nowrap">
                    <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${statusClass}">${statusText}</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
                    <button class="text-indigo-600 hover:text-indigo-900">Ping</button>
                    <button class="text-gray-600 hover:text-gray-900">Settings</button>
                </td>
            `;
            tableBody.appendChild(row);
        });
    }

    function updatePagination(totalItems, page, currentLimit) {
        currentPage = page;
        
        const totalPages = Math.ceil(totalItems / currentLimit);
        const startItem = totalItems === 0 ? 0 : (page - 1) * currentLimit + 1;
        const endItem = Math.min(page * currentLimit, totalItems);

        startItemSpan.textContent = startItem;
        endItemSpan.textContent = endItem;
        totalItemsSpan.textContent = totalItems;
        
        if (totalItems === 0) {
            prevPageBtn.style.display = 'none';
            nextPageBtn.style.display = 'none';
            pageInfoSpan.style.display = 'none';
        } else {
            prevPageBtn.style.display = '';
            nextPageBtn.style.display = '';
            pageInfoSpan.style.display = '';
            pageInfoSpan.textContent = `Page ${page} of ${totalPages}`;
            prevPageBtn.disabled = page <= 1;
            nextPageBtn.disabled = page >= totalPages;
        }
    }

    prevPageBtn.addEventListener("click", () => {
        if (currentPage > 1) {
            fetchTerminals(currentPage - 1, currentSearch, currentSortOrder, currentStatus);
        }
    });

    nextPageBtn.addEventListener("click", () => {
        fetchTerminals(currentPage + 1, currentSearch, currentSortOrder, currentStatus);
    });

    // Debounce search input
    let searchTimeout;
    searchInput.addEventListener("input", (e) => {
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(() => {
            currentSearch = e.target.value.trim();
            fetchTerminals(1, currentSearch, currentSortOrder, currentStatus);
        }, 300);
    });

    const sortOrder = document.getElementById("sortOrder");
    if (sortOrder) {
        sortOrder.addEventListener("change", (e) => {
            currentSortOrder = e.target.value;
            fetchTerminals(1, currentSearch, currentSortOrder, currentStatus);
        });
    }

    const filterStatus = document.getElementById("filterStatus");
    if (filterStatus) {
        filterStatus.addEventListener("change", (e) => {
            currentStatus = e.target.value;
            fetchTerminals(1, currentSearch, currentSortOrder, currentStatus);
        });
    }

    const resetFiltersBtn = document.getElementById("resetFilters");
    if (resetFiltersBtn) {
        resetFiltersBtn.addEventListener("click", () => {
            if (searchInput) searchInput.value = "";
            if (sortOrder) sortOrder.value = "desc";
            if (filterStatus) filterStatus.value = "";
            
            currentSearch = "";
            currentSortOrder = "desc";
            currentStatus = "";
            fetchTerminals(1, currentSearch, currentSortOrder, currentStatus);
        });
    }

    const addTerminalForm = document.getElementById("addTerminalForm");
    if (addTerminalForm) {
        addTerminalForm.addEventListener("submit", (e) => {
            e.preventDefault();
            const formData = new FormData(addTerminalForm);
            
            let terminalSn = formData.get("terminalSnPrefix");
            if (terminalSn && !terminalSn.startsWith("UC-TRM-")) {
                terminalSn = "UC-TRM-" + terminalSn;
            }

            const data = {
                terminalSn: terminalSn,
                deviceName: formData.get("deviceName")
            };

            const alertBox = document.getElementById("terminalFormAlert");
            alertBox.classList.add("hidden");

            const adminUsername = window.location.pathname.split('/')[2];
            fetch(`/v1/admin/${adminUsername}/terminals/add`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            })
            .then(res => res.json())
            .then(result => {
                alertBox.classList.remove("hidden", "bg-red-50", "text-red-600", "bg-green-50", "text-green-600");
                if (result.success) {
                    alertBox.classList.add("bg-green-50", "text-green-600");
                    alertBox.textContent = result.message || "Terminal registered successfully!";
                    setTimeout(() => {
                        document.getElementById('addTerminalModal').classList.add('hidden');
                        addTerminalForm.reset();
                        alertBox.classList.add("hidden");
                        fetchTerminals(1, currentSearch, currentSortOrder, currentStatus);
                    }, 1500);
                } else {
                    alertBox.classList.add("bg-red-50", "text-red-600");
                    alertBox.textContent = result.message || "An error occurred.";
                }
            })
            .catch(err => {
                alertBox.classList.remove("hidden", "bg-green-50", "text-green-600");
                alertBox.classList.add("bg-red-50", "text-red-600");
                alertBox.textContent = "Network error. Please try again.";
            });
        });
    }

    // Initial fetch
    fetchTerminals();
});
