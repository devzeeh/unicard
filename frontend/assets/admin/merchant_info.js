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

    if (statusLower === 'active') {
        document.getElementById('settlementDetailsContainer').classList.remove('hidden');
        document.getElementById('commissionRate').textContent = merchant.CommissionRate + '%';
        document.getElementById('settlementBank').textContent = merchant.SettlementBank || 'N/A';
        document.getElementById('settlementName').textContent = merchant.SettlementName || 'N/A';
        document.getElementById('settlementAcct').textContent = merchant.SettlementAcct || 'N/A';
    }

    if (statusLower === 'pending_approval' || statusLower === 'pending approval') {
        const actionButtons = document.getElementById('actionButtons');
        actionButtons.classList.remove('hidden');
        actionButtons.dataset.merchantId = merchant.MerchantID;
        actionButtons.dataset.businessName = merchant.BusinessName;
    }
}

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

    if (btnReject && actionButtons) {
        btnReject.addEventListener('click', () => {
            const merchantId = actionButtons.dataset.merchantId;
            if (!merchantId) return;
            
            if (!confirm("Are you sure you want to reject this merchant application?")) return;

            const adminUsername = window.location.pathname.split('/')[2];
            fetch(`/v1/admin/${adminUsername}/merchants/${merchantId}/reject`, {
                method: 'POST'
            })
            .then(res => res.json())
            .then(result => {
                if (result.success) {
                    alert("Merchant application rejected successfully.");
                    window.location.reload();
                } else {
                    alert(result.message || "Failed to reject merchant.");
                }
            })
            .catch(err => {
                console.error("Error rejecting merchant:", err);
                alert("Network error. Please try again.");
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
            const settlementBank = document.getElementById('approveSettlementBank').value;
            const settlementName = document.getElementById('approveSettlementName').value;
            const settlementAccount = document.getElementById('approveSettlementAccount').value;
            const terminalSn = document.getElementById('approveTerminalSn').value;
            const deviceName = document.getElementById('approveDeviceName').value;

            const payload = {
                commissionRate,
                settlementBank,
                settlementName,
                settlementAccount,
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
