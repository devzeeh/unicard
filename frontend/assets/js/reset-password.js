document.addEventListener("DOMContentLoaded", function () {
    // Find all the elements from the reset form
    const resetForm = document.getElementById('reset-form');
    const passwordInput = document.getElementById('new_password'); 
    const confirmPasswordInput = document.getElementById('confirm_password');
    const submitButton = document.getElementById('reset-submit-btn');
    const errorMessage = document.getElementById('error-message');

    // Validation checklist items
    const checklist = document.getElementById('validation-checklist');
    const lengthCheck = document.getElementById('length-check');
    const caseCheck = document.getElementById('case-check');
    const numCheck = document.getElementById('num-check');
    const matchCheck = document.getElementById('match-check');

    // Regex for validation
    const hasLower = new RegExp(/[a-z]/);
    const hasUpper = new RegExp(/[A-Z]/);
    const hasNumber = new RegExp(/[0-9]/);

    // Helper function to update the checklist icons and colors
    function updateChecklistItem(checkElement, isValid) {
        // Stop if the element doesn't exist
        if (!checkElement) return; 
        
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

    // Main validation function that runs on every input
    function validateForm() {
        const password = passwordInput.value;
        const confirmPassword = confirmPasswordInput.value;

        // Show the checklist as soon as the user starts typing in either password field
        if (password.length > 0 || confirmPassword.length > 0) {
            checklist.classList.remove('hidden');
        } else {
            checklist.classList.add('hidden');
        }

        // 1. Check password complexity
        const isLengthValid = password.length >= 8;
        const isCaseValid = hasLower.test(password) && hasUpper.test(password);
        const isNumValid = hasNumber.test(password);

        // 2. Check if passwords match
        // Passwords only match if they are not empty and are equal
        const passwordsMatch = password === confirmPassword && password.length > 0;

        // 3. Update the checklist UI
        updateChecklistItem(lengthCheck, isLengthValid);
        updateChecklistItem(caseCheck, isCaseValid);
        updateChecklistItem(numCheck, isNumValid);
        updateChecklistItem(matchCheck, passwordsMatch);
        
        // 4. Enable or disable the submit button
        const allValid = isLengthValid && isCaseValid && isNumValid && passwordsMatch;

        if (allValid) {
            submitButton.disabled = false;
            errorMessage.classList.add('hidden');
        } else {
            submitButton.disabled = true;
        }
    }

    // Add event listeners to both password fields
    if (passwordInput && confirmPasswordInput) {
        passwordInput.addEventListener('input', validateForm);
        confirmPasswordInput.addEventListener('input', validateForm);
    }

    // Handle form submission
    if (resetForm) {
        resetForm.addEventListener('submit', function (event) {
            event.preventDefault(); // Stop default form submission

            // Double-check validation before "submitting"
            validateForm(); 

            if (submitButton.disabled) {
                // Updated error message
                errorMessage.textContent = 'Please fix the errors in the password checklist.';
                errorMessage.classList.remove('hidden');
            } else {
                // --- THIS IS A FRONTEND-ONLY DEMO ---
                // In a real app, you would make a fetch() call to your backend
                // to set the new password.
                errorMessage.classList.add('hidden');
                alert('Password reset successfully! Redirecting to login...');
                
                // Redirect to login page after success
                window.location.href = "paycard_login.html";
            }
        });
    }
});