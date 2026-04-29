function isNumber(evt) {
    evt = (evt) ? evt : window.event;
    var charCode = (evt.which) ? evt.which : evt.keyCode;
    if (charCode > 31 && (charCode < 48 || charCode > 57)) return false;
    return true;
}

function isAlpha(evt) {
    evt = (evt) ? evt : window.event;
    var charCode = (evt.which) ? evt.which : evt.keyCode;
    if (charCode < 32) return true;
    if ((charCode >= 65 && charCode <= 90) || 
        (charCode >= 97 && charCode <= 122) || 
        charCode === 32) return true;
    return false;
}

function validateInput(el) {
    const errorEl = document.getElementById(`error-${el.id}`);
    if (!el.checkValidity() || el.value.trim() === "") {
        el.classList.add('border-red-500', 'focus:border-red-500', 'focus:ring-red-500');
        el.classList.remove('border-gray-300', 'focus:border-blue-500', 'focus:ring-blue-500');
        if (errorEl) {
            if (el.value.trim() === "") {
                errorEl.textContent = "This field is required.";
            } else if (el.type === "email") {
                errorEl.textContent = "Please enter a valid email address.";
            } else if (el.id === "contact_number") {
                errorEl.textContent = "Please enter a valid 11-digit contact number.";
            } else if (el.id === "card_id") {
                errorEl.textContent = "Please enter a valid 16-digit card number.";
            } else {
                errorEl.textContent = "Invalid format.";
            }
            errorEl.classList.remove('hidden');
        }
        return false;
    } else {
        el.classList.remove('border-red-500', 'focus:border-red-500', 'focus:ring-red-500');
        el.classList.add('border-gray-300', 'focus:border-blue-500', 'focus:ring-blue-500');
        if (errorEl) errorEl.classList.add('hidden');
        return true;
    }
}

document.addEventListener("DOMContentLoaded", function () {
    // STEP ELEMENTS 
    const step1 = document.getElementById('step-1');
    const stepOTP = document.getElementById('step-otp');
    const step2 = document.getElementById('step-2');
    const step3 = document.getElementById('step-3');
    const stepSubtitle = document.getElementById('step-subtitle');
    const stepProgress = document.getElementById('step-progress');

    // BUTTONS 
    const btnStep1 = document.getElementById('btn-step-1');
    const btnBackOTP = document.getElementById('btn-back-otp');
    const btnStepOTP = document.getElementById('btn-step-otp');
    const btnBack2 = document.getElementById('btn-back-2');
    const btnStep2 = document.getElementById('btn-step-2');
    const btnBack3 = document.getElementById('btn-back-3');
    const createAccountBtn = document.getElementById('create-account-btn');
    const signupForm = document.getElementById('signup-form');

    // INPUTS 
    const firstNameInput = document.getElementById('first_name');
    const lastNameInput = document.getElementById('last_name');
    const emailInput = document.getElementById('email');
    const contactNumberInput = document.getElementById('contact_number');
    const otpInput = document.getElementById('otp');
    const cardIdInput = document.getElementById('card_id');
    const passwordInput = document.getElementById('password');
    const confirmPasswordInput = document.getElementById('confirm_password');

    // FEEDBACK 
    const cardIdError = document.getElementById('card-id-error');
    const errorMessage = document.getElementById('error-message');
    const checklist = document.getElementById('validation-checklist');
    const lengthCheck = document.getElementById('length-check');
    const matchCheck = document.getElementById('match-check');
    
    // MODAL 
    const successModal = document.getElementById('success-modal');
    const modalCloseBtn = document.getElementById('modal-close-btn');

    // FORM DATA STORAGE 
    const formData = {
        firstName: '',
        lastName: '',
        email: '',
        contactNumber: '',
        cardId: '',
        password: '',
    };
    
    // HELPER FUNCTIONS 
    function showStep(stepNumber) {
        [step1, stepOTP, step2, step3].forEach(s => s?.classList.add('hidden'));
        errorMessage.classList.add('hidden'); 

        if (stepNumber === 1) {
            step1.classList.remove('hidden');
            stepSubtitle.textContent = 'Step 1 of 4: Your Details';
            if (stepProgress) stepProgress.style.width = '25%';
        } else if (stepNumber === 'otp') {
            stepOTP.classList.remove('hidden');
            stepSubtitle.textContent = 'Step 2 of 4: Verify Email';
            if (stepProgress) stepProgress.style.width = '50%';
        } else if (stepNumber === 2) {
            step2.classList.remove('hidden');
            stepSubtitle.textContent = 'Step 3 of 4: Card Verification';
            if (stepProgress) stepProgress.style.width = '75%';
        } else if (stepNumber === 3) {
            step3.classList.remove('hidden');
            stepSubtitle.textContent = 'Step 4 of 4: Create Password';
            if (stepProgress) stepProgress.style.width = '100%';
        }
    }

    function updateChecklistItem(checkElement, isValid) {
        const icon = checkElement.querySelector('i');
        if (isValid) {
            checkElement.classList.add('valid');
            icon.classList.remove('fa-times-circle', 'text-red-500');
            icon.classList.add('fa-check-circle', 'text-green-600');
        } else {
            checkElement.classList.remove('valid');
            icon.classList.remove('fa-check-circle', 'text-green-600');
            icon.classList.add('fa-times-circle', 'text-red-500');
        }
    }

    function showError(message) {
        errorMessage.textContent = message;
        errorMessage.classList.remove('hidden');
        errorMessage.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }

    // VALIDATION & API CALLS 

    async function validateStep1() {
        const inputs = [firstNameInput, lastNameInput, emailInput, contactNumberInput];
        let isValid = true;
        inputs.forEach(input => {
            if (!validateInput(input)) isValid = false;
        });

        if (!isValid) return false;

        // Backend check for Email/Phone availability
         try {
            btnStep1.disabled = true;
            btnStep1.textContent = 'Checking...';
            
            const response = await fetch('/v1/signup/check-details', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    email: emailInput.value.trim(),
                    contact_number: contactNumberInput.value.trim()
                })
            });
            
            const data = await response.json();
            if (!data.success) {
                showError(data.message);
                if (data.field === 'email') {
                    emailInput.classList.add('border-red-500', 'focus:border-red-500', 'focus:ring-red-500');
                    emailInput.classList.remove('border-gray-300', 'focus:border-blue-500', 'focus:ring-blue-500');
                }
                if (data.field === 'phone') {
                    contactNumberInput.classList.add('border-red-500', 'focus:border-red-500', 'focus:ring-red-500');
                    contactNumberInput.classList.remove('border-gray-300', 'focus:border-blue-500', 'focus:ring-blue-500');
                }
                return false;
            }

            formData.firstName = firstNameInput.value.trim();
            formData.lastName = lastNameInput.value.trim();
            formData.email = emailInput.value.trim();
            formData.contactNumber = contactNumberInput.value.trim();
            return true;
        } catch (err) {
            showError('Network error. Please try again.');
            return false;
        } finally {
            btnStep1.disabled = false;
            btnStep1.textContent = 'Next';
        }
    }

    async function validateOTP() {
        if (!validateInput(otpInput)) return false;

        try {
            btnStepOTP.disabled = true;
            btnStepOTP.textContent = 'Verifying...';
            
            const response = await fetch('/v1/signup/verify-otp', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    email: formData.email,
                    otp: otpInput.value.trim()
                })
            });
            
            const data = await response.json();
            if (!data.success) {
                showError(data.message);
                otpInput.classList.add('border-red-500', 'focus:border-red-500', 'focus:ring-red-500');
                otpInput.classList.remove('border-gray-300', 'focus:border-blue-500', 'focus:ring-blue-500');
                return false;
            }
            return true;
        } catch (err) {
            showError('Network error. Please try again.');
            return false;
        } finally {
            btnStepOTP.disabled = false;
            btnStepOTP.textContent = 'Verify';
        }
    }

    async function validateStep2() {
        if (!validateInput(cardIdInput)) return false;

        try {
            btnStep2.disabled = true;
            btnStep2.textContent = 'Verifying...';
            
            const response = await fetch('/v1/signup/check-card', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ card_number: cardIdInput.value.trim() })
            });
            
            const data = await response.json();
            if (!data.success) {
                showError(data.message);
                cardIdInput.classList.add('border-red-500', 'focus:border-red-500', 'focus:ring-red-500');
                cardIdInput.classList.remove('border-gray-300', 'focus:border-blue-500', 'focus:ring-blue-500');
                return false;
            }

            formData.cardId = cardIdInput.value.trim();
            return true;
        } catch (err) {
            showError('Network error. Please try again.');
            return false;
        } finally {
            btnStep2.disabled = false;
            btnStep2.textContent = 'Next';
        }
    }

    function updatePasswordChecklist() {
        const password = passwordInput.value;
        const confirmPassword = confirmPasswordInput.value;

        const isLengthValid = password.length >= 8;
        const passwordsMatch = password === confirmPassword && password.length > 0;

        updateChecklistItem(lengthCheck, isLengthValid);
        updateChecklistItem(matchCheck, passwordsMatch);
    }

    function validateStep3() {
        const p1 = validateInput(passwordInput);
        const p2 = validateInput(confirmPasswordInput);
        
        updatePasswordChecklist();

        const password = passwordInput.value;
        const confirmPassword = confirmPasswordInput.value;

        const isLengthValid = password.length >= 8;
        const passwordsMatch = password === confirmPassword && password.length > 0;

        if (p1 && p2 && isLengthValid && passwordsMatch) {
            formData.password = password;
            return true;
        }
        return false;
    }

    // EVENT LISTENERS 
    btnStep1.addEventListener('click', async () => {
        if (await validateStep1()) showStep('otp');
    });

    btnBackOTP.addEventListener('click', () => showStep(1));
    btnStepOTP.addEventListener('click', async () => {
        if (await validateOTP()) showStep(2);
    });

    btnBack2.addEventListener('click', () => showStep('otp'));
    btnStep2.addEventListener('click', async () => {
        if (await validateStep2()) showStep(3);
    });

    btnBack3.addEventListener('click', () => showStep(2));

    passwordInput.addEventListener('input', updatePasswordChecklist);
    confirmPasswordInput.addEventListener('input', updatePasswordChecklist);

    signupForm.addEventListener('submit', async (event) => {
        event.preventDefault(); 
        if (!validateStep3()) return;

        try {
            createAccountBtn.disabled = true;
            createAccountBtn.textContent = 'Creating...';

            const response = await fetch("/v1/signupauth", {
                method: "POST",
                headers: {"Content-Type": "application/json"},
                body: JSON.stringify({
                    first_name: formData.firstName,
                    last_name: formData.lastName,
                    email: formData.email,
                    contact_number: formData.contactNumber,
                    card_number: formData.cardId,
                    password: formData.password
                })
            });
            
            const data = await response.json();
            if (data.success) {
                successModal.classList.remove('hidden');
            } else {
                showError(data.message || 'Failed to create account');
            }
        } catch (err) {
            showError('Network error occurred.');
        } finally {
            createAccountBtn.disabled = false;
            createAccountBtn.textContent = 'Create Account';
        }
    });

    modalCloseBtn.addEventListener('click', () => window.location.href = "/login");
});