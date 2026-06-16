document.addEventListener("DOMContentLoaded", function () {
    console.log("Settings page script loaded.");

    const username = document.body.dataset.username;
    if (!username) {
        console.error("Username not found in body dataset.");
        return;
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

            const otpSent = await requestOtp();
            if (otpSent) {
                showOtpModal('password', {
                    current_password: currentPassword,
                    new_password: newPassword,
                    confirm_password: confirmPassword
                });
            }
        });
    }

    // --- Handle OTP Submit ---
    if (otpForm) {
        otpForm.addEventListener('submit', async (e) => {
            e.preventDefault();

            const otp = otpInput.value;
            if (otp.length !== 6) {
                otpErrorMsg.textContent = 'Please enter a valid 6-digit OTP.';
                otpErrorMsg.classList.remove('hidden');
                return;
            }

            otpSubmitBtn.disabled = true;
            otpSubmitBtn.textContent = 'Verifying...';
            otpErrorMsg.classList.add('hidden');

            try {
                let endpoint = '';
                let payload = { ...pendingData, otp: otp };

                if (currentOtpAction === 'password') {
                    endpoint = `/v1/user/${username}/settings/password`;
                }

                const response = await fetch(endpoint, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });

                const result = await response.json();

                if (!result.success) {
                    otpErrorMsg.textContent = result.message || 'Verification failed.';
                    otpErrorMsg.classList.remove('hidden');
                } else {
                    closeOtpModal();
                    if (currentOtpAction === 'password') {
                        passwordSuccessMsg.textContent = 'Password updated successfully!';
                        passwordSuccessMsg.classList.remove('hidden');
                        if (changePasswordForm) changePasswordForm.reset();
                    }
                }
            } catch (err) {
                console.error('OTP Submit error:', err);
                otpErrorMsg.textContent = 'Network error, please try again.';
                otpErrorMsg.classList.remove('hidden');
            } finally {
                otpSubmitBtn.disabled = false;
                otpSubmitBtn.textContent = 'Verify & Confirm';
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
