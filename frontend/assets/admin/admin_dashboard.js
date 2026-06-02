document.addEventListener("DOMContentLoaded", function () {
    const adminUsername = window.location.pathname.split('/')[2];
    fetch(`/v1/admin/${adminUsername}/dashboard-data`)
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                document.getElementById('grossRevenue').textContent = '₱' + data.data.grossRevenue.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
                document.getElementById('netRevenue').textContent = '₱' + data.data.netRevenue.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
                document.getElementById('totalUsers').textContent = data.data.totalUsers.toLocaleString();
                document.getElementById('totalCards').textContent = data.data.totalCards.toLocaleString();
                document.getElementById('activeMerchants').textContent = data.data.activeMerchants.toLocaleString();
                document.getElementById('activeTerminals').textContent = data.data.activeTerminals.toLocaleString();

                // Populate Merchants Table
                const tbody = document.getElementById('merchantsTableBody');
                tbody.innerHTML = '';
                if (data.data.merchants && data.data.merchants.length > 0) {
                    data.data.merchants.forEach(m => {
                        const tr = document.createElement('tr');
                        tr.className = 'hover:bg-gray-50';

                        // Status badge styling
                        let statusColor = 'bg-gray-100 text-gray-800';
                        if (m.status === 'active') {
                            statusColor = 'bg-green-100 text-green-800';
                        } else if (m.status === 'pending_approval') {
                            statusColor = 'bg-yellow-100 text-yellow-800';
                        } else if (m.status === 'suspended') {
                            statusColor = 'bg-red-100 text-red-800';
                        }

                        tr.innerHTML = `
                            <td class="p-3 whitespace-nowrap">
                                <div class="flex items-center">
                                    <div class="flex-shrink-0 h-10 w-10 bg-blue-100 rounded-lg flex items-center justify-center text-blue-600">
                                        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V10l-7-5-7 5v11m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path></svg>
                                    </div>
                                    <div class="ml-4">
                                        <div class="text-sm font-medium text-gray-900">${m.business_name}</div>
                                        <div class="text-xs text-gray-500">ID: ${m.merchant_id}</div>
                                    </div>
                                </div>
                            </td>
                            <td class="p-3 capitalize text-sm text-gray-600">${m.business_type.replace(/_/g, ' ')}</td>
                            <td class="p-3 whitespace-nowrap">
                                <div class="text-sm text-gray-900">${m.owner_name}</div>
                                <div class="text-xs text-gray-500">${m.business_email}</div>
                            </td>
                            <td class="p-3 text-sm text-gray-600">${m.business_phone}</td>
                            <td class="p-3">
                                <span class="px-2 py-1 text-xs font-medium rounded-full ${statusColor} capitalize">
                                    ${m.status.replace(/_/g, ' ')}
                                </span>
                            </td>
                            <td class="p-3 text-gray-500 text-xs">${new Date(m.created_at).toLocaleDateString()}</td>
                        `;
                        tbody.appendChild(tr);
                    });
                } else {
                    tbody.innerHTML = `
                        <tr>
                            <td colspan="6" class="p-8 text-center">
                                <div class="flex flex-col items-center justify-center text-gray-500">
                                    <i class="fas fa-store-slash text-4xl mb-3 text-gray-300"></i>
                                    <p class="text-base font-medium text-gray-900">No merchants registered yet</p>
                                    <p class="text-sm text-gray-500 mt-1">When merchants are added, they will appear here.</p>
                                </div>
                            </td>
                        </tr>
                    `;
                }
            } else {
                console.error("Failed to load dashboard data:", data.message);
            }
        })
        .catch(error => console.error("Error fetching dashboard data:", error));
});