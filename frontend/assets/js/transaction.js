document.addEventListener("DOMContentLoaded", function () {
    const transactionsBody = document.getElementById("transactions-table-body");

    function loadTransactions() {
        fetch("/v1/user/transactions")
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

                transactionsBody.innerHTML = "";
                const txs = data.transactions;

                if (txs && txs.length > 0) {
                    txs.forEach(tx => {
                        const tr = document.createElement("tr");
                        const isPayment = tx.type === "Payment";
                        const colorClass = isPayment ? "text-red-600" : "text-green-600";
                        const sign = isPayment ? "-" : "+";
                        const amount = Number(tx.amount).toFixed(2);
                        
                        // We do not have a running balance per transaction from the DB currently, 
                        // so we will just show a dash or N/A in the balance column for now.
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
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                                -
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
            })
            .catch(error => {
                console.error("Error loading transactions:", error);
                showError();
            });
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
});
