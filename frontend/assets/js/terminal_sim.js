document.addEventListener('DOMContentLoaded', () => {
    const simForm = document.getElementById('simForm');
    const submitBtn = document.getElementById('submitBtn');
    const simMessage = document.getElementById('simMessage');
    const activityTable = document.getElementById('activityTable');
    const emptyState = document.getElementById('emptyState');
    const activityCount = document.getElementById('activityCount');
    
    let txCount = 0;

    simForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const cardNumber = document.getElementById('cardNumber').value.trim();
        const type = document.querySelector('input[name="type"]:checked').value;
        const amount = parseFloat(document.getElementById('amount').value);
        const merchantId = document.getElementById('merchantId').value;

        if(!cardNumber || !amount || !merchantId) return;

        // Loading state
        const originalBtnText = submitBtn.innerHTML;
        submitBtn.innerHTML = `<svg class="animate-spin -ml-1 mr-2 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg> Processing...`;
        submitBtn.disabled = true;
        simMessage.classList.add('hidden');

        try {
            const res = await fetch('/v1/terminal-sim/transact', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ card_number: cardNumber, type, amount, merchant_id: merchantId })
            });

            const data = await res.json();
            
            simMessage.classList.remove('hidden');
            if(data.success) {
                simMessage.className = 'text-sm rounded-lg p-3 text-center bg-green-50 text-green-700 border border-green-200 mt-4';
                simMessage.textContent = `Success: Deducted ₱${amount.toFixed(2)} (Fee: ₱${(data.service_fee || 0).toFixed(2)})`;
                
                // Add to table
                addTxToTable(type, amount, data.service_fee || 0, 'Success');
                simForm.reset();
                document.querySelector('input[name="type"][value="Fare"]').checked = true; // reset radio
            } else {
                simMessage.className = 'text-sm rounded-lg p-3 text-center bg-red-50 text-red-700 border border-red-200 mt-4';
                simMessage.textContent = data.message || 'Transaction failed';
                addTxToTable(type, amount, 0, 'Failed');
            }
        } catch(err) {
            console.error(err);
            simMessage.classList.remove('hidden');
            simMessage.className = 'text-sm rounded-lg p-3 text-center bg-red-50 text-red-700 border border-red-200 mt-4';
            simMessage.textContent = 'Network error occurred.';
            addTxToTable(type, amount, 0, 'Error');
        } finally {
            submitBtn.innerHTML = originalBtnText;
            submitBtn.disabled = false;
        }
    });

    function addTxToTable(type, amount, fee, status) {
        emptyState.style.display = 'none';
        txCount++;
        activityCount.textContent = `${txCount} transaction${txCount > 1 ? 's' : ''}`;

        const row = document.createElement('tr');
        row.className = "hover:bg-slate-50 transition";
        
        const timeStr = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second:'2-digit' });
        
        let statusBadge = '';
        if(status === 'Success') {
            statusBadge = `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">Success</span>`;
        } else {
            statusBadge = `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">${status}</span>`;
        }

        let typeIcon = type === 'Fare' 
            ? `<span class="w-2 h-2 rounded-full bg-blue-500 mr-2"></span>` 
            : `<span class="w-2 h-2 rounded-full bg-purple-500 mr-2"></span>`;

        row.innerHTML = `
            <td class="px-6 py-4 text-slate-500">${timeStr}</td>
            <td class="px-6 py-4 text-slate-900 font-medium flex items-center">${typeIcon} ${type}</td>
            <td class="px-6 py-4 text-slate-900 font-semibold text-right">₱${amount.toFixed(2)}</td>
            <td class="px-6 py-4 text-slate-500 text-right">₱${fee.toFixed(2)}</td>
            <td class="px-6 py-4 text-center">${statusBadge}</td>
        `;

        activityTable.prepend(row); // Add to top
    }
});
