let unassignedTerminals = [];

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

function fetchMerchantData() {
    const pathParts = window.location.pathname.split('/');
    const adminUsername = pathParts[2];
    const merchantId = pathParts[4];

    fetch(`/v1/admin/${adminUsername}/merchants/${merchantId}/data`)
        .then(res => res.json())
        .then(result => {
            if (result.success && result.data) {
                populateMerchantUI(result.data);
            } else {
                console.error("Failed to load merchant data:", result.message);
                alert("Failed to load merchant data: " + (result.message || "Unknown error"));
            }
        })
        .catch(error => {
            console.error("Error fetching merchant data:", error);
            alert("Network error while loading merchant data.");
        });
}

function populateMerchantUI(merchant) {
    document.getElementById('businessName').textContent = merchant.BusinessName;
    document.getElementById('merchantId').textContent = merchant.MerchantID;
    document.getElementById('businessType').textContent = merchant.BusinessType.replace(/_/g, ' ');
    document.getElementById('registrationNum').textContent = merchant.RegistrationNum || 'N/A';
    document.getElementById('businessAddress').textContent = merchant.BusinessAddress;
    document.getElementById('createdAt').textContent = new Date(merchant.CreatedAt).toLocaleDateString();
    
    document.getElementById('ownerName').textContent = merchant.OwnerName;
    document.getElementById('businessEmail').textContent = merchant.BusinessEmail;
    document.getElementById('businessPhone').textContent = merchant.BusinessPhone;

    const renderDoc = (url, elId) => {
        const el = document.getElementById(elId);
        if (url) {
            el.innerHTML = `<button type="button" onclick="openDocumentViewer('${url}')" class="text-blue-600 hover:text-blue-800 hover:underline font-medium flex items-center gap-1">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"></path></svg>
                View Document
            </button>`;
        } else {
            el.textContent = "Not provided";
        }
    };

    renderDoc(merchant.DtiDocument, 'dtiDocumentLink');
    renderDoc(merchant.BirDocument, 'birDocumentLink');
    renderDoc(merchant.OtherDocument, 'otherDocumentLink');

    const statusEl = document.getElementById('merchantStatus');
    const statusLower = merchant.Status.toLowerCase();
    statusEl.textContent = merchant.Status;
    statusEl.className = 'capitalize px-4 py-2 text-sm font-semibold rounded-full';
    
    if (statusLower === 'active') {
        statusEl.classList.add('bg-green-100', 'text-green-800');
    } else if (statusLower === 'pending_approval' || statusLower === 'pending approval') {
        statusEl.classList.add('bg-yellow-100', 'text-yellow-800');
    } else {
        statusEl.classList.add('bg-red-100', 'text-red-800');
    }

    // Show settlement details regardless of status
    document.getElementById('settlementDetailsContainer').classList.remove('hidden');
    document.getElementById('commissionRate').textContent = merchant.CommissionRate + '%';
    document.getElementById('settlementBank').textContent = merchant.SettlementBank || 'N/A';
    document.getElementById('settlementName').textContent = merchant.SettlementName || 'N/A';
    document.getElementById('settlementAcct').textContent = merchant.SettlementAcct || 'N/A';

    const actionButtons = document.getElementById('actionButtons');
    const btnApprove = document.getElementById('btnApproveMerchant');
    const btnReject = document.getElementById('btnRejectMerchant');
    const btnSuspend = document.getElementById('btnSuspendMerchant');

    if (statusLower === 'pending_approval' || statusLower === 'pending approval' || statusLower === 'active') {
        actionButtons.classList.remove('hidden');
        actionButtons.dataset.merchantId = merchant.MerchantID;
        actionButtons.dataset.businessName = merchant.BusinessName;

        if (statusLower === 'active') {
            btnSuspend.classList.remove('hidden');
            btnApprove.classList.add('hidden');
            btnReject.classList.add('hidden');
        } else {
            btnSuspend.classList.add('hidden');
            btnApprove.classList.remove('hidden');
            btnReject.classList.remove('hidden');
        }
    } else {
        actionButtons.classList.add('hidden');
    }
}

window.openDocumentViewer = function(url) {
    const modal = document.getElementById('documentViewerModal');
    const iframe = document.getElementById('documentViewerFrame');
    const downloadBtn = document.getElementById('downloadDocumentBtn');
    
    if (modal && iframe && downloadBtn) {
        iframe.src = url;
        downloadBtn.href = url;
        modal.classList.remove('hidden');
    }
};

document.addEventListener('DOMContentLoaded', () => {
    fetchUnassignedTerminals();
    fetchMerchantData();

    const btnApprove = document.getElementById('btnApproveMerchant');
    const btnReject = document.getElementById('btnRejectMerchant');
    const actionButtons = document.getElementById('actionButtons');
    const approveModal = document.getElementById('approveMerchantModal');

    if (btnApprove && actionButtons) {
        btnApprove.addEventListener('click', () => {
            const businessName = actionButtons.dataset.businessName;
            const merchantId = actionButtons.dataset.merchantId;
            document.getElementById('approveModalBusinessName').textContent = businessName;
            document.getElementById('approveMerchantId').value = merchantId;

            const terminalSelect = document.getElementById('approveTerminalSn');
            populateTerminalDropdown(terminalSelect);
            approveModal.classList.remove('hidden');
        });
    }

    const rejectModal = document.getElementById('rejectMerchantModal');
    const confirmRejectBtn = document.getElementById('confirmRejectBtn');
    let rejectMerchantId = null;

    if (btnReject && actionButtons) {
        btnReject.addEventListener('click', () => {
            const merchantId = actionButtons.dataset.merchantId;
            const businessName = actionButtons.dataset.businessName;
            if (!merchantId) return;
            
            document.getElementById('rejectModalBusinessName').textContent = businessName;
            rejectMerchantId = merchantId;
            rejectModal.classList.remove('hidden');
        });
    }

    if (confirmRejectBtn) {
        confirmRejectBtn.addEventListener('click', () => {
            if (!rejectMerchantId) return;

            const adminUsername = window.location.pathname.split('/')[2];
            const alertBox = document.getElementById('rejectFormAlert');
            alertBox.classList.add('hidden');
            
            const reason = document.getElementById('rejectReason').value;
            if (!reason) {
                alertBox.classList.remove('hidden', 'bg-green-50', 'text-green-600');
                alertBox.classList.add('bg-red-50', 'text-red-600');
                alertBox.textContent = "Please provide a reason for rejection.";
                return;
            }

            confirmRejectBtn.disabled = true;
            confirmRejectBtn.textContent = 'Rejecting...';

            fetch(`/v1/admin/${adminUsername}/merchants/${rejectMerchantId}/reject`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ reason })
            })
            .then(res => res.json())
            .then(result => {
                alertBox.classList.remove('hidden', 'bg-red-50', 'text-red-600', 'bg-green-50', 'text-green-600');
                if (result.success) {
                    alertBox.classList.add('bg-green-50', 'text-green-600');
                    alertBox.textContent = result.message || "Merchant application rejected successfully.";
                    setTimeout(() => {
                        window.location.reload();
                    }, 1500);
                } else {
                    alertBox.classList.add('bg-red-50', 'text-red-600');
                    alertBox.textContent = result.message || "Failed to reject merchant.";
                    confirmRejectBtn.disabled = false;
                    confirmRejectBtn.textContent = 'Reject Application';
                }
            })
            .catch(err => {
                console.error("Error rejecting merchant:", err);
                alertBox.classList.remove('hidden', 'bg-green-50', 'text-green-600');
                alertBox.classList.add('bg-red-50', 'text-red-600');
                alertBox.textContent = "Network error. Please try again.";
                confirmRejectBtn.disabled = false;
                confirmRejectBtn.textContent = 'Reject Application';
            });
        });
    }

    const btnSuspend = document.getElementById('btnSuspendMerchant');
    const suspendModal = document.getElementById('suspendMerchantModal');
    const confirmSuspendBtn = document.getElementById('confirmSuspendBtn');
    let suspendMerchantId = null;

    if (btnSuspend && actionButtons) {
        btnSuspend.addEventListener('click', () => {
            const merchantId = actionButtons.dataset.merchantId;
            const businessName = actionButtons.dataset.businessName;
            if (!merchantId) return;
            
            document.getElementById('suspendModalBusinessName').textContent = businessName;
            suspendMerchantId = merchantId;
            suspendModal.classList.remove('hidden');
        });
    }

    if (confirmSuspendBtn) {
        confirmSuspendBtn.addEventListener('click', () => {
            if (!suspendMerchantId) return;

            const adminUsername = window.location.pathname.split('/')[2];
            const alertBox = document.getElementById('suspendFormAlert');
            alertBox.classList.add('hidden');
            
            const reason = document.getElementById('suspendReason').value;
            if (!reason) {
                alertBox.classList.remove('hidden', 'bg-green-50', 'text-green-600');
                alertBox.classList.add('bg-red-50', 'text-red-600');
                alertBox.textContent = "Please provide a reason for suspension.";
                return;
            }

            confirmSuspendBtn.disabled = true;
            confirmSuspendBtn.textContent = 'Suspending...';

            fetch(`/v1/admin/${adminUsername}/merchants/${suspendMerchantId}/suspend`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ reason })
            })
            .then(res => res.json())
            .then(result => {
                alertBox.classList.remove('hidden', 'bg-red-50', 'text-red-600', 'bg-green-50', 'text-green-600');
                if (result.success) {
                    alertBox.classList.add('bg-green-50', 'text-green-600');
                    alertBox.textContent = result.message || "Merchant account suspended successfully.";
                    setTimeout(() => {
                        window.location.reload();
                    }, 1500);
                } else {
                    alertBox.classList.add('bg-red-50', 'text-red-600');
                    alertBox.textContent = result.message || "Failed to suspend merchant.";
                    confirmSuspendBtn.disabled = false;
                    confirmSuspendBtn.textContent = 'Suspend Account';
                }
            })
            .catch(err => {
                console.error("Error suspending merchant:", err);
                alertBox.classList.remove('hidden', 'bg-green-50', 'text-green-600');
                alertBox.classList.add('bg-red-50', 'text-red-600');
                alertBox.textContent = "Network error. Please try again.";
                confirmSuspendBtn.disabled = false;
                confirmSuspendBtn.textContent = 'Suspend Account';
            });
        });
    }

    // Terminal selection auto-fill for approve form
    const approveTerminalSn = document.getElementById('approveTerminalSn');
    const approveDeviceName = document.getElementById('approveDeviceName');
    if (approveTerminalSn && approveDeviceName) {
        approveTerminalSn.addEventListener('change', (e) => {
            const selectedOption = e.target.options[e.target.selectedIndex];
            if (selectedOption && selectedOption.dataset.deviceName) {
                approveDeviceName.value = selectedOption.dataset.deviceName;
            } else {
                approveDeviceName.value = '';
            }
        });
    }

    const approveForm = document.getElementById('approveForm');
    if (approveForm) {
        approveForm.addEventListener('submit', (e) => {
            e.preventDefault();
            
            const merchantId = document.getElementById('approveMerchantId').value;
            const commissionRate = document.getElementById('approveCommissionRate').value;
            const terminalSn = document.getElementById('approveTerminalSn').value;
            const deviceName = document.getElementById('approveDeviceName').value;

            const payload = {
                commissionRate,
                terminalSn,
                deviceName
            };

            const alertBox = document.getElementById('approveFormAlert');
            alertBox.classList.add('hidden');

            const adminUsername = window.location.pathname.split('/')[2];
            fetch(`/v1/admin/${adminUsername}/merchants/${merchantId}/approve`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            })
            .then(res => res.json())
            .then(result => {
                alertBox.classList.remove('hidden', 'bg-red-50', 'text-red-600', 'bg-green-50', 'text-green-600');
                if (result.success) {
                    alertBox.classList.add('bg-green-50', 'text-green-600');
                    alertBox.textContent = result.message || "Merchant approved successfully!";
                    setTimeout(() => {
                        window.location.reload();
                    }, 1500);
                } else {
                    alertBox.classList.add('bg-red-50', 'text-red-600');
                    alertBox.textContent = result.message || "An error occurred.";
                }
            })
            .catch(err => {
                console.error("Error approving merchant:", err);
                alertBox.classList.remove('hidden', 'bg-green-50', 'text-green-600');
                alertBox.classList.add('bg-red-50', 'text-red-600');
                alertBox.textContent = "Network error. Please try again.";
            });
        });
    }
});
