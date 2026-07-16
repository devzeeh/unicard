document.addEventListener("DOMContentLoaded", () => {
    if (!window.CURRENT_USERNAME) {
        console.error("CURRENT_USERNAME is not defined. Cannot fetch transactions.");
        return;
    }

    const formatCurrency = (amount) => {
        return new Intl.NumberFormat('en-PH', {
            style: 'currency',
            currency: 'PHP'
        }).format(amount);
    };

    const formatCurrencyStats = (amount) => {
        let absAmt = Math.abs(amount);
        const getFormattedNum = (n) => {
            if (n >= 100) return Math.floor(n * 10) / 10;
            return Math.floor(n * 100) / 100;
        };

        if (absAmt >= 1000000000) {
            let num = absAmt / 1000000000;
            return (amount < 0 ? '-₱' : '₱') + getFormattedNum(num) + 'B';
        } else if (absAmt >= 1000000) {
            let num = absAmt / 1000000;
            return (amount < 0 ? '-₱' : '₱') + getFormattedNum(num) + 'M';
        } else if (absAmt >= 10000) {
            let num = absAmt / 1000;
            return (amount < 0 ? '-₱' : '₱') + getFormattedNum(num) + 'k';
        }
        return new Intl.NumberFormat('en-PH', { style: 'currency', currency: 'PHP' }).format(amount);
    };

    const validTransactionTypes = new Set(['payment','refund','reversal','topup','withdrawal']);
    const shouldShowAmount = (type) => typeof type === 'string' && validTransactionTypes.has(type.toLowerCase());

    const fetchStats = async () => {
        try {
            const response = await fetch(`/v1/merchant/${window.CURRENT_USERNAME}/incomes`);
            const json = await response.json();
            if (json.success && json.data && json.data.stats) {
                const stats = json.data.stats;
                const availableBalanceEl = document.getElementById('availableBalance');
                if (availableBalanceEl) availableBalanceEl.textContent = formatCurrencyStats(stats.available_balance);
                window.MERCHANT_AVAILABLE_BALANCE = parseFloat(stats.available_balance) || 0;
                
                const monthIncomesEl = document.getElementById('monthIncomes');
                if (monthIncomesEl) monthIncomesEl.textContent = formatCurrencyStats(stats.monthly_net_income);
                
                const grossRevenueEl = document.getElementById('grossRevenue');
                if (grossRevenueEl) grossRevenueEl.textContent = formatCurrencyStats(stats.gross_revenue);
                
                const platformFeeEl = document.getElementById('platformFee');
                if (platformFeeEl) platformFeeEl.textContent = formatCurrencyStats(stats.platform_fee);
            }
        } catch (error) {
            console.error("Error fetching stats:", error);
        }
    };
    fetchStats();

    const fetchTransactions = async () => {
        try {
            const searchVal = document.getElementById('txSearch')?.value || '';
            const typeVal = document.getElementById('txType')?.value || 'all';
            const sortVal = document.getElementById('txSort')?.value || 'desc';
            
            const params = new URLSearchParams();
            if (searchVal) params.append('search', searchVal);
            if (typeVal && typeVal !== 'all') params.append('type', typeVal);
            if (sortVal) params.append('sort', sortVal);

            const queryString = params.toString() ? `?${params.toString()}` : '';
            const response = await fetch(`/v1/merchant/${window.CURRENT_USERNAME}/transactions${queryString}`);
            const json = await response.json();

            if (json.success && json.data) {
                const tbody = document.getElementById('transactionsTableBody');
                if (!tbody) return;

                if (json.data.length === 0) {
                    tbody.innerHTML = `
                        <tr>
                            <td colspan="5" class="px-6 py-12 text-center text-gray-500">
                                <div class="flex flex-col items-center justify-center space-y-3">
                                    <div class="p-4 bg-gray-50 rounded-full">
                                        <svg class="w-10 h-10 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>
                                    </div>
                                    <span class="font-medium text-lg">No transactions found</span>
                                    <p class="text-sm text-gray-400">Your recent transaction history will appear here.</p>
                                </div>
                            </td>
                        </tr>
                    `;
                    return;
                }

                tbody.innerHTML = '';
                json.data.forEach(tx => {
                    const dateObj = new Date(tx.created_at);
                    const formattedDate = dateObj.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
                    const timeStr = dateObj.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true });

                    const showAmount = shouldShowAmount(tx.transaction_type);
                    const isPayment = tx.transaction_type.toLowerCase() === 'payment';
                    const amountColor = showAmount ? (isPayment ? 'text-green-600' : 'text-red-600') : '';
                    const sign = isPayment ? '+' : '-';
                    const amountStr = showAmount ? `${sign}${formatCurrency(Math.abs(tx.amount))}` : '';

                    let statusBadgeClass = 'bg-green-100 text-green-800';
                    if (tx.status.toLowerCase() === 'failed' || tx.status.toLowerCase() === 'declined') {
                        statusBadgeClass = 'bg-red-100 text-red-800';
                    } else if (tx.status.toLowerCase() === 'pending') {
                        statusBadgeClass = 'bg-yellow-100 text-yellow-800';
                    } else if (tx.status.toLowerCase() !== 'completed') {
                        statusBadgeClass = 'bg-gray-100 text-gray-800';
                    }

                    const tr = document.createElement('tr');
                    tr.className = 'hover:bg-blue-50 transition-colors duration-150 cursor-pointer';
                    tr.innerHTML = `
                        <td class="px-6 py-4 whitespace-nowrap">
                            <div class="text-sm font-semibold text-gray-900">${formattedDate}</div>
                            <div class="text-xs text-gray-500 mt-1">${timeStr}</div>
                        </td>
                        <td class="px-6 py-4 max-w-0 overflow-hidden">
                            ${tx.description ? `<div class="text-sm font-medium text-gray-900 truncate">${tx.description}</div>` : ''}
                            <div class="text-xs text-gray-500 truncate ${tx.description ? 'mt-1' : ''}">ID: ${tx.transaction_id}</div>
                        </td>
                        <td class="px-6 py-4 whitespace-nowrap">
                            <div class="text-sm text-gray-600 capitalize">${tx.transaction_type}</div>
                        </td>
                        <td class="px-6 py-4 whitespace-nowrap text-right">
                            <div class="text-sm font-semibold ${amountColor}">${amountStr}</div>
                            ${showAmount && tx.service_fee > 0 ? `<div class="text-[10px] text-gray-400 mt-1">Fee: ${formatCurrency(Math.abs(tx.service_fee))}</div>` : ''}
                        </td>
                        <td class="px-6 py-4 whitespace-nowrap text-right">
                            <span class="px-3 py-1 inline-flex text-xs leading-5 font-semibold rounded-full capitalize ${statusBadgeClass}">
                                ${tx.status}
                            </span>
                        </td>
                    `;
                    tr.onclick = function () {
                        openTxnModal(tx);
                    };
                    tbody.appendChild(tr);
                });
            } else {
                console.error("Failed to fetch transactions:", json.message);
            }
        } catch (error) {
            console.error("Error fetching transactions:", error);
        }
    };

    fetchTransactions();

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

        const dateObj = new Date(tx.created_at || tx.date);
        const formattedDate = dateObj.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
        const timeStr = dateObj.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true });

        document.getElementById("modalTxnId").textContent = tx.transaction_id || 'N/A';
        document.getElementById("modalTxnDate").textContent = `${formattedDate} at ${timeStr}`;
        document.getElementById("modalTxnType").textContent = tx.transaction_type ? tx.transaction_type.charAt(0).toUpperCase() + tx.transaction_type.slice(1) : "N/A";

        document.getElementById("modalTxnCardNumber").textContent = tx.card_number || 'N/A';
        const terminalIdEl = document.getElementById("modalTxnTerminalId");
        const terminalLabelEl = document.getElementById("modalTxnTerminalLabel");
        if (tx.terminal_id && tx.terminal_id !== 'N/A' && tx.terminal_id.trim() !== '') {
            terminalIdEl.textContent = tx.terminal_id;
            if (terminalLabelEl) terminalLabelEl.classList.remove("hidden");
        } else {
            terminalIdEl.textContent = '';
            if (terminalLabelEl) terminalLabelEl.classList.add("hidden");
        }

        document.getElementById("modalTxnStatus").textContent = tx.status || 'completed';
        document.getElementById("modalTxnFee").textContent = formatCurrency(Number(tx.service_fee || 0));

        document.getElementById("modalTxnDesc").textContent = tx.description || 'N/A';

        const showAmount = shouldShowAmount(tx.transaction_type);
        const isPayment = (tx.transaction_type || '').toLowerCase() === "payment";
        const sign = isPayment ? "+" : "-";
        const colorClass = isPayment ? "text-green-600" : "text-red-600";
        
        const grossAmt = Number(tx.amount || 0);
        const grossEl = document.getElementById("modalTxnGross");
        const feeEl = document.getElementById("modalTxnFee");
        const netAmtEl = document.getElementById("modalTxnNet");
        const grossRow = grossEl.closest('.flex');
        const feeRow = feeEl.closest('.flex');
        const netRow = netAmtEl.closest('.flex');

        if (showAmount) {
            if (grossRow) grossRow.classList.remove('hidden');
            if (feeRow) feeRow.classList.remove('hidden');
            if (netRow) netRow.classList.remove('hidden');
            grossEl.textContent = formatCurrency(grossAmt);
            feeEl.textContent = formatCurrency(Number(tx.service_fee || 0));
            const netValue = Number(tx.net_merchant_payout || grossAmt);
            netAmtEl.textContent = `${sign}${formatCurrency(Math.abs(netValue))}`;
            netAmtEl.className = `font-bold text-lg ${colorClass}`;
        } else {
            if (grossRow) grossRow.classList.add('hidden');
            if (feeRow) feeRow.classList.add('hidden');
            if (netRow) netRow.classList.add('hidden');
            grossEl.textContent = '';
            feeEl.textContent = '';
            netAmtEl.textContent = '';
            netAmtEl.className = 'font-bold text-lg';
        }

        const isSystemEvent = grossAmt === 0 && Number(tx.service_fee || 0) === 0;
        
        if (isSystemEvent) {
            let sysType = "System Notification";
            if (tx.transaction_id && tx.transaction_id.toLowerCase().startsWith("welcome")) {
                sysType = "Account Approval";
            } else if (tx.description && tx.description.toLowerCase().includes("username")) {
                sysType = "Profile Update";
            } else if (tx.description && tx.description.toLowerCase().includes("settlement")) {
                sysType = "Bank Update";
            } else if (tx.transaction_type) {
                sysType = tx.transaction_type.charAt(0).toUpperCase() + tx.transaction_type.slice(1);
            }

            document.getElementById("modalTxnType").textContent = sysType;
            document.getElementById("modalTxnCardNumber").closest('.flex').classList.add("hidden");
            document.getElementById("modalTxnStatus").closest('.flex').classList.add("hidden");
            document.getElementById("modalTxnGross").closest('.flex').classList.add("hidden");
            document.getElementById("modalTxnFee").closest('.flex').classList.add("hidden");
            document.getElementById("modalTxnNet").closest('.flex').classList.add("hidden");
        } else {
            document.getElementById("modalTxnType").textContent = tx.transaction_type ? tx.transaction_type.charAt(0).toUpperCase() + tx.transaction_type.slice(1) : "N/A";
            document.getElementById("modalTxnCardNumber").closest('.flex').classList.remove("hidden");
            document.getElementById("modalTxnStatus").closest('.flex').classList.remove("hidden");
            document.getElementById("modalTxnGross").closest('.flex').classList.remove("hidden");
            document.getElementById("modalTxnFee").closest('.flex').classList.remove("hidden");
            document.getElementById("modalTxnNet").closest('.flex').classList.remove("hidden");
        }

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

    // --- Withdraw Modal Logic ---
    const withdrawModal = document.getElementById('withdrawModal');
    const withdrawModalContent = document.getElementById('withdrawModalContent');
    const closeWithdrawModalBtn = document.getElementById('closeWithdrawModalBtn');
    const cancelWithdrawBtn = document.getElementById('cancelWithdrawBtn');
    const submitWithdrawBtn = document.getElementById('submitWithdrawBtn');
    const withdrawMaxBtn = document.getElementById('withdrawMaxBtn');
    const withdrawAmountInput = document.getElementById('withdrawAmount');
    const withdrawErrorMsg = document.getElementById('withdrawErrorMsg');
    const withdrawModalBalance = document.getElementById('withdrawModalBalance');

    window.openWithdrawModal = function() {
        const bal = window.MERCHANT_AVAILABLE_BALANCE || 0;
        withdrawModalBalance.textContent = formatCurrency(bal);
        withdrawAmountInput.value = '';
        withdrawErrorMsg.classList.add('hidden');
        
        withdrawModal.classList.remove('hidden');
        setTimeout(() => {
            withdrawModal.classList.add('opacity-100');
            withdrawModalContent.classList.add('scale-100', 'opacity-100');
            withdrawModalContent.classList.remove('scale-95', 'opacity-0');
        }, 10);
    };

    window.closeWithdrawModal = function() {
        withdrawModalContent.classList.add('scale-95', 'opacity-0');
        withdrawModalContent.classList.remove('scale-100', 'opacity-100');
        withdrawModal.classList.remove('opacity-100');
        setTimeout(() => {
            withdrawModal.classList.add('hidden');
        }, 300);
    };

    if (closeWithdrawModalBtn) closeWithdrawModalBtn.addEventListener('click', window.closeWithdrawModal);
    if (cancelWithdrawBtn) cancelWithdrawBtn.addEventListener('click', window.closeWithdrawModal);
    
    const openBtn = document.getElementById('openWithdrawModalBtn');
    if (openBtn) openBtn.addEventListener('click', window.openWithdrawModal);

    const formatInputNumber = (val) => {
        let str = String(val).replace(/[^0-9.]/g, '');
        const parts = str.split('.');
        if (parts.length > 2) {
            str = parts[0] + '.' + parts.slice(1).join('');
        }
        const beforeDot = str.split('.')[0];
        const afterDot = str.split('.')[1];
        let formatted = beforeDot.replace(/\B(?=(\d{3})+(?!\d))/g, ",");
        if (afterDot !== undefined) {
            formatted += '.' + afterDot;
        }
        return formatted;
    };

    if (withdrawAmountInput) {
        withdrawAmountInput.addEventListener('input', (e) => {
            e.target.value = formatInputNumber(e.target.value);
        });
    }

    if (withdrawMaxBtn) {
        withdrawMaxBtn.addEventListener('click', () => {
            withdrawAmountInput.value = formatInputNumber(window.MERCHANT_AVAILABLE_BALANCE || 0);
        });
    }

    if (submitWithdrawBtn) {
        submitWithdrawBtn.addEventListener('click', async () => {
            const amount = parseFloat(withdrawAmountInput.value.replace(/,/g, ''));
            if (isNaN(amount) || amount <= 0) {
                withdrawErrorMsg.textContent = "Please enter a valid amount to withdraw.";
                withdrawErrorMsg.classList.remove('hidden');
                return;
            }
            if (amount > (window.MERCHANT_AVAILABLE_BALANCE || 0)) {
                withdrawErrorMsg.textContent = "Amount exceeds available balance.";
                withdrawErrorMsg.classList.remove('hidden');
                return;
            }

            submitWithdrawBtn.disabled = true;
            submitWithdrawBtn.innerHTML = '<svg class="animate-spin h-5 w-5 mr-2 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg> Processing...';

            try {
                const res = await fetch(`/v1/merchant/${window.CURRENT_USERNAME}/withdraw`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ amount: amount })
                });
                const json = await res.json();
                
                if (json.success) {
                    window.closeWithdrawModal();
                    alert("Withdrawal successful! Note: it may take a few hours to reflect in your bank account.");
                    fetchStats(); 
                    fetchTransactions();
                } else {
                    withdrawErrorMsg.textContent = json.message || "Failed to process withdrawal.";
                    withdrawErrorMsg.classList.remove('hidden');
                }
            } catch (err) {
                console.error(err);
                withdrawErrorMsg.textContent = "Network error. Please try again.";
                withdrawErrorMsg.classList.remove('hidden');
            } finally {
                submitWithdrawBtn.disabled = false;
                submitWithdrawBtn.innerHTML = '<span>Confirm Withdrawal</span>';
            }
        });
    }

    // --- Filter Event Listeners ---
    const debounce = (func, delay) => {
        let timeoutId;
        return (...args) => {
            clearTimeout(timeoutId);
            timeoutId = setTimeout(() => func.apply(null, args), delay);
        };
    };

    const txSearch = document.getElementById('txSearch');
    const txType = document.getElementById('txType');
    const txSort = document.getElementById('txSort');

    if (txSearch) txSearch.addEventListener('input', debounce(fetchTransactions, 300));
    if (txType) txType.addEventListener('change', fetchTransactions);
    if (txSort) txSort.addEventListener('change', fetchTransactions);
});
