document.addEventListener("DOMContentLoaded", function () {
    console.log("Settings page script loaded.");

    const username = document.body.dataset.username;
    if (!username) {
        console.error("Username not found in body dataset.");
        return;
    }

    // --- Update Email ---
    const updateEmailForm = document.getElementById('update-email-form');
    const updateEmailBtn = document.getElementById('update-email-btn');
    const emailErrorMsg = document.getElementById('settings-email-error');
    const emailSuccessMsg = document.getElementById('settings-email-success');

    if (updateEmailForm) {
        updateEmailForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            emailErrorMsg.classList.add('hidden');
            emailSuccessMsg.classList.add('hidden');
            
            const newEmail = document.getElementById('settings-email').value;

            try {
                const response = await fetch(`/u/${username}/profile/edit`, {
                    method: 'PATCH',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ email: newEmail })
                });

                const result = await response.json();

                if (!result.success) {
                    emailErrorMsg.textContent = result.message || 'Failed to update email.';
                    emailErrorMsg.classList.remove('hidden');
                    return;
                }

                emailSuccessMsg.textContent = 'Email updated successfully!';
                emailSuccessMsg.classList.remove('hidden');
                
            } catch (err) {
                console.error('Email update error:', err);
                emailErrorMsg.textContent = 'Network error, please try again.';
                emailErrorMsg.classList.remove('hidden');
            }
        });
    }

    // --- Update Password ---
    const changePasswordForm = document.getElementById('settings-change-password-form');
    const passwordErrorMsg = document.getElementById('settings-password-error');
    const passwordSuccessMsg = document.getElementById('settings-password-success');

    if (changePasswordForm) {
        changePasswordForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            passwordErrorMsg.classList.add('hidden');
            passwordSuccessMsg.classList.add('hidden');

            const currentPassword = document.getElementById('settings_current_password').value;
            const newPassword = document.getElementById('settings_new_password').value;
            const confirmPassword = document.getElementById('settings_confirm_password').value;

            if (newPassword !== confirmPassword) {
                passwordErrorMsg.textContent = "New password and confirm password do not match.";
                passwordErrorMsg.classList.remove('hidden');
                return;
            }

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
                    passwordErrorMsg.textContent = result.message || 'Failed to change password.';
                    passwordErrorMsg.classList.remove('hidden');
                    return;
                }

                passwordSuccessMsg.textContent = 'Password updated successfully!';
                passwordSuccessMsg.classList.remove('hidden');
                changePasswordForm.reset();
                
            } catch (err) {
                console.error('Password change error:', err);
                passwordErrorMsg.textContent = 'Network error, please try again.';
                passwordErrorMsg.classList.remove('hidden');
            }
        });
    }

    // --- Mock Preferences Toggles ---
    const toggles = ['toggle-2fa', 'toggle-email-notif', 'toggle-sms-notif', 'toggle-dark-mode'];
    
    toggles.forEach(toggleId => {
        const toggle = document.getElementById(toggleId);
        if (toggle) {
            toggle.addEventListener('change', (e) => {
                const settingName = toggleId.replace('toggle-', '').replace(/-/g, ' ');
                const status = e.target.checked ? 'enabled' : 'disabled';
                // Simply log for now, as there is no backend for this yet
                console.log(`${settingName} has been ${status}.`);
                // Could display a toast notification here
            });
        }
    });

});
