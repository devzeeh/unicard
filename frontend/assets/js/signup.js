document.addEventListener("DOMContentLoaded", function () {
    // STEP ELEMENTS 
    const step1 = document.getElementById('step-1');
    const step2 = document.getElementById('step-2');
    const step3 = document.getElementById('step-3');
    const stepSubtitle = document.getElementById('step-subtitle');
    const stepProgress = document.getElementById('step-progress');

    // BUTTONS 
    const btnStep1 = document.getElementById('btn-step-1');
    const btnBack2 = document.getElementById('btn-back-2');
    const btnStep2 = document.getElementById('btn-step-2');
    const btnBack3 = document.getElementById('btn-back-3');
    const createAccountBtn = document.getElementById('create-account-btn');
    const signupForm = document.getElementById('signup-form');

    // INPUT FIELDS 
    const firstNameInput = document.getElementById('first_name');
    const lastNameInput = document.getElementById('last_name');
    const emailInput = document.getElementById('email');
    const contactNumberInput = document.getElementById('contact_number');
    const cardNumberInput = document.getElementById('card_number');
    const passwordInput = document.getElementById('password');
    const confirmPasswordInput = document.getElementById('confirm_password');

    // ERROR DISPLAYS 
    const errorMessage = document.getElementById('error-message');
    const cardNumberError = document.getElementById('card-number-error');

    // MODAL
    const successModal = document.getElementById('success-modal');
    const modalCloseBtn = document.getElementById('modal-close-btn');

    // STATE 
    let formData = {
        firstName: '',
        lastName: '',
        email: '',
        contactNumber: '',
        cardNumber: '',
        password: ''
    };
    
    // HELPER FUNCTIONS 
    function showStep(stepNumber) {
        [step1, step2, step3].forEach(s => s?.classList.add('hidden'));
        errorMessage.classList.add('hidden'); 

        if (stepNumber === 1) {
            step1.classList.remove('hidden');
            stepSubtitle.textContent = 'Step 1 of 3: Your Details';
            if (stepProgress) stepProgress.style.width = '33.33%';
        } else if (stepNumber === 2) {
            step2.classList.remove('hidden');
            stepSubtitle.textContent = 'Step 2 of 3: Card Verification';
            if (stepProgress) stepProgress.style.width = '66.66%';
        } else if (stepNumber === 3) {
            step3.classList.remove('hidden');
            stepSubtitle.textContent = 'Step 3 of 3: Create Password';
            if (stepProgress) stepProgress.style.width = '100%';
        }
    }

    function showError(msg) {
        errorMessage.textContent = msg;
        errorMessage.classList.remove('hidden');
        window.scrollTo({ top: 0, behavior: 'smooth' });
    }

    // VALIDATION FUNCTIONS 
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

    async function validateStep2() {
        if (!validateInput(cardNumberInput)) return false;

        try {
            btnStep2.disabled = true;
            btnStep2.textContent = 'Verifying...';
            
            const response = await fetch('/v1/signup/check-card', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ card_number: cardNumberInput.value.trim() })
            });
            
            const data = await response.json();
            if (!data.success) {
                showError(data.message);
                cardNumberInput.classList.add('border-red-500', 'focus:border-red-500', 'focus:ring-red-500');
                cardNumberInput.classList.remove('border-gray-300', 'focus:border-blue-500', 'focus:ring-blue-500');
                return false;
            }

            formData.cardNumber = cardNumberInput.value.trim();
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
        const isMatchValid = password === confirmPassword && confirmPassword !== '';

        const lengthCheck = document.getElementById('length-check');
        const matchCheck = document.getElementById('match-check');

        const successSvg = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="w-5 h-5 text-green-500"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>`;
        const errorSvg = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="w-5 h-5 text-red-500"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>`;

        lengthCheck.querySelector('svg').outerHTML = isLengthValid ? successSvg : errorSvg;
        matchCheck.querySelector('svg').outerHTML = isMatchValid ? successSvg : errorSvg;

        lengthCheck.className = `flex items-center space-x-2 text-sm ${isLengthValid ? 'text-green-600' : 'text-gray-600'}`;
        matchCheck.className = `flex items-center space-x-2 text-sm ${isMatchValid ? 'text-green-600' : 'text-gray-600'}`;

        return isLengthValid && isMatchValid;
    }

    passwordInput.addEventListener('input', updatePasswordChecklist);
    confirmPasswordInput.addEventListener('input', updatePasswordChecklist);

    function validateStep3() {
        const inputs = [passwordInput, confirmPasswordInput];
        let isValid = true;
        inputs.forEach(input => {
            if (!validateInput(input)) isValid = false;
        });

        if (!updatePasswordChecklist()) return false;
        if (!isValid) return false;

        formData.password = passwordInput.value;
        return true;
    }

    // EVENT LISTENERS 
    btnStep1.addEventListener('click', async () => {
        if (await validateStep1()) showStep(2);
    });

    btnBack2.addEventListener('click', () => showStep(1));
    btnStep2.addEventListener('click', async () => {
        if (await validateStep2()) showStep(3);
    });

    btnBack3.addEventListener('click', () => showStep(2));

    signupForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        if (!validateStep3()) return;

        try {
            createAccountBtn.disabled = true;
            createAccountBtn.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="mr-2 w-5 h-5 animate-spin"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg> Creating Account...';
            
            const response = await fetch('/v1/signupauth', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    first_name: formData.firstName,
                    last_name: formData.lastName,
                    email: formData.email,
                    contact_number: formData.contactNumber,
                    card_number: formData.cardNumber,
                    password: formData.password
                })
            });

            const data = await response.json();

            if (data.success) {
                successModal.classList.remove('hidden');
            } else {
                showError(data.message || 'Registration failed. Please try again.');
            }
        } catch (error) {
            showError('A network error occurred. Please try again.');
        } finally {
            createAccountBtn.disabled = false;
            createAccountBtn.innerHTML = 'Create Account';
        }
    });

    modalCloseBtn.addEventListener('click', () => {
        window.location.href = '/login';
    });

    // Initialize formatting functions
    window.validateInput = function(input) {
        const errorElement = document.getElementById(`error-${input.id}`);
        let isValid = true;
        let errorMsg = '';

        input.classList.remove('border-red-500', 'focus:border-red-500', 'focus:ring-red-500', 'border-gray-300', 'focus:border-blue-500', 'focus:ring-blue-500');
        if(errorElement) errorElement.classList.add('hidden');

        if (!input.value.trim() && input.hasAttribute('required')) {
            isValid = false;
            errorMsg = 'This field is required';
        } else if (input.hasAttribute('pattern')) {
            const regex = new RegExp(`^${input.getAttribute('pattern')}$`);
            if (!regex.test(input.value)) {
                isValid = false;
                errorMsg = 'Invalid format';
            }
        }

        if (!isValid) {
            input.classList.add('border-red-500', 'focus:border-red-500', 'focus:ring-red-500');
            if(errorElement) {
                errorElement.textContent = errorMsg;
                errorElement.classList.remove('hidden');
            }
        } else {
            input.classList.add('border-gray-300', 'focus:border-blue-500', 'focus:ring-blue-500');
        }

        return isValid;
    };

    window.isNumber = function(evt) {
        evt = (evt) ? evt : window.event;
        var charCode = (evt.which) ? evt.which : evt.keyCode;
        if (charCode > 31 && (charCode < 48 || charCode > 57)) {
            return false;
        }
        return true;
    };

    window.isAlpha = function(evt) {
        evt = (evt) ? evt : window.event;
        var charCode = (evt.which) ? evt.which : evt.keyCode;
        if ((charCode >= 65 && charCode <= 90) || 
            (charCode >= 97 && charCode <= 122) || 
            charCode == 32) {
            return true;
        }
        return false;
    };

    showStep(1);
});