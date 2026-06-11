document.addEventListener("DOMContentLoaded", function () {
    const quickBtns = document.querySelectorAll('.quick-amount-btn');
    const amountInput = document.getElementById('amount');

    if (quickBtns && amountInput) {
        quickBtns.forEach(btn => {
            btn.addEventListener('click', () => {
                amountInput.value = btn.innerText;
            });
        });
    }

    const form = document.getElementById('topup-form');
    const submitBtn = document.getElementById('topup-submit-btn');
    const errorDiv = document.getElementById('amount-error');

    // Function to run validation
    const validateForm = () => {
        const amountText = amountInput.value.trim();
        const amount = parseFloat(amountText);
        const method = document.querySelector('input[name="payment_method"]:checked');
        
        let isValid = true;

        if (amountText === '') {
            isValid = false;
            if (errorDiv) errorDiv.classList.add('hidden');
        } else if (amountText.startsWith('-')) {
            isValid = false;
            if (errorDiv) {
                errorDiv.textContent = 'Negative values are not allowed.';
                errorDiv.classList.remove('hidden');
            }
        } else if (isNaN(amount) || amount < 50) {
            isValid = false;
            if (errorDiv) {
                errorDiv.textContent = 'Minimum top-up amount is ₱50.00.';
                errorDiv.classList.remove('hidden');
            }
        } else {
            if (errorDiv) errorDiv.classList.add('hidden');
        }
        
        submitBtn.disabled = !(isValid && method);
    };

    if (form && submitBtn) {
        form.addEventListener('input', validateForm);
    }
    
    // Also attach to quick amount buttons so clicking them triggers validation
    if (quickBtns && amountInput) {
        quickBtns.forEach(btn => {
            btn.addEventListener('click', () => {
                amountInput.value = btn.innerText;
                validateForm();
            });
        });
    }
});
