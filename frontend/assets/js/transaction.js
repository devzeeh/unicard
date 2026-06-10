document.addEventListener("DOMContentLoaded", function () {
    const transactionsBody = document.getElementById("transactions-table-body");

    function loadTransactions() {
        const pathSegments = window.location.pathname.split('/');
        let userId = null;
        if (pathSegments.length >= 2 && pathSegments[1] !== '') {
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

                transactionsBody.innerHTML = "";
                const txs = data.transactions;

                if (txs && txs.length > 0) {
                    txs.forEach(tx => {
                        const tr = document.createElement("tr");
                        const isPayment = tx.type === "Payment";
                        const colorClass = isPayment ? "text-red-600" : "text-green-600";
                        const sign = isPayment ? "-" : "+";
                        const amount = Number(tx.amount).toFixed(2);
                        let txDate = "N/A";
                        let txTime = "";
                        if (tx.date) {
                            const d = new Date(tx.date);
                            if (!isNaN(d.getTime())) {
                                txDate = d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
                                txTime = d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
                            } else {
                                txDate = tx.date;
                            }
                        }
                        const status = tx.status || "Completed";

                        let statusColor = "bg-green-100 text-green-800";
                        if (status.toLowerCase() === "pending") {
                            statusColor = "bg-yellow-100 text-yellow-800";
                        } else if (status.toLowerCase() === "failed") {
                            statusColor = "bg-red-100 text-red-800";
                        }

                        tr.innerHTML = `
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                <div>\${txDate}</div>
                                <div class="text-xs text-gray-400 mt-1">\${txTime}</div>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                                \${tx.description}
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                \${tx.type}
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm \${colorClass} text-right font-medium">
                                \${sign}₱\${amount}
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-right">
                                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full \${statusColor}">
                                    \${status}
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
