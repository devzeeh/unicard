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
                document.getElementById('totalMerchants').textContent = data.data.totalMerchants.toLocaleString();
                document.getElementById('pendingMerchants').textContent = data.data.pendingMerchants.toLocaleString();
                document.getElementById('suspendedMerchants').textContent = data.data.suspendedMerchants.toLocaleString();
                document.getElementById('rejectedMerchants').textContent = data.data.rejectedMerchants.toLocaleString();
                
                document.getElementById('totalTerminals').textContent = data.data.totalTerminals.toLocaleString();
                document.getElementById('activeTerminals').textContent = data.data.activeTerminals.toLocaleString();
                document.getElementById('inactiveTerminals').textContent = data.data.inactiveTerminals.toLocaleString();

                // Populate Merchants Table
                window.dashboardMerchants = data.data.merchants || [];
                const tbody = document.getElementById('merchantsTableBody');

                function renderDashboardMerchants(merchantsToRender) {
                    tbody.innerHTML = '';
                    if (merchantsToRender && merchantsToRender.length > 0) {
                        merchantsToRender.forEach(m => {
                            const tr = document.createElement('tr');
                            tr.className = 'hover:bg-blue-50 cursor-pointer transition duration-150';

                            tr.onclick = () => {
                                window.location.href = `/admin/${adminUsername}/merchants/${m.merchant_id}`;
                            };

                            // Status badge styling
                            let statusColor = 'bg-gray-100 text-gray-800';
                            if (m.status === 'active') {
                                statusColor = 'bg-green-100 text-green-800';
                            } else if (m.status === 'pending_approval' || m.status === 'pending approval') {
                                statusColor = 'bg-yellow-100 text-yellow-800';
                            } else if (m.status === 'suspended') {
                                statusColor = 'bg-orange-100 text-orange-800';
                            } else if (m.status === 'rejected') {
                                statusColor = 'bg-red-100 text-red-800';
                            } else if (m.status === 'approved') {
                                statusColor = 'bg-blue-100 text-blue-800';
                            }

                            tr.innerHTML = `
                                <td class="px-6 py-4 whitespace-nowrap max-w-[250px]" title="${m.business_name}">
                                    <div class="truncate">
                                        <div class="text-sm font-medium text-gray-900 truncate">${m.business_name}</div>
                                        <div class="text-xs text-gray-500 truncate">ID: ${m.merchant_id}</div>
                                    </div>
                                </td>
                                <td class="px-6 py-4 capitalize text-sm text-gray-600 max-w-[150px] truncate" title="${m.business_type}">${m.business_type.replace(/_/g, ' ')}</td>
                                <td class="px-6 py-4 whitespace-nowrap max-w-[250px]" title="${m.owner_name} / ${m.business_email}">
                                    <div class="text-sm text-gray-900 truncate">${m.owner_name}</div>
                                    <div class="text-xs text-gray-500 truncate">${m.business_email}</div>
                                </td>
                                <td class="px-6 py-4 text-sm text-gray-600">${m.business_phone}</td>
                                <td class="px-6 py-4">
                                    <span class="px-2 py-1 text-xs font-medium rounded-full ${statusColor} capitalize">
                                        ${m.status.replace(/_/g, ' ')}
                                    </span>
                                </td>
                                <td class="px-6 py-4 text-gray-500 text-xs">${new Date(m.created_at).toLocaleDateString()}</td>
                            `;
                            tbody.appendChild(tr);
                        });
                    } else {
                        tbody.innerHTML = `
                            <tr>
                                <td colspan="6" class="p-8 text-center">
                                    <div class="flex flex-col items-center justify-center text-gray-500">
                                        <i class="fas fa-store-slash text-4xl mb-3 text-gray-300"></i>
                                        <p class="text-base font-medium text-gray-900">No merchants found</p>
                                        <p class="text-sm text-gray-500 mt-1">Try a different search term or add merchants.</p>
                                    </div>
                                </td>
                            </tr>
                        `;
                    }
                }

                // Initial render
                renderDashboardMerchants(window.dashboardMerchants);
            } else {
                console.error("Failed to load dashboard data:", data.message);
            }
        })
        .catch(error => console.error("Error fetching dashboard data:", error));
});