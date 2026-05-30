let originalMerchants = []; // store all fetched merchants
let allMerchants = []; // store filtered merchants
let currentPage = 1; // current page
const itemsPerPage = 10;    // items per page
let currentSearchQuery = '';
let currentSortOrder = 'desc';

function fetchMerchants() {
    fetch('/v1/admin/merchants-data')
        .then(response => response.json())
        .then(result => {
            if (result.success && result.data && result.data.merchants) {
                originalMerchants = result.data.merchants;
                applyFiltersAndSort();
            }
        })
        .catch(error => console.error('Error fetching merchants:', error));
}

function applyFiltersAndSort() {
    let filtered = originalMerchants;

    if (currentSearchQuery) {
        const queryTerms = currentSearchQuery.toLowerCase().trim().split(/\s+/);
        filtered = filtered.filter(m => {
            const searchableText = `${m.business_name || ''} ${m.owner_name || ''} ${m.merchant_id || ''}`.toLowerCase();
            return queryTerms.every(term => searchableText.includes(term));
        });
    }

    filtered.sort((a, b) => {
        const dateA = new Date(a.created_at || 0);
        const dateB = new Date(b.created_at || 0);
        return currentSortOrder === 'desc' ? dateB - dateA : dateA - dateB;
    });

    allMerchants = filtered;
    currentPage = 1;
    renderTable();
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

    const startIndex = (currentPage - 1) * itemsPerPage;
    const endIndex = Math.min(startIndex + itemsPerPage, allMerchants.length);
    const pageData = allMerchants.slice(startIndex, endIndex);

    const queryTerms = currentSearchQuery ? currentSearchQuery.toLowerCase().trim().split(/\s+/) : [];

    pageData.forEach(merchant => {
        const tr = document.createElement('tr');
        
        const highlightedBusinessName = highlightText(merchant.business_name, queryTerms);
        const highlightedMerchantId = highlightText(merchant.merchant_id, queryTerms);
        const highlightedOwnerName = highlightText(merchant.owner_name, queryTerms);

        tr.innerHTML = `
            <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                    <div class="flex-shrink-0 h-10 w-10 bg-blue-100 rounded-lg flex items-center justify-center text-blue-600">
                        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V10l-7-5-7 5v11m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path></svg>
                    </div>
                    <div class="ml-4">
                        <div class="text-sm font-medium text-gray-900">${highlightedBusinessName}</div>
                        <div class="text-xs text-gray-500">ID: ${highlightedMerchantId}</div>
                    </div>
                </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600 capitalize">${merchant.business_type.replace(/_/g, ' ')}</td>
            <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm text-gray-900">${highlightedOwnerName}</div>
                <div class="text-xs text-gray-500">${merchant.business_email}</div>
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

    document.getElementById('pageStart').textContent = allMerchants.length > 0 ? startIndex + 1 : 0;
    document.getElementById('pageEnd').textContent = endIndex;
    document.getElementById('totalItems').textContent = allMerchants.length;

    renderPagination();
}

function renderPagination() {
    const paginationControls = document.getElementById('paginationControls');
    paginationControls.innerHTML = '';

    const totalPages = Math.ceil(allMerchants.length / itemsPerPage);

    if (totalPages <= 1) return;

    const prevBtn = document.createElement('button');
    prevBtn.className = `px-3 py-1 rounded-md text-sm font-medium ${currentPage === 1 ? 'text-gray-400 cursor-not-allowed' : 'text-blue-600 hover:bg-blue-50'}`;
    prevBtn.textContent = 'Previous';
    prevBtn.disabled = currentPage === 1;
    prevBtn.onclick = () => { if (currentPage > 1) { currentPage--; renderTable(); } };
    paginationControls.appendChild(prevBtn);

    for (let i = 1; i <= totalPages; i++) {
        const btn = document.createElement('button');
        btn.className = `px-3 py-1 rounded-md text-sm font-medium ${currentPage === i ? 'bg-blue-600 text-white' : 'text-gray-700 hover:bg-gray-50'}`;
        btn.textContent = i;
        btn.onclick = () => { currentPage = i; renderTable(); };
        paginationControls.appendChild(btn);
    }

    const nextBtn = document.createElement('button');
    nextBtn.className = `px-3 py-1 rounded-md text-sm font-medium ${currentPage === totalPages ? 'text-gray-400 cursor-not-allowed' : 'text-blue-600 hover:bg-blue-50'}`;
    nextBtn.textContent = 'Next';
    nextBtn.disabled = currentPage === totalPages;
    nextBtn.onclick = () => { if (currentPage < totalPages) { currentPage++; renderTable(); } };
    paginationControls.appendChild(nextBtn);
}

document.addEventListener('DOMContentLoaded', () => {
    fetchMerchants();

    const searchInput = document.getElementById('searchInput');
    if (searchInput) {
        searchInput.addEventListener('input', (e) => {
            currentSearchQuery = e.target.value;
            applyFiltersAndSort();
        });
    }

    const sortOrder = document.getElementById('sortOrder');
    if (sortOrder) {
        sortOrder.addEventListener('change', (e) => {
            currentSortOrder = e.target.value;
            applyFiltersAndSort();
        });
    }

    const resetFiltersBtn = document.getElementById('resetFilters');
    if (resetFiltersBtn) {
        resetFiltersBtn.addEventListener('click', () => {
            if (searchInput) searchInput.value = '';
            if (sortOrder) sortOrder.value = 'desc';
            currentSearchQuery = '';
            currentSortOrder = 'desc';
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
    removeBtn.addEventListener('click', function() {
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

    fetch('/v1/admin/merchants/add', {
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
