// Global helper functions for input restriction
function isValidEmail(email) {
    const regex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    return regex.test(email);
}

function isNumber(evt) {
    evt = (evt) ? evt : window.event;
    var charCode = (evt.which) ? evt.which : evt.keyCode;
    // Allow only numbers
    if (charCode > 31 && (charCode < 48 || charCode > 57)) {
        return false;
    }
    return true;
}

function isAlpha(evt) {
    evt = (evt) ? evt : window.event;
    var charCode = (evt.which) ? evt.which : evt.keyCode;
    // Allow letters and spaces. Also allow control characters < 32.
    if (charCode < 32) return true;
    if ((charCode >= 65 && charCode <= 90) || 
        (charCode >= 97 && charCode <= 122) || 
        charCode === 32) {
        return true;
    }
    return false;
}

document.addEventListener("DOMContentLoaded", function () {
    // STEP ELEMENTS 
    const step1 = document.getElementById('step-1');
    const step2 = document.getElementById('step-2');
    const step3 = document.getElementById('step-3');
    const stepSubtitle = document.getElementById('step-subtitle');

    // BUTTONS 
    const btnStep1 = document.getElementById('btn-step-1');
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
    
    // INPUTS 
    const cardIdInput = document.getElementById('card_id');
    const cardIdError = document.getElementById('card-id-error');
    
    // INPUTS 
    const passwordInput = document.getElementById('password');
    const confirmPasswordInput = document.getElementById('confirm_password');
    const checklist = document.getElementById('validation-checklist');
    const lengthCheck = document.getElementById('length-check');
    const matchCheck = document.getElementById('match-check');
    
    // MODAL (ADDED) 
    const successModal = document.getElementById('success-modal');
    const modalCloseBtn = document.getElementById('modal-close-btn');

    // GLOBAL 
    const errorMessage = document.getElementById('error-message');

    // FORM DATA STORAGE 
    const formData = {
        firstName: '',
        lastName: '',
        email: '',
        cardId: '',
        password: '',
    };
    
    //  ROBUSTNESS CHECK 
    if (!step1 || !step2 || !step3 || !stepSubtitle || !btnStep1 || !btnBack2 || !btnStep2 || !btnBack3 || !createAccountBtn || !signupForm || !firstNameInput || !emailInput || !cardIdInput || !cardIdError || !passwordInput || !confirmPasswordInput || !checklist || !lengthCheck || !matchCheck || !errorMessage || !successModal || !modalCloseBtn) {
        console.error("Signup Script Error: Not all required HTML elements were found on the page.");
        return; // Stop the script
    }

    // INITIALIZATION 

    // HELPER FUNCTIONS 
    function showStep(stepNumber) {
        step1.classList.add('hidden');
        step2.classList.add('hidden');
        step3.classList.add('hidden');
        errorMessage.classList.add('hidden'); 

        if (stepNumber === 1) {
            step1.classList.remove('hidden');
            stepSubtitle.textContent = 'Step 1 of 3: Your Details';
        } else if (stepNumber === 2) {
            step2.classList.remove('hidden');
            stepSubtitle.textContent = 'Step 2 of 3: Card Verification';
        } else if (stepNumber === 3) {
            step3.classList.remove('hidden');
            stepSubtitle.textContent = 'Step 3 of 3: Create Password';
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
    }

    // VALIDATION LOGIC 

    // Real-time validation for Name and Email fields
    function validateStep1Realtime() {
        const firstName = firstNameInput.value.trim();
        const lastName = lastNameInput.value.trim();
        const email = emailInput.value.trim();
        const contactNumber = contactNumberInput.value.trim();
        
        const isNameValid = firstName !== '' && lastName !== '';
        const isEmailValid = email !== '' && isValidEmail(email);
        const isContactValid = contactNumber !== '';

        if (isNameValid && isEmailValid && isContactValid) {
            errorMessage.classList.add('hidden');
        }
    }

    // Click validation (shows error)
    function validateStep1() {
        const firstName = firstNameInput.value.trim();
        const lastName = lastNameInput.value.trim();
        const email = emailInput.value.trim();
        const contactNumber = contactNumberInput.value.trim();
        
        if (firstName === '' || lastName === '' || email === '' || !isValidEmail(email) || contactNumber === '') {
            showError('Please fill all fields and provide a valid email address.');
            return false;
        }
        
        // TODO: Backend check for email
        formData.firstName = firstName;
        formData.lastName = lastName;
        formData.email = email;
        formData.contactNumber = contactNumber;
        return true;
    }

    // Validate Card ID (real-time and click)
    function validateStep2() {
        const cardId = cardIdInput.value.trim();

        if (cardId === "") {
            cardIdError.textContent = 'Please enter your Card ID.';
            cardIdError.classList.remove('hidden');
            return false;
        }

        const isCardIdOnlyNumbers = /^\d+$/.test(cardId);
        const isCardIdValidLength = cardId.length === 10;
        
        if (!isCardIdOnlyNumbers) {
            cardIdError.textContent = 'Card ID must contain only numbers.';
            cardIdError.classList.remove('hidden');
            return false;
        } else if (!isCardIdValidLength) {
            cardIdError.textContent = 'Card ID must be exactly 10 digits long.';
            cardIdError.classList.remove('hidden');
            return false;
        }

        // TODO: Backend check for card ID
        
        cardIdError.classList.add('hidden');
        formData.cardId = cardId;
        return true;
    }

    // Validate Password
    function validateStep3() {
        const password = passwordInput.value;
        const confirmPassword = confirmPasswordInput.value;

        const isLengthValid = password.length >= 8;
        const passwordsMatch = password === confirmPassword && password.length > 0;

        updateChecklistItem(lengthCheck, isLengthValid);
        updateChecklistItem(matchCheck, passwordsMatch);
        
        const allValid = isLengthValid && passwordsMatch;

        if (allValid) {
            formData.password = password;
            return true;
        }
        return false;
    }

    // EVENT LISTENERS 

    // Add real-time listeners for Step 1 
    firstNameInput.addEventListener('input', validateStep1Realtime);
    lastNameInput.addEventListener('input', validateStep1Realtime);
    emailInput.addEventListener('input', validateStep1Realtime);
    contactNumberInput.addEventListener('input', validateStep1Realtime);

    // Next from Step 1
    btnStep1.addEventListener('click', function () {
        if (validateStep1()) {
            showStep(2);
        }
    });

    // Back from Step 2
    btnBack2.addEventListener('click', function () {
        showStep(1);
    });

    // Next from Step 2
    btnStep2.addEventListener('click', function () {
        if (validateStep2()) {
            showStep(3);
        }
    });
    
    // Real-time validation for Card ID as user types
    cardIdInput.addEventListener('input', validateStep2);

    // Back from Step 3
    btnBack3.addEventListener('click', function () {
        showStep(2);
    });

    // Real-time validation for Password fields
    passwordInput.addEventListener('input', validateStep3);
    confirmPasswordInput.addEventListener('input', validateStep3);

    // Final Form Submission
    signupForm.addEventListener('submit', function (event) {
        event.preventDefault(); 

        if (validateStep3()) {
            fetch("/v1/signupauth", {
                method: "POST",
                headers: {"Content-Type": "application/json"},
                body: JSON.stringify({
                    firstName: formData.firstName,
                    lastName: formData.lastName,
                    email: formData.email,
                    contactNumber: formData.contactNumber,
                    cardNumber: formData.cardId,
                    password: formData.password
                })
            })
            .then(res => res.json())
            .then(data => {
                if (data.success) {
                    successModal.classList.remove('hidden');
                } else {
                    showError(data.message || 'Failed to create account');
                }
            })
            .catch(err => {
                console.error(err);
                showError('Network error occurred.');
            });
        } else {
            showError('Please correct the errors in the password fields.');
        }
    });

    // MODAL BUTTON (ADDED) 
    // Add event listener for the modal's "Go to Login" button
    modalCloseBtn.addEventListener('click', function() {
        window.location.href = "login.html";
    });
});