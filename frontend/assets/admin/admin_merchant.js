let currentPage = 1; // current page
const itemsPerPage = 10;    // items per page
let currentSearchQuery = '';
let currentSortOrder = 'desc';
let currentCategory = '';
let currentStatus = '';
let totalItemsCount = 0;
let currentMerchants = [];
let unassignedTerminals = [];

window.renderAssignedTerminals = function (terminals) {
    if (!terminals || terminals.length === 0) {
        return '<span class="text-gray-400 italic">No terminals</span>';
    }
    return terminals.map(t => {
        return `<div class="mb-1 max-w-[200px]" title="${t.device_name || t.terminal_id}">
            <div class="text-sm text-gray-900 truncate">${t.device_name || t.terminal_id}</div>
            <div class="text-xs text-gray-500 truncate">SN: ${t.terminal_sn}</div>
        </div>`;
    }).join('');
};

function fetchUnassignedTerminals() {
    const adminUsername = window.location.pathname.split('/')[2];
    fetch(`/v1/admin/${adminUsername}/terminals/unassigned`)
        .then(res => res.json())
        .then(result => {
            if (result.success && result.data) {
                unassignedTerminals = result.data;
                document.querySelectorAll('.terminal-sn-select').forEach(populateTerminalDropdown);
            }
        })
        .catch(error => console.error("Error fetching unassigned terminals", error));
}

function populateTerminalDropdown(selectElement) {
    selectElement.innerHTML = '<option value="" disabled selected>Select a terminal</option>';
    unassignedTerminals.forEach(t => {
        const opt = document.createElement('option');
        opt.value = t.terminal_sn;
        opt.textContent = t.terminal_sn;
        opt.dataset.deviceName = t.device_name;
        selectElement.appendChild(opt);
    });
}

function fetchMerchants() {
    const queryParams = new URLSearchParams({
        page: currentPage,
        limit: itemsPerPage,
        search: currentSearchQuery,
        sort: currentSortOrder,
        category: currentCategory,
        status: currentStatus
    });

    const adminUsername = window.location.pathname.split('/')[2];
    fetch(`/v1/admin/${adminUsername}/merchants-data?${queryParams.toString()}`)
        .then(response => response.json())
        .then(result => {
            if (result.success && result.data) {
                currentMerchants = result.data.merchants || [];
                totalItemsCount = result.data.totalItems || 0;
                renderTable();
            }
        })
        .catch(error => console.error('Error fetching merchants:', error));
}

function applyFiltersAndSort() {
    currentPage = 1;
    fetchMerchants();
}

function highlightText(text, queryTerms) {
    if (!text) return '';
    if (!queryTerms || queryTerms.length === 0) {
        const div = document.createElement('div');
        div.innerText = text;
        return div.innerHTML;
    }

    const escapedTerms = queryTerms.filter(t => t.length > 0).map(term => term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'));
    if (escapedTerms.length === 0) {
        const div = document.createElement('div');
        div.innerText = text;
        return div.innerHTML;
    }

    const regex = new RegExp(`(${escapedTerms.join('|')})`, 'gi');

    const div = document.createElement('div');
    div.innerText = text;
    const safeText = div.innerHTML;

    return safeText.replace(regex, '<mark class="bg-yellow-200 text-gray-900 rounded-sm px-0.5">$1</mark>');
}

function renderTable() {
    const tbody = document.getElementById('merchantTableBody');
    tbody.innerHTML = '';

    if (!currentMerchants || currentMerchants.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" class="px-6 py-8 text-center text-sm text-gray-500">No merchants found matching your criteria.</td></tr>`;
        document.getElementById('pageStart').textContent = '0';
        document.getElementById('pageEnd').textContent = '0';
        document.getElementById('totalItems').textContent = '0';
        document.getElementById('paginationControls').innerHTML = '';
        return;
    }

    const queryTerms = currentSearchQuery ? currentSearchQuery.toLowerCase().trim().split(/\s+/) : [];

    currentMerchants.forEach(merchant => {
        const tr = document.createElement('tr');

        const highlightedBusinessName = highlightText(merchant.business_name, queryTerms);
        const highlightedMerchantId = highlightText(merchant.merchant_id, queryTerms);
        const highlightedOwnerName = highlightText(merchant.owner_name, queryTerms);
        const highlightedEmail = highlightText(merchant.business_email, queryTerms);

        tr.className = 'hover:bg-gray-50 cursor-pointer transition duration-150';
        tr.onclick = (e) => {
            if (e.target.closest('button')) return; // Ignore button clicks
            const adminUsername = window.location.pathname.split('/')[2];
            window.location.href = `/admin/${adminUsername}/merchants/${merchant.merchant_id}`;
        };

        tr.innerHTML = `
            <td class="px-6 py-4 whitespace-nowrap max-w-[250px]" title="${merchant.business_name}">
                <div class="text-sm font-medium text-gray-900 truncate">${highlightedBusinessName}</div>
                <div class="text-xs text-gray-500 truncate">ID: ${highlightedMerchantId}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600 capitalize max-w-[150px] truncate" title="${merchant.business_type}">${merchant.business_type.replace(/_/g, ' ')}</td>
            <td class="px-6 py-4 whitespace-nowrap max-w-[250px]" title="${merchant.owner_name} / ${merchant.business_email}">
                <div class="text-sm text-gray-900 truncate">${highlightedOwnerName}</div>
                <div class="text-xs text-gray-500 truncate">${highlightedEmail}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm text-gray-900">
                    ${window.renderAssignedTerminals ? window.renderAssignedTerminals(merchant.terminals) : '<span class="text-gray-400 italic">No terminals</span>'}
                </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${merchant.status.toLowerCase() === 'active' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'} capitalize">${merchant.status}</span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
                <button class="text-indigo-600 hover:text-indigo-900">Edit</button>
            </td>
        `;
        tbody.appendChild(tr);
    });

    const startIndex = (currentPage - 1) * itemsPerPage;
    const endIndex = Math.min(startIndex + itemsPerPage, totalItemsCount);

    document.getElementById('pageStart').textContent = totalItemsCount > 0 ? startIndex + 1 : 0;
    document.getElementById('pageEnd').textContent = endIndex;
    document.getElementById('totalItems').textContent = totalItemsCount;

    renderPagination();
}

function renderPagination() {
    const paginationControls = document.getElementById('paginationControls');
    paginationControls.innerHTML = '';

    const totalPages = Math.ceil(totalItemsCount / itemsPerPage);

    if (totalPages <= 1) return;

    const prevBtn = document.createElement('button');
    prevBtn.className = `px-3 py-1 rounded-md text-sm font-medium ${currentPage === 1 ? 'text-gray-400 cursor-not-allowed' : 'text-blue-600 hover:bg-blue-50'}`;
    prevBtn.textContent = 'Previous';
    prevBtn.disabled = currentPage === 1;
    prevBtn.onclick = () => { if (currentPage > 1) { currentPage--; fetchMerchants(); } };
    paginationControls.appendChild(prevBtn);

    for (let i = 1; i <= totalPages; i++) {
        const btn = document.createElement('button');
        btn.className = `px-3 py-1 rounded-md text-sm font-medium ${currentPage === i ? 'bg-blue-600 text-white' : 'text-gray-700 hover:bg-gray-50'}`;
        btn.textContent = i;
        btn.onclick = () => { currentPage = i; fetchMerchants(); };
        paginationControls.appendChild(btn);
    }

    const nextBtn = document.createElement('button');
    nextBtn.className = `px-3 py-1 rounded-md text-sm font-medium ${currentPage === totalPages ? 'text-gray-400 cursor-not-allowed' : 'text-blue-600 hover:bg-blue-50'}`;
    nextBtn.textContent = 'Next';
    nextBtn.disabled = currentPage === totalPages;
    nextBtn.onclick = () => { if (currentPage < totalPages) { currentPage++; fetchMerchants(); } };
    paginationControls.appendChild(nextBtn);
}

document.addEventListener('DOMContentLoaded', () => {
    fetchMerchants();
    fetchUnassignedTerminals();

    // Delegate change event for terminal selection
    const container = document.getElementById('merchantBlocksContainer');
    container.addEventListener('change', function (e) {
        if (e.target.classList.contains('terminal-sn-select')) {
            const selectedOption = e.target.options[e.target.selectedIndex];
            const block = e.target.closest('.merchant-block');
            const deviceNameInput = block.querySelector('.device-name-input');
            if (selectedOption && selectedOption.dataset.deviceName) {
                deviceNameInput.value = selectedOption.dataset.deviceName;
            } else {
                deviceNameInput.value = '';
            }
        }
    });

    // Auto-format fields on focus out to trim spaces and format as Title Case
    container.addEventListener('focusout', function (e) {
        if (e.target.tagName === 'INPUT') {
            let val = e.target.value;
            if (e.target.type === 'text' && !['businessPhone', 'commissionRate', 'registrationNum', 'deviceName'].includes(e.target.name)) {
                e.target.value = val.trim().toLowerCase().replace(/\b\w/g, c => c.toUpperCase());
            } else if (e.target.type === 'email') {
                e.target.value = val.trim().toLowerCase();
            } else {
                e.target.value = val.trim();
            }
        }
    });

    const searchInput = document.getElementById('searchInput');
    let searchTimeout = null;
    if (searchInput) {
        searchInput.addEventListener('input', (e) => {
            currentSearchQuery = e.target.value;
            // Debounce search
            if (searchTimeout) clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                applyFiltersAndSort();
            }, 300);
        });
    }

    const sortOrder = document.getElementById('sortOrder');
    if (sortOrder) {
        sortOrder.addEventListener('change', (e) => {
            currentSortOrder = e.target.value;
            applyFiltersAndSort();
        });
    }

    const filterCategory = document.getElementById('filterCategory');
    if (filterCategory) {
        filterCategory.addEventListener('change', (e) => {
            currentCategory = e.target.value;
            applyFiltersAndSort();
        });
    }

    const filterStatus = document.getElementById('filterStatus');
    if (filterStatus) {
        filterStatus.addEventListener('change', (e) => {
            currentStatus = e.target.value;
            applyFiltersAndSort();
        });
    }

    const resetFiltersBtn = document.getElementById('resetFilters');
    if (resetFiltersBtn) {
        resetFiltersBtn.addEventListener('click', () => {
            if (searchInput) searchInput.value = '';
            if (sortOrder) sortOrder.value = 'desc';
            if (filterCategory) filterCategory.value = '';
            if (filterStatus) filterStatus.value = '';

            currentSearchQuery = '';
            currentSortOrder = 'desc';
            currentCategory = '';
            currentStatus = '';
            applyFiltersAndSort();
        });
    }
});

document.getElementById('addAnotherMerchantBtn').addEventListener('click', () => {
    const container = document.getElementById('merchantBlocksContainer');
    const firstBlock = container.querySelector('.merchant-block');
    const newBlock = firstBlock.cloneNode(true);

    // Clear inputs (keep default commission rate)
    const inputs = newBlock.querySelectorAll('input, select');
    inputs.forEach(input => {
        if (input.name !== 'commissionRate' && input.name !== 'businessType') {
            input.value = '';
        }
    });

    const blockCount = container.querySelectorAll('.merchant-block').length + 1;
    newBlock.querySelector('.merchant-title').textContent = `Merchant #${blockCount}`;

    const removeBtn = newBlock.querySelector('.remove-merchant-btn');
    removeBtn.classList.remove('hidden');
    removeBtn.addEventListener('click', function () {
        newBlock.remove();
        updateMerchantTitles();
    });

    container.appendChild(newBlock);
});

function updateMerchantTitles() {
    const blocks = document.querySelectorAll('.merchant-block');
    blocks.forEach((block, index) => {
        block.querySelector('.merchant-title').textContent = `Merchant #${index + 1}`;
    });
}

document.getElementById('onboardForm').addEventListener('submit', function (e) {
    e.preventDefault();

    const blocks = document.querySelectorAll('.merchant-block');
    const merchantsData = [];

    blocks.forEach(block => {
        const inputs = block.querySelectorAll('input, select');
        const merchantObj = {};
        inputs.forEach(input => {
            if (input.name) merchantObj[input.name] = input.value;
        });
        merchantsData.push(merchantObj);
    });

    const alertBox = document.getElementById('formAlert');
    alertBox.classList.add('hidden');

    const adminUsername = window.location.pathname.split('/')[2];
    fetch(`/v1/admin/${adminUsername}/merchants/add`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(merchantsData)
    })
        .then(response => response.json())
        .then(result => {
            alertBox.classList.remove('hidden', 'bg-red-50', 'text-red-600', 'bg-green-50', 'text-green-600');
            if (result.success) {
                alertBox.classList.add('bg-green-50', 'text-green-600');
                alertBox.textContent = result.message || "Merchants onboarded successfully!";
                setTimeout(() => window.location.reload(), 1500);
            } else {
                alertBox.classList.add('bg-red-50', 'text-red-600');
                alertBox.textContent = result.message || "An error occurred.";
            }
        })
        .catch(error => {
            alertBox.classList.remove('hidden', 'bg-green-50', 'text-green-600');
            alertBox.classList.add('bg-red-50', 'text-red-600');
            alertBox.textContent = "Network error. Please try again.";
            console.error('Error onboarding merchant:', error);
        });
});
