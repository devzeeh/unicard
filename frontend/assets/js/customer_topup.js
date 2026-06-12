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
    
    // Breakdown elements
    const breakdownContainer = document.getElementById('breakdown-container');
    const breakdownAmount = document.getElementById('breakdown-amount');
    const breakdownTotal = document.getElementById('breakdown-total');

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
        
        // Update breakdown
        if (isValid) {
            breakdownAmount.textContent = amount.toFixed(2);
            breakdownTotal.textContent = (amount + 15).toFixed(2);
            breakdownContainer.classList.remove('hidden');
        } else {
            breakdownContainer.classList.add('hidden');
        }

        submitBtn.disabled = !(isValid && method);
    };

    if (form && submitBtn) {
        form.addEventListener('input', validateForm);

        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const amountText = amountInput.value.trim();
            const amount = parseFloat(amountText);
            const method = document.querySelector('input[name="payment_method"]:checked');

            if (!method || method.value !== 'xendit' || isNaN(amount) || amount < 50) {
                return;
            }

            // Get username from URL path: /u/{username}/topup
            const pathParts = window.location.pathname.split('/');
            let username = '';
            if (pathParts.length >= 3 && pathParts[1] === 'u') {
                username = pathParts[2];
            }

            if (!username) {
                if (errorDiv) {
                    errorDiv.textContent = 'Could not determine user profile from URL.';
                    errorDiv.classList.remove('hidden');
                }
                return;
            }

            try {
                submitBtn.disabled = true;
                submitBtn.innerText = 'Processing...';

                const response = await fetch(`/v1/user/${username}/topup/checkout`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ amount })
                });

                if (!response.ok) {
                    const errorText = await response.text();
                    throw new Error(errorText || 'Failed to initialize checkout');
                }

                const data = await response.json();
                if (data.url) {
                    window.location.href = data.url;
                } else {
                    throw new Error('No checkout URL returned');
                }
            } catch (err) {
                console.error('Checkout error:', err);
                if (errorDiv) {
                    errorDiv.textContent = err.message;
                    errorDiv.classList.remove('hidden');
                }
                submitBtn.disabled = false;
                submitBtn.innerText = 'Top Up Now';
            }
        });
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
