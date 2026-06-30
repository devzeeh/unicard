let currentPage = 1;
let currentStatus = '';
let currentSearch = '';
let totalPages = 1;

const adminUsername = window.location.pathname.split('/')[2];

document.addEventListener('DOMContentLoaded', function() {
    loadTerminalRequests();
    setupEventListeners();
});

function setupEventListeners() {
    const searchInput = document.getElementById('searchInput');
    const statusFilter = document.getElementById('statusFilter');
    const prevBtn = document.getElementById('prevBtn');
    const nextBtn = document.getElementById('nextBtn');
    
    if (!searchInput) console.warn('searchInput element not found');
    if (!statusFilter) console.warn('statusFilter element not found');
    if (!prevBtn) console.warn('prevBtn element not found');
    if (!nextBtn) console.warn('nextBtn element not found');
    
    if (searchInput) {
        searchInput.addEventListener('input', debounce(function(e) {
            currentSearch = e.target.value;
            currentPage = 1;
            loadTerminalRequests();
        }, 300));
    }

    if (statusFilter) {
        statusFilter.addEventListener('change', function(e) {
            currentStatus = e.target.value;
            currentPage = 1;
            loadTerminalRequests();
        });
    }

    if (prevBtn) {
        prevBtn.addEventListener('click', function() {
            if (currentPage > 1) {
                currentPage--;
                loadTerminalRequests();
            }
        });
    }

    if (nextBtn) {
        nextBtn.addEventListener('click', function() {
            if (currentPage < totalPages) {
                currentPage++;
                loadTerminalRequests();
            }
        });
    }
}

function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

async function loadTerminalRequests() {
    try {
        const params = new URLSearchParams();
        params.append('page', currentPage);
        params.append('limit', 10);
        
        if (currentStatus) {
            params.append('status', currentStatus);
        }
        
        if (currentSearch) {
            params.append('search', currentSearch);
        }

        const url = `/v1/admin/${adminUsername}/terminal-requests-data?${params.toString()}`;
        console.log('Fetching terminal requests from:', url);

        const response = await fetch(url);
        console.log('Response status:', response.status);
        console.log('Response ok:', response.ok);

        if (!response.ok) {
            const errorText = await response.text();
            console.error('Response error:', errorText);
            throw new Error(`HTTP ${response.status}: ${errorText}`);
        }

        const data = await response.json();
        console.log('Response data:', data);

        if (!data.success) {
            showError(data.message || 'Failed to load terminal requests');
            return;
        }

        const requests = data.data || [];
        totalPages = data.total_pages || 1;

        renderTable(requests);
        updatePagination(data.total_items, data.current_page, totalPages);
        updatePendingCount(requests);

    } catch (error) {
        console.error('Error loading terminal requests:', error);
        showError('Failed to load terminal requests: ' + error.message);
    }
}

function renderTable(requests) {
    const tbody = document.getElementById('requestsTableBody');
    
    if (!requests || requests.length === 0) {
        tbody.innerHTML = `
            <tr class="border-b border-gray-200">
                <td colspan="7" class="px-6 py-12 text-center text-gray-500">
                    No terminal requests found
                </td>
            </tr>
        `;
        return;
    }

    tbody.innerHTML = requests.map(req => {
        const safeReq = {
            request_id: escapeHtml(req.request_id),
            merchant_id: escapeHtml(req.merchant_id),
            business_name: escapeHtml(req.business_name || req.merchant_id),
            owner_name: escapeHtml(req.owner_name || ''),
            terminal_sn: escapeHtml(req.terminal_sn || '-'),
            status: escapeHtml(req.status),
            requested_at: formatDate(req.requested_at),
            handled_by: req.handled_by ? escapeHtml(req.handled_by) : '-',
            notes: escapeHtml(req.notes || 'No reason provided')
        };
        
        // Data attribute for the row click
        const rowDataStr = encodeURIComponent(JSON.stringify(safeReq));

        return `
        <tr class="border-b border-gray-200 hover:bg-gray-50 transition cursor-pointer" onclick="openRequestDetailsModal('${rowDataStr}', event)">
            <td class="px-6 py-4">
                <span class="font-mono text-sm font-semibold text-blue-600">${safeReq.request_id}</span>
            </td>
            <td class="px-6 py-4">
                <div>
                    <p class="text-sm font-semibold text-gray-900">${safeReq.business_name}</p>
                    <p class="text-xs text-gray-500">${safeReq.owner_name}</p>
                </div>
            </td>
            <td class="px-6 py-4">
                <span class="font-mono text-sm text-gray-600">${safeReq.terminal_sn}</span>
            </td>
            <td class="px-6 py-4">
                <span class="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold ${getStatusBadgeClasses(req.status)}">
                    ${capitalizeFirst(req.status)}
                </span>
            </td>
            <td class="px-6 py-4 text-sm text-gray-600">
                ${safeReq.requested_at}
            </td>
            <td class="px-6 py-4 text-sm text-gray-600">
                ${safeReq.handled_by}
            </td>
            <td class="px-6 py-4 text-center">
                ${req.status === 'pending' ? `
                    <div class="flex gap-2 justify-center">
                        <button onclick="openApproveModal('${safeReq.request_id}', '${safeReq.merchant_id}', '${escapeHtml(req.business_name || '')}', '${escapeHtml(req.terminal_sn || '')}', '${escapeHtml(req.notes || '')}'); event.stopPropagation();" 
                            class="px-3 py-1.5 text-xs font-medium text-white bg-green-600 rounded hover:bg-green-700 transition">
                            Approve
                        </button>
                        <button onclick="openRejectModal('${safeReq.request_id}', '${safeReq.merchant_id}', '${escapeHtml(req.business_name || '')}'); event.stopPropagation();" 
                            class="px-3 py-1.5 text-xs font-medium text-white bg-red-600 rounded hover:bg-red-700 transition">
                            Reject
                        </button>
                    </div>
                ` : `
                    <span class="text-xs text-gray-500">No action</span>
                `}
            </td>
        </tr>
    `}).join('');
}

function updatePagination(totalItems, currentPage, totalPages) {
    const startItem = (currentPage - 1) * 10 + 1;
    const endItem = Math.min(currentPage * 10, totalItems);
    
    document.getElementById('paginationInfo').textContent = 
        `Showing ${startItem} to ${endItem} of ${totalItems} requests`;
    
    document.getElementById('prevBtn').disabled = currentPage === 1;
    document.getElementById('nextBtn').disabled = currentPage === totalPages;
}

function updatePendingCount(requests) {
    const pendingCount = requests.filter(r => r.status === 'pending').length;
    document.getElementById('pendingCount').textContent = pendingCount;
}

function getStatusBadgeClasses(status) {
    switch(status) {
        case 'pending': return 'bg-amber-100 text-amber-800';
        case 'approved': return 'bg-green-100 text-green-800';
        case 'rejected': return 'bg-red-100 text-red-800';
        default: return 'bg-gray-100 text-gray-800';
    }
}

function capitalizeFirst(str) {
    return str.charAt(0).toUpperCase() + str.slice(1);
}

function formatDate(dateString) {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { 
        month: 'short', 
        day: 'numeric', 
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function escapeHtml(text) {
    if (!text) return '';
    const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
    };
    return text.replace(/[&<>"']/g, m => map[m]);
}

// Approve Modal Functions
let approveData = null;
let unassignedTerminals = [];

async function loadUnassignedTerminals() {
    try {
        const response = await fetch(`/v1/admin/${adminUsername}/terminals/unassigned`);
        if (response.ok) {
            const data = await response.json();
            if (data.success && data.data) {
                unassignedTerminals = data.data;
            }
        }
    } catch (e) {
        console.error("Failed to load unassigned terminals", e);
    }
}

async function openApproveModal(requestId, merchantId, businessName, terminalSN, notes) {
    approveData = { requestId, merchantId, businessName, terminalSN };
    document.getElementById('approveRequestId').textContent = requestId;
    document.getElementById('approveMerchantName').textContent = `${businessName} (${merchantId})`;
    document.getElementById('approveMerchantReason').textContent = notes || 'No reason provided';
    
    const selectEl = document.getElementById('approveTerminalSN');
    selectEl.innerHTML = '<option value="" disabled selected>-- Select a terminal --</option>';
    
    // Auto-select if requested specific terminal
    if (terminalSN) {
        const opt = document.createElement('option');
        opt.value = terminalSN;
        opt.textContent = `${terminalSN} (Merchant Requested)`;
        opt.selected = true;
        selectEl.appendChild(opt);
    }
    
    if (unassignedTerminals.length === 0) {
        await loadUnassignedTerminals();
    }
    
    unassignedTerminals.forEach(term => {
        if (term.terminal_sn !== terminalSN) {
            const opt = document.createElement('option');
            opt.value = term.terminal_sn;
            opt.textContent = `${term.terminal_sn} - ${term.device_name}`;
            selectEl.appendChild(opt);
        }
    });

    document.getElementById('approveModal').classList.remove('hidden');
}

function closeApproveModal() {
    document.getElementById('approveModal').classList.add('hidden');
    approveData = null;
}

async function confirmApprove() {
    if (!approveData) return;

    const terminalSN = document.getElementById('approveTerminalSN').value;
    if (!terminalSN) {
        showError('Please select a terminal to assign');
        return;
    }
    
    try {
        const response = await fetch(
            `/v1/admin/${adminUsername}/terminal-requests/${approveData.requestId}/approve`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    assign_terminal_sn: terminalSN,
                    notes: ''
                })
            }
        );

        const data = await response.json();

        if (!response.ok || !data.success) {
            showError(data.message || 'Failed to approve terminal request');
            return;
        }

        showSuccess('Terminal request approved successfully');
        closeApproveModal();
        loadTerminalRequests();

    } catch (error) {
        console.error('Error approving terminal request:', error);
        showError('Failed to approve terminal request');
    }
}

// Reject Modal Functions
let rejectData = null;

function openRejectModal(requestId, merchantId, businessName) {
    rejectData = { requestId, merchantId, businessName };
    document.getElementById('rejectRequestId').textContent = requestId;
    document.getElementById('rejectMerchantName').textContent = `${businessName} (${merchantId})`;
    document.getElementById('rejectReason').value = '';
    document.getElementById('rejectModal').classList.remove('hidden');
}

function closeRejectModal() {
    document.getElementById('rejectModal').classList.add('hidden');
    rejectData = null;
}

async function confirmReject() {
    if (!rejectData) return;

    const reason = document.getElementById('rejectReason').value.trim();
    
    try {
        const response = await fetch(
            `/v1/admin/${adminUsername}/terminal-requests/${rejectData.requestId}/reject`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    reason: reason
                })
            }
        );

        const data = await response.json();

        if (!response.ok || !data.success) {
            showError(data.message || 'Failed to reject terminal request');
            return;
        }

        showSuccess('Terminal request rejected successfully');
        closeRejectModal();
        loadTerminalRequests();

    } catch (error) {
        console.error('Error rejecting terminal request:', error);
        showError('Failed to reject terminal request');
    }
}

// Toast Notifications
function showError(message) {
    showToast(message, 'error');
}

function showSuccess(message) {
    showToast(message, 'success');
}

function showToast(message, type = 'info') {
    const toastId = 'toast-' + Date.now();
    const bgColor = type === 'success' ? 'bg-green-500' : type === 'error' ? 'bg-red-500' : 'bg-blue-500';
    
    const toast = document.createElement('div');
    toast.id = toastId;
    toast.className = `fixed bottom-4 right-4 ${bgColor} text-white px-6 py-3 rounded-lg shadow-lg z-50 animate-fade-in`;
    toast.textContent = message;
    
    document.body.appendChild(toast);
    
    setTimeout(() => {
        toast.remove();
    }, 3000);
}

// Close modals when clicking outside
document.addEventListener('click', function(e) {
    const approveModal = document.getElementById('approveModal');
    const rejectModal = document.getElementById('rejectModal');
    const detailsModal = document.getElementById('requestDetailsModal');
    
    if (e.target === approveModal) {
        closeApproveModal();
    }
    
    if (e.target === rejectModal) {
        closeRejectModal();
    }
    
    if (e.target === detailsModal) {
        closeRequestDetailsModal();
    }
});

// Request Details Modal Functions
window.openRequestDetailsModal = function(dataStr, event) {
    if (event && event.target.closest('button')) return; // Ignore if clicking approve/reject buttons
    
    try {
        const req = JSON.parse(decodeURIComponent(dataStr));
        document.getElementById('detailRequestId').textContent = req.request_id;
        document.getElementById('detailMerchant').textContent = `${req.business_name} (${req.merchant_id})`;
        document.getElementById('detailTerminalSn').textContent = req.terminal_sn;
        document.getElementById('detailRequestedAt').textContent = req.requested_at;
        document.getElementById('detailHandledBy').textContent = req.handled_by;
        document.getElementById('detailNotes').textContent = req.notes;
        
        const statusEl = document.getElementById('detailStatus');
        statusEl.textContent = capitalizeFirst(req.status);
        statusEl.className = `capitalize px-2 py-1 text-xs font-medium rounded-full ${getStatusBadgeClasses(req.status.toLowerCase())}`;
        
        document.getElementById('requestDetailsModal').classList.remove('hidden');
    } catch (e) {
        console.error("Error parsing request details:", e);
    }
}

window.closeRequestDetailsModal = function() {
    document.getElementById('requestDetailsModal').classList.add('hidden');
}

// Add fade-in animation
const style = document.createElement('style');
style.textContent = `
    @keyframes fadeIn {
        from {
            opacity: 0;
            transform: translateY(10px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }
    .animate-fade-in {
        animation: fadeIn 0.3s ease-in-out;
    }
`;
document.head.appendChild(style);
