document.addEventListener("DOMContentLoaded", function () {
    console.log("Profile page script loaded.");

    // --- Profile Edit Elements ---
    const editProfileBtn = document.getElementById('edit-profile-btn');
    const profileView = document.getElementById('profile-details-view');
    const profileEditForm = document.getElementById('profile-details-edit');

    if (editProfileBtn && profileView && profileEditForm) {
        editProfileBtn.addEventListener('click', () => {
            const isEditing = editProfileBtn.textContent.trim() === 'Save';

            if (isEditing) {
                // --- THIS IS A FRONTEND-ONLY DEMO ---
                // In a real app, you would send a fetch() request to your backend
                // to save the new details from the form.
                
                // For now, we'll just update the text and switch back
                const newName = document.getElementById('full_name').value;
                const newEmail = document.getElementById('email').value;
                const newPhone = document.getElementById('phone').value;

                profileView.querySelector('dd:nth-of-type(1)').textContent = newName;
                profileView.querySelector('dd:nth-of-type(2)').textContent = newEmail;
                profileView.querySelector('dd:nth-of-type(3)').textContent = newPhone;
                
                profileView.classList.remove('hidden');
                profileEditForm.classList.add('hidden');
                editProfileBtn.innerHTML = '<i class="fas fa-pencil-alt mr-1"></i> <span>Edit</span>';
            } else {
                // Switch to edit mode
                profileView.classList.add('hidden');
                profileEditForm.classList.remove('hidden');
                editProfileBtn.innerHTML = '<i class="fas fa-save mr-1"></i> <span>Save</span>';
            }
        });
    }

    // --- Change Password Elements ---
    const passwordForm = document.getElementById('change-password-form');
    const newPasswordInput = document.getElementById('new_password');
    const confirmPasswordInput = document.getElementById('confirm_password');
    const passwordSubmitBtn = document.getElementById('change-password-btn');
    const passwordErrorMsg = document.getElementById('password-error-message');
    
    // Checklist elements
    const checklist = document.getElementById('password-checklist');
    const lengthCheck = document.getElementById('length-check');
    const caseCheck = document.getElementById('case-check');
    const numCheck = document.getElementById('num-check');
    const matchCheck = document.getElementById('match-check');

    // Regex for validation
    const hasLower = new RegExp(/[a-z]/);
    const hasUpper = new RegExp(/[A-Z]/);
    const hasNumber = new RegExp(/[0-9]/);

    // Helper to update checklist items
    function updateChecklistItem(checkElement, isValid) {
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

    // Main validation function
    function validatePasswordForm() {
        const password = newPasswordInput.value;
        const confirmPassword = confirmPasswordInput.value;

        if (password.length > 0 || confirmPassword.length > 0) {
            checklist.classList.remove('hidden');
        } else {
            checklist.classList.add('hidden');
        }

        const isLengthValid = password.length >= 8;
        const isCaseValid = hasLower.test(password) && hasUpper.test(password);
        const isNumValid = hasNumber.test(password);
        const passwordsMatch = password === confirmPassword && password.length > 0;

        updateChecklistItem(lengthCheck, isLengthValid);
        updateChecklistItem(caseCheck, isCaseValid);
        updateChecklistItem(numCheck, isNumValid);
        updateChecklistItem(matchCheck, passwordsMatch);

        const allValid = isLengthValid && isCaseValid && isNumValid && passwordsMatch;
        passwordSubmitBtn.disabled = !allValid;
    }

    if (passwordForm) {
        newPasswordInput.addEventListener('input', validatePasswordForm);
        confirmPasswordInput.addEventListener('input', validatePasswordForm);

        passwordForm.addEventListener('submit', (e) => {
            e.preventDefault();
            if (passwordSubmitBtn.disabled) {
                passwordErrorMsg.textContent = 'Please fix the errors in the password checklist.';
                passwordErrorMsg.classList.remove('hidden');
            } else {
                // --- FRONTEND-ONLY DEMO ---
                // In a real app, you'd send a fetch() request to your backend
                // to verify the *current* password and set the new one.
                passwordErrorMsg.classList.add('hidden');
                alert('Password changed successfully!');
                passwordForm.reset();
                checklist.classList.add('hidden');
                passwordSubmitBtn.disabled = true;
            }
        });
    }

    // --- Delete Account Modal Elements ---
    const deleteAccountBtn = document.getElementById('delete-account-btn');
    const deleteModal = document.getElementById('delete-account-modal');
    const deleteModalContent = document.getElementById('delete-modal-content');
    const deleteModalCloseBtn = document.getElementById('delete-modal-close-button');
    const deleteModalCancelBtn = document.getElementById('delete-modal-cancel-button');
    const deleteModalConfirmBtn = document.getElementById('delete-modal-confirm-button');
    const deleteConfirmText = document.getElementById('delete-confirm-text');

    if (deleteAccountBtn && deleteModal && deleteModalContent && deleteModalCloseBtn && deleteModalCancelBtn && deleteModalConfirmBtn && deleteConfirmText) {
        
        function openDeleteModal() {
            deleteModal.classList.remove('hidden');
            setTimeout(() => {
                deleteModal.classList.add('opacity-100');
                deleteModalContent.classList.add('scale-100', 'opacity-100');
                deleteModalContent.classList.remove('scale-95', 'opacity-0');
            }, 10);
        }

        function closeDeleteModal() {
            deleteModalContent.classList.add('scale-95', 'opacity-0');
            deleteModalContent.classList.remove('scale-100', 'opacity-1ci00');
            deleteModal.classList.add('hidden', 'opacity-0');
            
            // Reset the form in the modal
            deleteConfirmText.value = '';
            deleteModalConfirmBtn.disabled = true;
            deleteModalConfirmBtn.classList.add('bg-gray-400', 'cursor-not-allowed');
            deleteModalConfirmBtn.classList.remove('bg-red-600', 'hover:bg-red-700');
            
            setTimeout(() => {
                deleteModal.classList.add('hidden');
            }, 300);
        }

        // Check if user has typed "DELETE"
        deleteConfirmText.addEventListener('input', () => {
            if (deleteConfirmText.value === 'DELETE') {
                deleteModalConfirmBtn.disabled = false;
                deleteModalConfirmBtn.classList.remove('bg-gray-400', 'cursor-not-allowed');
                deleteModalConfirmBtn.classList.add('bg-red-600', 'hover:bg-red-700');
            } else {
                deleteModalConfirmBtn.disabled = true;
                deleteModalConfirmBtn.classList.add('bg-gray-400', 'cursor-not-allowed');
                deleteModalConfirmBtn.classList.remove('bg-red-600', 'hover:bg-red-700');
            }
        });

        deleteAccountBtn.addEventListener('click', openDeleteModal);
        deleteModalCloseBtn.addEventListener('click', closeDeleteModal);
        deleteModalCancelBtn.addEventListener('click', closeDeleteModal);
        
        deleteModal.addEventListener('click', (e) => {
            if (e.target === deleteModal) {
                closeDeleteModal();
            }
        });

        deleteModalConfirmBtn.addEventListener('click', () => {
            // --- FRONTEND-ONLY DEMO ---
            // In a real app, send fetch() request to delete the user
            alert('Account deleted. Redirecting...');
            window.location.href = "paycard_login.html";
        });
    }

});