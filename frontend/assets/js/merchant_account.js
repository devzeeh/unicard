document.addEventListener("DOMContentLoaded", () => {
    if (!window.CURRENT_USERNAME) {
        console.error("CURRENT_USERNAME is not defined. Cannot fetch account data.");
        return;
    }

    const fetchAccountData = async () => {
        try {
            const response = await fetch(`/v1/merchant/${window.CURRENT_USERNAME}/account`);
            const json = await response.json();

            if (json.success && json.data) {
                const data = json.data;
                const details = data.business_details;
                const bank = data.business_bank_details;
                const docs = data.business_document || [];

                // Basic Info Header
                const merchantIdEl = document.getElementById('displayMerchantId');
                if (merchantIdEl) merchantIdEl.textContent = data.merchant_id || 'N/A';
                
                const memberSinceEl = document.getElementById('displayMemberSince');
                if (memberSinceEl) memberSinceEl.textContent = data.member_since || 'N/A';

                const statusEl = document.getElementById('displayAccountStatus');
                if (statusEl) {
                    const isOk = data.account_status.toLowerCase() === 'active' || data.account_status.toLowerCase() === 'approved';
                    statusEl.innerHTML = `<span class="px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider ${isOk ? 'bg-emerald-100 text-emerald-800' : 'bg-amber-100 text-amber-800'}">${data.account_status}</span>`;
                }

                // Business Details Form
                if (document.getElementById('bizName')) document.getElementById('bizName').value = details.business_name || '';
                if (document.getElementById('bizType')) document.getElementById('bizType').value = details.business_type || '';
                if (document.getElementById('bizEmail')) document.getElementById('bizEmail').value = details.business_email || '';
                if (document.getElementById('bizPhone')) document.getElementById('bizPhone').value = details.business_phone || '';
                if (document.getElementById('bizAddress')) document.getElementById('bizAddress').value = details.business_address || '';
                if (document.getElementById('bizCity')) document.getElementById('bizCity').value = details.city || '';
                if (document.getElementById('bizPostal')) document.getElementById('bizPostal').value = details.postal_code || '';

                // Bank Details
                if (document.getElementById('bankName')) document.getElementById('bankName').value = bank.bank_name || '';
                if (document.getElementById('bankHolder')) document.getElementById('bankHolder').value = bank.account_holder_name || '';
                if (document.getElementById('bankAccount')) document.getElementById('bankAccount').value = bank.account_number || '';

                // Documents
                const docsContainer = document.getElementById('documentsContainer');
                if (docsContainer) {
                    if (docs.length === 0) {
                        docsContainer.innerHTML = '<p class="text-sm text-gray-500 italic">No documents uploaded.</p>';
                    } else {
                        docsContainer.innerHTML = '';
                        docs.forEach(doc => {
                            const isApproved = doc.status.toLowerCase() === 'approved';
                            const isPending = doc.status.toLowerCase() === 'pending';
                            let statusColor = 'text-red-600 bg-red-50';
                            let icon = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />';
                            if (isApproved) {
                                statusColor = 'text-emerald-600 bg-emerald-50';
                                icon = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />';
                            } else if (isPending) {
                                statusColor = 'text-amber-600 bg-amber-50';
                                icon = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />';
                            }

                            const docDiv = document.createElement('div');
                            docDiv.className = 'flex items-center justify-between p-4 border border-gray-100 rounded-xl bg-gray-50/50 hover:bg-gray-50 transition-colors';
                            docDiv.innerHTML = `
                                <div class="flex items-center space-x-4">
                                    <div class="h-10 w-10 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center">
                                        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                                    </div>
                                    <div>
                                        <p class="text-sm font-bold text-gray-900">${doc.document_type}</p>
                                        ${doc.message ? `<p class="text-xs text-gray-500 mt-0.5">${doc.message}</p>` : ''}
                                    </div>
                                </div>
                                <div class="flex items-center space-x-3">
                                    <span class="inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-semibold uppercase tracking-wider ${statusColor}">
                                        <svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">${icon}</svg>
                                        ${doc.status}
                                    </span>
                                    ${doc.document_url ? `<a href="${doc.document_url}" target="_blank" class="text-blue-600 hover:text-blue-800 p-2 hover:bg-blue-50 rounded-lg transition-colors"><svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg></a>` : ''}
                                </div>
                            `;
                            docsContainer.appendChild(docDiv);
                        });
                    }
                }
                
                // Remove all animate-pulse classes to reveal inputs
                document.querySelectorAll('.animate-pulse').forEach(el => {
                    el.classList.remove('animate-pulse');
                    if(el.tagName === 'DIV' && el.classList.contains('h-10')) {
                        el.outerHTML = ''; // We will replace the pulse divs with actual inputs in the HTML
                    }
                });
                
                // Reveal the form content
                const formContent = document.getElementById('accountFormContent');
                if (formContent) {
                    formContent.classList.remove('opacity-0');
                    formContent.classList.add('opacity-100');
                }
                const loadingSkeleton = document.getElementById('loadingSkeleton');
                if (loadingSkeleton) {
                    loadingSkeleton.style.display = 'none';
                }

            } else {
                console.error("Failed to fetch account data:", json.message);
            }
        } catch (error) {
            console.error("Error fetching account data:", error);
        }
    };

    fetchAccountData();
});
