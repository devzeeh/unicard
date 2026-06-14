document.addEventListener("DOMContentLoaded", function () {
    console.log("Profile page script loaded.");

    const username = document.body.dataset.username;

    // --- Profile Edit Elements ---
    const editProfileBtn = document.getElementById('edit-profile-btn');
    const cancelEditBtn = document.getElementById('cancel-edit-btn');
    const saveProfileBtn = document.getElementById('save-profile-btn');
    const profileActions = document.getElementById('profile-edit-actions');
    const profileView = document.getElementById('profile-details-view');
    const profileEditForm = document.getElementById('profile-details-edit');

    if (editProfileBtn && cancelEditBtn && profileActions && profileView && profileEditForm && saveProfileBtn) {
        editProfileBtn.addEventListener('click', () => {
            // Pre-fill edit form with current values
            document.getElementById('full_name').value = document.getElementById('profile-view-name').innerText.trim();
            document.getElementById('email').value = document.getElementById('profile-view-email').innerText.trim();
            document.getElementById('phone').value = document.getElementById('profile-view-phone').innerText.trim();

            profileView.classList.add('hidden');
            profileEditForm.classList.remove('hidden');
            editProfileBtn.classList.add('hidden');
            profileActions.classList.remove('hidden');
        });

        cancelEditBtn.addEventListener('click', (e) => {
            e.preventDefault();
            profileEditForm.classList.add('hidden');
            profileView.classList.remove('hidden');
            profileActions.classList.add('hidden');
            editProfileBtn.classList.remove('hidden');
        });

        saveProfileBtn.addEventListener('click', async (e) => {
            e.preventDefault();

            const newName = document.getElementById('full_name').value;
            const newEmail = document.getElementById('email').value;
            const newPhone = document.getElementById('phone').value;

            try {
                const response = await fetch(`/u/${username}/profile/edit`, {
                    method: 'PATCH',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        full_name: newName,
                        email: newEmail,
                        phone_number: newPhone
                    })
                });

                const result = await response.json();

                if (!result.success) {
                    alert(result.message || 'Failed to update profile');
                    return;
                }

                const nameSpan = document.getElementById('profile-view-name');
                const emailSpan = document.getElementById('profile-view-email');
                const phoneSpan = document.getElementById('profile-view-phone');

                if (nameSpan) nameSpan.innerText = newName;
                if (emailSpan) emailSpan.innerText = newEmail;
                if (phoneSpan) phoneSpan.innerText = newPhone;

                profileEditForm.classList.add('hidden');
                profileView.classList.remove('hidden');
                profileActions.classList.add('hidden');
                editProfileBtn.classList.remove('hidden');

            } catch (err) {
                console.error('Profile update error:', err);
                alert('Network error, please try again.');
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

        const passwordsMatch = password === confirmPassword && password.length > 0;
        
        passwordSubmitBtn.disabled = !passwordsMatch;
    }

    if (passwordForm) {
        const verifyCurrentPasswordBtn = document.getElementById('verify-current-password-btn');
        const currentPasswordInput = document.getElementById('current_password');
        const newPasswordSection = document.getElementById('new-password-section');
        let currentPasswordVerified = false;

        if (verifyCurrentPasswordBtn) {
            verifyCurrentPasswordBtn.addEventListener('click', async () => {
                const currentPassword = currentPasswordInput.value;
                if (!currentPassword) {
                    passwordErrorMsg.textContent = 'Please enter your current password.';
                    passwordErrorMsg.classList.remove('hidden');
                    return;
                }

                passwordErrorMsg.classList.add('hidden');
                verifyCurrentPasswordBtn.disabled = true;
                verifyCurrentPasswordBtn.textContent = 'Verifying...';

                try {
                    const response = await fetch(`/v1/user/${username}/profile/verify-password`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ current_password: currentPassword })
                    });

                    const result = await response.json();

                    if (!result.success) {
                        passwordErrorMsg.textContent = result.message || 'Current password verification failed.';
                        passwordErrorMsg.classList.remove('hidden');
                        verifyCurrentPasswordBtn.disabled = false;
                        verifyCurrentPasswordBtn.textContent = 'Verify';
                        return;
                    }

                    // Verification successful
                    currentPasswordVerified = true;
                    currentPasswordInput.disabled = true; // prevent changing it
                    verifyCurrentPasswordBtn.classList.add('hidden');
                    newPasswordSection.classList.remove('hidden');
                    
                    // Trigger validation to check if new inputs are valid (which they won't be yet, so button stays disabled)
                    validatePasswordForm();

                } catch (err) {
                    console.error('Password verification error:', err);
                    passwordErrorMsg.textContent = 'Network error, please try again.';
                    passwordErrorMsg.classList.remove('hidden');
                    verifyCurrentPasswordBtn.disabled = false;
                    verifyCurrentPasswordBtn.textContent = 'Verify';
                }
            });
        }

        newPasswordInput.addEventListener('input', validatePasswordForm);
        confirmPasswordInput.addEventListener('input', validatePasswordForm);

        passwordForm.addEventListener('submit', async (e) => {
            e.preventDefault();

            if (!currentPasswordVerified) {
                passwordErrorMsg.textContent = 'Please verify your current password first.';
                passwordErrorMsg.classList.remove('hidden');
                return;
            }

            if (passwordSubmitBtn.disabled) {
                passwordErrorMsg.textContent = 'Passwords do not match.';
                passwordErrorMsg.classList.remove('hidden');
                return;
            }

            const currentPassword = currentPasswordInput.value;
            const newPassword = newPasswordInput.value;
            const confirmPassword = confirmPasswordInput.value;

            try {
                const response = await fetch(`/u/${username}/profile/password`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        current_password: currentPassword,
                        new_password: newPassword,
                        confirm_password: confirmPassword
                    })
                });

                const result = await response.json();

                if (!result.success) {
                    passwordErrorMsg.textContent = result.message || 'Failed to change password';
                    passwordErrorMsg.classList.remove('hidden');
                    return;
                }

                passwordErrorMsg.classList.add('hidden');
                alert('Password changed successfully!');
                
                // Reset form state
                passwordForm.reset();
                if (checklist) checklist.classList.add('hidden');
                passwordSubmitBtn.disabled = true;
                currentPasswordVerified = false;
                currentPasswordInput.disabled = false;
                if (verifyCurrentPasswordBtn) {
                    verifyCurrentPasswordBtn.classList.remove('hidden');
                    verifyCurrentPasswordBtn.disabled = false;
                    verifyCurrentPasswordBtn.textContent = 'Verify';
                }
                newPasswordSection.classList.add('hidden');

            } catch (err) {
                console.error('Password change error:', err);
                passwordErrorMsg.textContent = 'Network error, please try again.';
                passwordErrorMsg.classList.remove('hidden');
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
            deleteModalContent.classList.remove('scale-100', 'opacity-100');
            deleteModal.classList.add('hidden', 'opacity-0');

            deleteConfirmText.value = '';
            deleteModalConfirmBtn.disabled = true;
            deleteModalConfirmBtn.classList.add('bg-gray-400', 'cursor-not-allowed');
            deleteModalConfirmBtn.classList.remove('bg-red-600', 'hover:bg-red-700');

            setTimeout(() => {
                deleteModal.classList.add('hidden');
            }, 300);
        }

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
            // Not yet wired to backend
            alert('Account deleted. Redirecting...');
            window.location.href = "paycard_login.html";
        });
    }

});