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

                // Document Status and Message
                const docStatusEl = document.getElementById('displayDocumentStatus');
                if (docStatusEl && data.document_status) {
                    const ds = data.document_status.toLowerCase();
                    let dsColor = 'text-amber-600 bg-amber-100';
                    if (ds === 'approved') dsColor = 'text-emerald-600 bg-emerald-100';
                    else if (ds === 'rejected') dsColor = 'text-red-600 bg-red-100';
                    docStatusEl.innerHTML = `<span class="px-2.5 py-1 rounded-lg text-xs font-bold uppercase tracking-wider ${dsColor}">Docs: ${data.document_status}</span>`;
                }

                const msgContainer = document.getElementById('docMessageContainer');
                const msgEl = document.getElementById('displayAccountMessage');
                if (msgContainer && msgEl && data.account_message && data.account_message.trim() !== '') {
                    msgEl.textContent = data.account_message;
                    msgContainer.classList.remove('hidden');
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

                let realAccountNumber = bank.account_number || '';
                let maskedAccountNumber = realAccountNumber;
                if (realAccountNumber.length > 4) {
                    maskedAccountNumber = "**** **** **** " + realAccountNumber.slice(-4);
                }
                
                const bankInput = document.getElementById('bankAccount');
                if (bankInput) {
                    bankInput.value = maskedAccountNumber;
                    bankInput.dataset.realValue = realAccountNumber;
                }

                const actionContainer = document.getElementById('bankDetailsActionContainer');
                const editBtn = document.getElementById('editBankDetailsBtn');
                const saveBtn = document.getElementById('saveBankDetailsBtn');
                const cancelBtn = document.getElementById('bankDetailsCancelBtn');
                if (actionContainer) actionContainer.classList.remove('hidden');

                const initialBankDetails = {
                    bankName: bank.bank_name || '',
                    bankHolder: bank.account_holder_name || '',
                    bankAccount: bank.account_number || ''
                };

                // Toggle Account Number visibility
                const toggleBtn = document.getElementById('toggleBankAccount');
                const eyeOpen = document.getElementById('eyeIconOpen');
                const eyeClosed = document.getElementById('eyeIconClosed');
                let isAccountMasked = true;

                if (toggleBtn && bankInput && eyeOpen && eyeClosed) {
                    toggleBtn.addEventListener('click', () => {
                        if (isAccountMasked) {
                            // Show real number
                            bankInput.value = bankInput.dataset.realValue || '';
                            eyeOpen.classList.remove('hidden');
                            eyeClosed.classList.add('hidden');
                            isAccountMasked = false;
                        } else {
                            // Mask the number
                            let currentVal = bankInput.value;
                            let masked = currentVal;
                            if (currentVal.length > 4) {
                                masked = "**** **** **** " + currentVal.slice(-4);
                            }
                            bankInput.value = masked;
                            eyeOpen.classList.add('hidden');
                            eyeClosed.classList.remove('hidden');
                            isAccountMasked = true;
                        }
                    });

                    // Update realValue as they type when unmasked
                    bankInput.addEventListener('input', (e) => {
                        if (!isAccountMasked) {
                            bankInput.dataset.realValue = e.target.value;
                            checkEnableBankSave();
                        }
                    });
                }

                function checkEnableBankSave() {
                    if (!saveBtn) return;
                    const currentAccountVal = isAccountMasked ? bankInput.dataset.realValue : document.getElementById('bankAccount').value;
                    const current = {
                        bankName: document.getElementById('bankName').value,
                        bankHolder: document.getElementById('bankHolder').value,
                        bankAccount: currentAccountVal
                    };
                    const isChanged = current.bankName !== initialBankDetails.bankName ||
                                      current.bankHolder !== initialBankDetails.bankHolder ||
                                      current.bankAccount !== initialBankDetails.bankAccount;
                    if (isChanged) {
                        saveBtn.disabled = false;
                        saveBtn.classList.remove('opacity-50', 'cursor-not-allowed');
                    } else {
                        saveBtn.disabled = true;
                        saveBtn.classList.add('opacity-50', 'cursor-not-allowed');
                    }
                }

                if (editBtn) {
                    editBtn.addEventListener('click', () => {
                        ['bankName', 'bankHolder', 'bankAccount'].forEach(id => {
                            const el = document.getElementById(id);
                            if (el) {
                                el.removeAttribute('readonly');
                                el.removeAttribute('disabled');
                                el.classList.remove('bg-gray-50', 'text-gray-600', 'cursor-not-allowed');
                                el.classList.add('bg-white', 'text-gray-900', 'ring-2', 'ring-indigo-100');
                                // Only cursor-pointer for non-text inputs, but here select needs it
                                if (el.tagName === 'SELECT') el.classList.add('cursor-pointer');
                                if (id !== 'bankAccount') {
                                    el.addEventListener('input', checkEnableBankSave);
                                }
                                el.addEventListener('change', checkEnableBankSave);
                            }
                        });
                        
                        // Force unmask when editing begins
                        if (isAccountMasked && toggleBtn) {
                            toggleBtn.click();
                        }

                        editBtn.classList.add('hidden');
                        saveBtn.classList.remove('hidden');
                        if (cancelBtn) cancelBtn.classList.remove('hidden');
                        checkEnableBankSave();
                    });
                }
                if (cancelBtn) {
                    cancelBtn.addEventListener('click', () => window.location.reload());
                }

                if (saveBtn) {
                    saveBtn.addEventListener('click', async () => {
                        // Reset errors
                        const errorDiv = document.getElementById('bankDetailsError');
                        if (errorDiv) errorDiv.classList.add('hidden');
                        ['bankName', 'bankHolder', 'bankAccount'].forEach(id => {
                            const el = document.getElementById(id);
                            if (el) {
                                el.classList.remove('ring-red-500', 'border-red-500');
                            }
                        });

                        const finalAccountNumber = isAccountMasked ? bankInput.dataset.realValue : document.getElementById('bankAccount').value;
                        const newBank = {
                            bank_name: document.getElementById('bankName').value,
                            account_holder_name: document.getElementById('bankHolder').value,
                            account_number: finalAccountNumber
                        };
                        
                        let hasError = false;
                        if (!newBank.bank_name) {
                            document.getElementById('bankName').classList.add('ring-red-500', 'border-red-500');
                            hasError = true;
                        }
                        if (!newBank.account_holder_name) {
                            document.getElementById('bankHolder').classList.add('ring-red-500', 'border-red-500');
                            hasError = true;
                        }
                        if (!newBank.account_number) {
                            document.getElementById('bankAccount').classList.add('ring-red-500', 'border-red-500');
                            hasError = true;
                        }

                        if (hasError) {
                            if (errorDiv) {
                                errorDiv.textContent = 'Please fill in all required fields marked with *';
                                errorDiv.classList.remove('hidden');
                            }
                            return;
                        }

                        saveBtn.textContent = 'Saving...';
                        saveBtn.disabled = true;
                        try {
                            const res = await fetch(`/v1/merchant/${window.CURRENT_USERNAME}/update-bank`, {
                                method: 'POST',
                                headers: { 'Content-Type': 'application/json' },
                                body: JSON.stringify(newBank)
                            });
                            const j = await res.json();
                            if (j.success) {
                                alert('Bank details updated successfully!');
                                window.location.reload();
                            } else {
                                alert('Failed: ' + j.message);
                                saveBtn.textContent = 'Save Changes';
                                saveBtn.disabled = false;
                            }
                        } catch (e) {
                            console.error(e);
                            alert('Network error');
                            saveBtn.textContent = 'Save Changes';
                            saveBtn.disabled = false;
                        }
                    });
                }

                // Documents
                const docsContainer = document.getElementById('documentsContainer');
                if (docsContainer) {
                    docsContainer.innerHTML = '';
                    
                    const structure = details.business_structure || '';
                    const regLabel = 'DTI/SEC Registration';
                    
                    const expectedDocs = [
                        { type: regLabel, key: 'business_document' },
                        { type: 'BIR Certificate', key: 'bir_document' },
                        { type: 'Other Document', key: 'other_document' }
                    ];

                    expectedDocs.forEach(ed => {
                        // Find if merchant has uploaded this doc type
                        const found = docs.find(d => d.document_type === ed.type);
                        
                        let docType = ed.type;
                        let status = found ? (found.document_status || 'Pending') : 'Missing';
                        let message = found ? found.message : 'Please upload this document';
                        let docUrl = found ? found.document_url : '';
                        
                        const isApproved = status.toLowerCase() === 'approved';
                        const isPending = status.toLowerCase() === 'pending';
                        const isMissing = status.toLowerCase() === 'missing';
                        
                        let statusColor = 'text-red-600 bg-red-50';
                        let icon = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />';
                        if (isApproved) {
                            statusColor = 'text-emerald-600 bg-emerald-50';
                            icon = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />';
                        } else if (isPending) {
                            statusColor = 'text-amber-600 bg-amber-50';
                            icon = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />';
                        }

                        const isImage = docUrl && docUrl.match(/\.(jpeg|jpg|gif|png|webp)$/i);
                        const defaultIconHtml = `<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>`;
                        const previewHtml = isImage ? `<img src="${docUrl}" class="h-full w-full object-cover" alt="${docType} Preview">` : defaultIconHtml;

                        const docDiv = document.createElement('div');
                        docDiv.className = 'flex items-center justify-between p-4 border border-gray-100 rounded-xl bg-gray-50/50 hover:bg-gray-50 transition-colors';
                        docDiv.innerHTML = `
                            <div class="flex items-center space-x-4">
                                <div class="doc-preview h-14 w-14 flex-shrink-0 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center overflow-hidden border border-blue-100/60 shadow-sm">
                                    ${previewHtml}
                                </div>
                                <div>
                                    <p class="text-sm font-bold text-gray-900">${docType}</p>
                                    ${message ? `<p class="text-xs text-gray-500 mt-0.5">${message}</p>` : ''}
                                </div>
                            </div>
                            <div class="flex items-center space-x-3">
                                <span class="inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-semibold uppercase tracking-wider ${statusColor}">
                                    <svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">${icon}</svg>
                                    ${status}
                                </span>
                                ${docUrl ? `<button type="button" onclick="openDocumentViewer('${docUrl.replace(/\\/g, '/').replace(/'/g, "\\'")}')" class="text-blue-600 hover:text-blue-800 p-2 hover:bg-blue-50 rounded-lg transition-colors"><svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg></button>` : ''}
                                ${!isApproved ? `
                                <div class="w-72 flex flex-col gap-2">
                                    <label class="doc-dropzone flex flex-col items-center justify-center w-full h-24 border-2 border-slate-300 border-dashed rounded-lg bg-slate-50 transition-colors pointer-events-none opacity-60">
                                        <div class="flex flex-col items-center justify-center pt-3 pb-4 text-center px-2">
                                            <svg class="w-6 h-6 mb-2 text-slate-400" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 20 16"><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 13h3a3 3 0 0 0 0-6h-.025A5.56 5.56 0 0 0 16 6.5 5.5 5.5 0 0 0 5.207 5.021C5.137 5.017 5.071 5 5 5a4 4 0 0 0 0 8h2.167M10 15V6m0 0L8 8m2-2 2 2"/></svg>
                                            <p class="mb-1 text-xs text-slate-600 font-semibold dropzone-text truncate w-full">Click or drag & drop</p>
                                            <p class="text-[10px] text-slate-400 dropzone-hint">PDF, JPG, PNG (Max 4MB)</p>
                                        </div>
                                        <input type="file" accept="image/*,.pdf,.doc,.docx,.webp" class="hidden doc-upload-input" data-doctype="${docType}" disabled>
                                    </label>
                                </div>
                                ` : ''}
                            </div>
                        `;
                        docsContainer.appendChild(docDiv);
                    });
                }
                
                // Store actual selected files securely so native cancel doesn't erase them
                const selectedFilesMap = new Map();
                const globalDocBtn = document.getElementById('globalDocUploadBtn');
                let isDocEditMode = false;
                
                function checkEnableSave() {
                    const hasFiles = selectedFilesMap.size > 0;
                    if (globalDocBtn) {
                        globalDocBtn.disabled = !hasFiles;
                        if (hasFiles) {
                            globalDocBtn.classList.remove('opacity-50', 'cursor-not-allowed');
                            globalDocBtn.classList.add('hover:bg-indigo-700');
                        } else {
                            globalDocBtn.classList.add('opacity-50', 'cursor-not-allowed');
                            globalDocBtn.classList.remove('hover:bg-indigo-700');
                        }
                    }
                }

                const globalDocCancelBtn = document.getElementById('globalDocCancelBtn');
                
                if (globalDocCancelBtn) {
                    globalDocCancelBtn.addEventListener('click', () => window.location.reload());
                }

                if (globalDocBtn) {
                    globalDocBtn.addEventListener('click', async () => {
                        if (!isDocEditMode) {
                            // Enter edit mode
                            isDocEditMode = true;
                            globalDocBtn.textContent = 'Save Changes';
                            globalDocBtn.classList.remove('bg-gray-100', 'text-gray-700', 'hover:bg-gray-200');
                            globalDocBtn.classList.add('bg-indigo-600', 'text-white', 'opacity-50', 'cursor-not-allowed');
                            globalDocBtn.disabled = true;
                            
                            if (globalDocCancelBtn) globalDocCancelBtn.classList.remove('hidden');
                            
                            // Enable dropzones
                            document.querySelectorAll('.doc-dropzone').forEach(dz => {
                                dz.classList.remove('pointer-events-none', 'opacity-60');
                                dz.classList.add('cursor-pointer', 'hover:bg-slate-100');
                                const input = dz.querySelector('input');
                                if (input) input.disabled = false;
                            });
                        } else {
                            // Save Changes mode (Upload)
                            if (selectedFilesMap.size === 0) return;
                            
                            globalDocBtn.textContent = 'Uploading...';
                            globalDocBtn.disabled = true;
                            globalDocBtn.classList.add('opacity-70');
                            
                            let successCount = 0;
                            let failCount = 0;
                            
                            for (const [docType, file] of selectedFilesMap.entries()) {
                                const formData = new FormData();
                                formData.append('document', file);
                                formData.append('document_type', docType);
                                
                                try {
                                    const res = await fetch(`/v1/merchant/${window.CURRENT_USERNAME}/upload-document`, {
                                        method: 'POST',
                                        body: formData
                                    });
                                    const j = await res.json();
                                    if (j.success) successCount++;
                                    else {
                                        failCount++;
                                        console.error(j.message);
                                    }
                                } catch (err) {
                                    console.error(err);
                                    failCount++;
                                }
                            }
                            
                            if (failCount === 0) {
                                alert('Documents uploaded successfully!');
                            } else {
                                alert(`Uploaded ${successCount} documents. Failed ${failCount}.`);
                            }
                            window.location.reload();
                        }
                    });
                }

                document.querySelectorAll('.doc-dropzone').forEach(dropzone => {
                    const input = dropzone.querySelector('input');
                    const text = dropzone.querySelector('.dropzone-text');
                    const hint = dropzone.querySelector('.dropzone-hint');
                    const svgIcon = dropzone.querySelector('svg');
                    
                    const originalSvg = '<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 13h3a3 3 0 0 0 0-6h-.025A5.56 5.56 0 0 0 16 6.5 5.5 5.5 0 0 0 5.207 5.021C5.137 5.017 5.071 5 5 5a4 4 0 0 0 0 8h2.167M10 15V6m0 0L8 8m2-2 2 2"/>';
                    const successSvg = '<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />';
                    
                    function resetDropzone() {
                        text.textContent = 'Click or drag & drop';
                        text.classList.remove('text-emerald-600', 'font-bold');
                        hint.textContent = 'PDF, JPG, PNG (Max 4MB)';
                        hint.classList.remove('text-emerald-500', 'font-medium');
                        if(svgIcon) {
                            svgIcon.innerHTML = originalSvg;
                            svgIcon.classList.remove('text-emerald-500');
                            svgIcon.classList.add('text-slate-400');
                        }
                        dropzone.classList.remove('border-indigo-400', 'bg-indigo-50', 'border-emerald-400', 'bg-emerald-50');
                    }
                    
                    function setSuccessDropzone(filename) {
                        text.textContent = filename;
                        text.title = filename;
                        text.classList.add('text-emerald-600', 'font-bold');
                        hint.textContent = 'Ready to save';
                        hint.classList.add('text-emerald-500', 'font-medium');
                        if(svgIcon) {
                            svgIcon.innerHTML = successSvg;
                            svgIcon.classList.remove('text-slate-400');
                            svgIcon.classList.add('text-emerald-500');
                        }
                        dropzone.classList.remove('border-indigo-400', 'bg-indigo-50');
                        dropzone.classList.add('border-emerald-400', 'bg-emerald-50');
                    }

                    dropzone.addEventListener('dragover', (e) => {
                        e.preventDefault();
                        if (input && !input.disabled) {
                            dropzone.classList.add('bg-indigo-50', 'border-indigo-300');
                        }
                    });
                    
                    dropzone.addEventListener('dragleave', (e) => {
                        e.preventDefault();
                        if (input && !input.disabled && !selectedFilesMap.has(input.dataset.doctype)) {
                            dropzone.classList.remove('bg-indigo-50', 'border-indigo-300');
                        }
                    });
                    
                    dropzone.addEventListener('drop', (e) => {
                        e.preventDefault();
                        if (input && !input.disabled && e.dataTransfer.files && e.dataTransfer.files.length > 0) {
                            const file = e.dataTransfer.files[0];
                            processFile(file, input.dataset.doctype);
                        } else if (!selectedFilesMap.has(input.dataset.doctype)) {
                            dropzone.classList.remove('bg-indigo-50', 'border-indigo-300');
                        }
                    });
                    
                    if (input) {
                        input.addEventListener('change', (e) => {
                            const file = e.target.files[0];
                            if (!file) {
                                // Ignore native cancel event to preserve previously stored file
                                return;
                            }
                            processFile(file, input.dataset.doctype);
                            // Reset input so the same file can trigger a change event again if needed
                            e.target.value = '';
                        });
                    }

                    function updatePreview(file) {
                        const previewContainer = dropzone.closest('.flex.items-center.justify-between').querySelector('.doc-preview');
                        if (previewContainer) {
                            if (file && file.type.startsWith('image/')) {
                                const url = URL.createObjectURL(file);
                                previewContainer.innerHTML = `<img src="${url}" class="h-full w-full object-cover" alt="New Preview">`;
                            } else {
                                previewContainer.innerHTML = `<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>`;
                            }
                        }
                    }

                    function processFile(file, docType) {
                        const maxSize = 4 * 1024 * 1024;
                        if (file.size > maxSize) {
                            alert('File is too large. Please upload a file smaller than 4MB.');
                            if (!selectedFilesMap.has(docType)) resetDropzone();
                            return;
                        }

                        const validTypes = ['image/jpeg', 'image/png', 'image/webp', 'application/pdf', 'application/msword', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'];
                        if (!validTypes.includes(file.type)) {
                            alert('Invalid file format. Please upload an image, PDF, or Word document.');
                            if (!selectedFilesMap.has(docType)) resetDropzone();
                            return;
                        }
                        
                        selectedFilesMap.set(docType, file);
                        setSuccessDropzone(file.name);
                        checkEnableSave();
                        updatePreview(file);
                    }
                });
                
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
            } else {
                console.error("Failed to fetch account data:", json.message);
            }
        } catch (error) {
            console.error("Error fetching account data:", error);
        } finally {
            const loadingSkeleton = document.getElementById('loadingSkeleton');
            if (loadingSkeleton) {
                loadingSkeleton.style.display = 'none';
            }
        }
    };

    fetchAccountData();
});

let currentZoom = 1;
let isDragging = false;
let startX = 0, startY = 0, translateX = 0, translateY = 0;

window.openDocumentViewer = function(url) {
    const modal = document.getElementById('documentViewerModal');
    const iframe = document.getElementById('documentViewerFrame');
    const img = document.getElementById('documentViewerImage');
    const zoomControls = document.getElementById('imageZoomControls');
    
    if (modal && iframe && img) {
        const isImage = url.match(/\.(jpeg|jpg|gif|png|webp)$/i);
        if (isImage) {
            iframe.classList.add('hidden');
            img.classList.remove('hidden');
            if (zoomControls) zoomControls.classList.remove('hidden');
            img.src = url;
            resetZoom();
        } else {
            img.classList.add('hidden');
            iframe.classList.remove('hidden');
            if (zoomControls) zoomControls.classList.add('hidden');
            iframe.src = url + '#toolbar=0&navpanes=0&scrollbar=0';
        }
        modal.classList.remove('hidden');
    }
};

function resetZoom() {
    currentZoom = 1;
    translateX = 0;
    translateY = 0;
    updateZoomTransform();
}

function updateZoomTransform() {
    const img = document.getElementById('documentViewerImage');
    const zoomLevelEl = document.getElementById('zoomLevel');
    if (img) {
        img.style.transform = `translate(${translateX}px, ${translateY}px) scale(${currentZoom})`;
    }
    if (zoomLevelEl) {
        zoomLevelEl.textContent = Math.round(currentZoom * 100) + '%';
    }
}

document.addEventListener('DOMContentLoaded', () => {
    // Zoom Controls setup (will attach if elements exist)
    const zoomInBtn = document.getElementById('zoomInBtn');
    const zoomOutBtn = document.getElementById('zoomOutBtn');
    const zoomResetBtn = document.getElementById('zoomResetBtn');
    const img = document.getElementById('documentViewerImage');
    
    if (zoomInBtn) zoomInBtn.addEventListener('click', () => { currentZoom = Math.min(currentZoom + 0.25, 4); updateZoomTransform(); });
    if (zoomOutBtn) zoomOutBtn.addEventListener('click', () => { currentZoom = Math.max(currentZoom - 0.25, 0.5); updateZoomTransform(); });
    if (zoomResetBtn) zoomResetBtn.addEventListener('click', resetZoom);
    
    if (img) {
        img.parentElement.addEventListener('wheel', (e) => {
            if (img.classList.contains('hidden')) return;
            e.preventDefault();
            if (e.deltaY < 0) {
                currentZoom = Math.min(currentZoom + 0.1, 4);
            } else {
                currentZoom = Math.max(currentZoom - 0.1, 0.5);
            }
            updateZoomTransform();
        });

        img.parentElement.addEventListener('mousedown', (e) => {
            if (img.classList.contains('hidden')) return;
            isDragging = true;
            img.parentElement.style.cursor = 'grabbing';
            startX = e.clientX - translateX;
            startY = e.clientY - translateY;
        });

        window.addEventListener('mousemove', (e) => {
            if (!isDragging) return;
            translateX = e.clientX - startX;
            translateY = e.clientY - startY;
            updateZoomTransform();
        });

        window.addEventListener('mouseup', () => {
            isDragging = false;
            if (img) img.parentElement.style.cursor = 'default';
        });
    }
});
