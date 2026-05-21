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

    if (form && submitBtn) {
        form.addEventListener('input', () => {
            const amount = parseFloat(amountInput.value);
            const method = document.querySelector('input[name="payment_method"]:checked');
            
            if (amount > 0 && method) {
                submitBtn.disabled = false;
            } else {
                submitBtn.disabled = true;
            }
        });
    }
});
