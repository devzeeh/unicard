document.addEventListener("DOMContentLoaded", function () {
    const transactionsBody = document.getElementById("transactions-table-body");

    let allTransactions = [];
    let currentPage = 1;
    const itemsPerPage = 10;

    // DOM Elements
    const searchTxn = document.getElementById("searchTxn");
    const filterType = document.getElementById("filterType");
    const filterStatus = document.getElementById("filterStatus");
    const sortOrder = document.getElementById("sortOrder");
    const pageStart = document.getElementById("pageStart");
    const pageEnd = document.getElementById("pageEnd");
    const totalItems = document.getElementById("totalItems");
    const prevPageBtn = document.getElementById("prevPageBtn");
    const nextPageBtn = document.getElementById("nextPageBtn");

    // Event Listeners for Filters
    if (searchTxn) searchTxn.addEventListener("input", () => { currentPage = 1; renderTransactions(); });
    if (filterType) filterType.addEventListener("change", () => { currentPage = 1; renderTransactions(); });
    if (filterStatus) filterStatus.addEventListener("change", () => { currentPage = 1; renderTransactions(); });
    if (sortOrder) sortOrder.addEventListener("change", () => { currentPage = 1; renderTransactions(); });
    if (prevPageBtn) prevPageBtn.addEventListener("click", () => { if (currentPage > 1) { currentPage--; renderTransactions(); } });
    if (nextPageBtn) nextPageBtn.addEventListener("click", () => { currentPage++; renderTransactions(); });

    function loadTransactions() {
        const pathSegments = window.location.pathname.split('/');
        let userId = null;
        if (pathSegments.length >= 3 && pathSegments[1] === 'u') {
            userId = pathSegments[2];
        } else if (pathSegments.length >= 2 && pathSegments[1] !== '') {
            userId = pathSegments[1];
        }

        let endpoint = "/v1/user/transactions";
        if (userId) {
            endpoint = "/v1/user/" + encodeURIComponent(userId) + "/transactions";
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
                if (!data || !data.success) {
                    showError();
                    return;
                }
                allTransactions = data.transactions || [];
                renderTransactions();
            })
            .catch(error => {
                console.error("Error loading transactions:", error);
                showError();
            });
    }

    function renderTransactions() {
        if (!transactionsBody) return;

        let filtered = [...allTransactions];

        // Apply Search
        const searchVal = searchTxn ? searchTxn.value.toLowerCase() : "";
        if (searchVal) {
            filtered = filtered.filter(tx => 
                (tx.description && tx.description.toLowerCase().includes(searchVal)) ||
                (tx.transaction_id && tx.transaction_id.toLowerCase().includes(searchVal)) ||
                (tx.terminal_id && tx.terminal_id.toLowerCase().includes(searchVal))
            );
        }

        // Apply Type Filter
        const typeVal = filterType ? filterType.value.toLowerCase() : "";
        if (typeVal) {
            filtered = filtered.filter(tx => tx.type && tx.type.toLowerCase() === typeVal);
        }

        // Apply Status Filter
        const statusVal = filterStatus ? filterStatus.value.toLowerCase() : "";
        if (statusVal) {
            filtered = filtered.filter(tx => tx.status && tx.status.toLowerCase() === statusVal);
        }

        // Apply Sort
        const sortVal = sortOrder ? sortOrder.value : "desc";
        filtered.sort((a, b) => {
            const dateA = new Date(a.date + " " + a.time).getTime() || 0;
            const dateB = new Date(b.date + " " + b.time).getTime() || 0;
            return sortVal === "asc" ? dateA - dateB : dateB - dateA;
        });

        // Pagination setup
        const total = filtered.length;
        const totalPages = Math.ceil(total / itemsPerPage) || 1;
        if (currentPage > totalPages) currentPage = totalPages;
        
        const startIdx = (currentPage - 1) * itemsPerPage;
        const endIdx = Math.min(startIdx + itemsPerPage, total);
        const paginated = filtered.slice(startIdx, endIdx);

        // Update Pagination UI
        if (pageStart) pageStart.textContent = total === 0 ? 0 : startIdx + 1;
        if (pageEnd) pageEnd.textContent = endIdx;
        if (totalItems) totalItems.textContent = total;
        if (prevPageBtn) prevPageBtn.disabled = currentPage <= 1;
        if (nextPageBtn) nextPageBtn.disabled = currentPage >= totalPages;

        transactionsBody.innerHTML = "";

        if (paginated.length > 0) {
            paginated.forEach(tx => {
                const tr = document.createElement("tr");
                const isPayment = tx.type && tx.type.toLowerCase() === "payment";
                const colorClass = isPayment ? "text-red-600" : "text-green-600";
                const sign = isPayment ? "-" : "+";
                const amount = Number(tx.amount).toFixed(2);
                const displayType = tx.type ? tx.type.charAt(0).toUpperCase() + tx.type.slice(1) : "";
                
                let txDate = "N/A";
                let txTime = "";
                if (tx.date) {
                    const d = new Date(tx.date);
                    if (!isNaN(d.getTime())) {
                        txDate = d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
                        txTime = d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
                    } else {
                        txDate = tx.date;
                        txTime = tx.time || "";
                    }
                }
                const status = tx.status || "Completed";

                let statusColor = "bg-green-100 text-green-800";
                if (status.toLowerCase() === "pending") {
                    statusColor = "bg-yellow-100 text-yellow-800";
                } else if (status.toLowerCase() === "failed") {
                    statusColor = "bg-red-100 text-red-800";
                }

                tr.className = "hover:bg-slate-50 transition-colors cursor-pointer border-b border-slate-100";
                tr.onclick = function() {
                    openTxnModal(tx, txDate, txTime);
                };

                tr.innerHTML = `
                    <td class="px-6 py-4 whitespace-nowrap">
                        <div class="font-medium text-gray-900">${txDate}</div>
                        <div class="text-xs text-gray-500 mt-0.5">${txTime}</div>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <div class="font-medium text-gray-900">${tx.description}</div>
                        <div class="text-xs text-gray-500 mt-0.5">ID: ${tx.transaction_id || 'N/A'}</div>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        ${displayType}
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm ${colorClass} text-right font-medium">
                        ${sign}₱${amount}
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-right">
                        <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${statusColor}">
                            ${status}
                        </span>
                    </td>
                `;
                transactionsBody.appendChild(tr);
            });
        } else {
            transactionsBody.innerHTML = `
                <tr>
                    <td colspan="5" class="px-6 py-10 text-center text-sm text-gray-500">
                        No transactions found.
                    </td>
                </tr>
            `;
        }
    }

    function showError() {
        if (transactionsBody) {
            transactionsBody.innerHTML = `
                <tr>
                    <td colspan="5" class="px-6 py-10 text-center text-sm text-red-500">
                        Failed to load transactions.
                    </td>
                </tr>
            `;
        }
    }

    loadTransactions();

    // --- Modal Logic ---
    const txnModal = document.getElementById("txnModal");
    const txnModalContent = document.getElementById("txnModalContent");
    const closeTxnModalBtn = document.getElementById("closeTxnModalBtn");
    const closeTxnModalBottomBtn = document.getElementById("closeTxnModalBottomBtn");

    if (txnModal && closeTxnModalBtn) {
        closeTxnModalBtn.onclick = closeTxnModal;
        closeTxnModalBottomBtn.onclick = closeTxnModal;
        txnModal.onclick = function(e) {
            if (e.target === txnModal) closeTxnModal();
        };
    }

    function openTxnModal(tx, txDate, txTime) {
        if (!txnModal) return;
        
        document.getElementById("modalTxnId").textContent = tx.transaction_id || 'N/A';
        document.getElementById("modalTxnMerchant").textContent = tx.description || 'N/A';
        document.getElementById("modalTxnTerminal").textContent = tx.terminal_id || 'N/A';
        document.getElementById("modalTxnDate").textContent = `${txDate} at ${txTime}`;
        document.getElementById("modalTxnType").textContent = tx.type ? tx.type.charAt(0).toUpperCase() + tx.type.slice(1) : "N/A";
        
        const isPayment = tx.type && tx.type.toLowerCase() === "payment";
        const sign = isPayment ? "-" : "+";
        const colorClass = isPayment ? "text-red-600" : "text-green-600";
        const amtEl = document.getElementById("modalTxnAmount");
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
