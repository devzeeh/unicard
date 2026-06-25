document.addEventListener("DOMContentLoaded", () => {
    if (!window.CURRENT_USERNAME) {
        console.error("CURRENT_USERNAME is not defined. Cannot fetch dashboard data.");
        return;
    }

    const formatCurrency = (amount) => {
        let absAmt = Math.abs(amount);
        if (absAmt >= 1000000) {
            let num = absAmt / 1000000;
            let formatted = num.toFixed(1).replace(/\.0$/, '');
            return (amount < 0 ? '-₱' : '₱') + formatted + 'M';
        } else if (absAmt >= 10000) {
            let num = absAmt / 1000;
            let formatted = num.toFixed(1).replace(/\.0$/, '');
            return (amount < 0 ? '-₱' : '₱') + formatted + 'k';
        }
        return new Intl.NumberFormat('en-PH', { style: 'currency', currency: 'PHP' }).format(amount);
    };

    const fetchDashboardData = async () => {
        try {
            const response = await fetch(`/v1/merchant/${window.CURRENT_USERNAME}/dashboard`);
            const json = await response.json();

            if (json.success && json.data) {
                const data = json.data;

                window.MERCHANT_AVAILABLE_BALANCE = data.available_balance || 0;

                const destBankNameEl = document.getElementById("destinationBankName");
                if (destBankNameEl) destBankNameEl.textContent = data.settlement_bank || "Not Configured";

                const destAccNumEl = document.getElementById("destinationAccountNum");
                if (destAccNumEl) destAccNumEl.textContent = data.settlement_account || "N/A";

                // Update Overview Cards
                const availableBalanceEl = document.getElementById("availableBalance");
                if (availableBalanceEl) {
                    availableBalanceEl.textContent = formatCurrency(data.available_balance);
                }

                const grossRevenueEl = document.getElementById('grossRevenue');
                if (grossRevenueEl) {
                    grossRevenueEl.textContent = formatCurrency(data.gross_revenue);
                }

                const monthlyNetIncomeEl = document.getElementById('monthlyNetIncome');
                if (monthlyNetIncomeEl) {
                    monthlyNetIncomeEl.textContent = formatCurrency(data.monthly_net_income);
                }

                const totalRefundsEl = document.getElementById('totalRefunds');
                if (totalRefundsEl) {
                    totalRefundsEl.textContent = formatCurrency(data.total_refunds);
                }

                const totalPlatformFeeEl = document.getElementById('totalPlatformFee');
                if (totalPlatformFeeEl) {
                    totalPlatformFeeEl.textContent = formatCurrency(data.total_service_fee);
                }

                const totalTransactionsEl = document.getElementById('totalTransactions');
                if (totalTransactionsEl) {
                    totalTransactionsEl.textContent = data.total_transactions;
                }

                const accountStatusEl = document.getElementById('accountStatus');
                if (accountStatusEl) {
                    accountStatusEl.textContent = data.account_status;
                    if (data.account_status.toLowerCase() === 'active' || data.account_status.toLowerCase() === 'approved') {
                        accountStatusEl.className = "text-3xl font-bold text-emerald-600 tracking-tight capitalize h-9 flex items-center";
                    } else {
                        accountStatusEl.className = "text-3xl font-bold text-amber-600 tracking-tight capitalize h-9 flex items-center";
                    }
                }

                // Update Transactions Table
                const tbody = document.getElementById('transactionsTableBody');
                if (tbody) {
                    if (!data.recent_transactions || data.recent_transactions.length === 0) {
                        tbody.innerHTML = `
                            <tr>
                                <td colspan="4" class="px-6 py-8 text-center text-gray-500">
                                    <div class="flex flex-col items-center justify-center space-y-2">
                                        <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4M8 16l-4-4 4-4"/></svg>
                                        <span class="font-medium">No recent transactions</span>
                                    </div>
                                </td>
                            </tr>
                        `;
                        return;
                    }

                    tbody.innerHTML = '';
                    data.recent_transactions.forEach(tx => {
                        const dateObj = new Date(tx.created_at);
                        const formattedDate = dateObj.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
                        const timeStr = dateObj.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true });

                        const isPayment = tx.transaction_type.toLowerCase() === 'payment';
                        const amountColor = isPayment ? 'text-green-600' : 'text-red-600';
                        const sign = isPayment ? '+' : '-';
                        const amount = Number(tx.amount).toFixed(2);
                        const displayType = tx.transaction_type.charAt(0).toUpperCase() + tx.transaction_type.slice(1);

                        const tr = document.createElement('tr');
                        tr.className = 'hover:bg-slate-50 transition-colors cursor-pointer border-b border-slate-100';
                        tr.innerHTML = `
                            <td class="px-6 py-4 whitespace-nowrap">
                                <div class="font-medium text-gray-900">${formattedDate}</div>
                                <div class="text-xs text-gray-500 mt-0.5">${timeStr}</div>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap">
                                ${tx.description ? `<div class="font-medium text-gray-900">${tx.description}</div>` : ''}
                                <div class="text-xs text-gray-500 ${tx.description ? 'mt-0.5' : ''}">ID: ${tx.transaction_id}</div>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                ${displayType}
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm ${amountColor} text-right font-medium">
                                ${sign}₱${amount}
                            </td>
                        `;
                        tr.onclick = function () {
                            openTxnModal(tx);
                        };

                        tbody.appendChild(tr);
                    });
                }
            } else {
                console.error("Failed to fetch dashboard data:", json.message);
            }
        } catch (error) {
            console.error("Error fetching dashboard data:", error);
        }
    };

    fetchDashboardData();

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
        document.getElementById("modalTxnFee").textContent = `₱${Number(tx.service_fee || 0).toFixed(2)}`;

        document.getElementById("modalTxnDesc").textContent = tx.description || 'N/A';

        const isPayment = (tx.transaction_type || '').toLowerCase() === "payment";
        const sign = isPayment ? "+" : "-";
        const colorClass = isPayment ? "text-green-600" : "text-red-600";
        
        const grossAmt = Number(tx.amount || 0);
        document.getElementById("modalTxnGross").textContent = `₱${grossAmt.toFixed(2)}`;

        const netAmtEl = document.getElementById("modalTxnNet");
        const netValue = Number(tx.net_merchant_payout || grossAmt);
        netAmtEl.textContent = `${sign}₱${netValue.toFixed(2)}`;
        netAmtEl.className = `font-bold text-lg ${colorClass}`;

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
        if (withdrawModalBalance) withdrawModalBalance.textContent = formatCurrency(bal);
        if (withdrawAmountInput) withdrawAmountInput.value = '';
        if (withdrawErrorMsg) withdrawErrorMsg.classList.add('hidden');
        
        if (withdrawModal) {
            withdrawModal.classList.remove('hidden');
            setTimeout(() => {
                withdrawModal.classList.add('opacity-100');
                withdrawModalContent.classList.add('scale-100', 'opacity-100');
                withdrawModalContent.classList.remove('scale-95', 'opacity-0');
            }, 10);
        }
    };

    window.closeWithdrawModal = function() {
        if (!withdrawModal) return;
        withdrawModalContent.classList.add('scale-95', 'opacity-0');
        withdrawModalContent.classList.remove('scale-100', 'opacity-100');
        withdrawModal.classList.remove('opacity-100');
        setTimeout(() => {
            withdrawModal.classList.add('hidden');
        }, 300);
    };

    if (closeWithdrawModalBtn) closeWithdrawModalBtn.addEventListener('click', closeWithdrawModal);
    if (cancelWithdrawBtn) cancelWithdrawBtn.addEventListener('click', closeWithdrawModal);

    const openBtn = document.getElementById('openWithdrawModalBtn');
    if (openBtn) openBtn.addEventListener('click', window.openWithdrawModal);

    if (withdrawMaxBtn) {
        withdrawMaxBtn.addEventListener('click', () => {
            withdrawAmountInput.value = window.MERCHANT_AVAILABLE_BALANCE || 0;
        });
    }

    if (submitWithdrawBtn) {
        submitWithdrawBtn.addEventListener('click', async () => {
            const amount = parseFloat(withdrawAmountInput.value);
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
                    closeWithdrawModal();
                    alert("Withdrawal successful! Note: it may take a few hours to reflect in your bank account.");
                    location.reload();
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
});
