document.addEventListener("DOMContentLoaded", function () {
    // Step Elements
    const forgotForm = document.getElementById('forgot-form');
    const emailStep = document.getElementById('email-step');
    const otpStep = document.getElementById('otp-step');
    const passwordStep = document.getElementById('password-step');

    // Texts
    const formTitle = document.getElementById('form-title');
    const formSubtitle = document.getElementById('form-subtitle');

    // Buttons
    const sendLinkBtn = document.getElementById('send-link-btn');
    const confirmOtpBtn = document.getElementById('confirm-otp-btn');
    const submitButton = document.getElementById('reset-submit-btn');

    // Password Elements
    const passwordInput = document.getElementById('new_password');
    const confirmPasswordInput = document.getElementById('confirm_password');

    // Input Elements
    const emailInput = document.getElementById('email');
    const otpInputs = Array.from(document.querySelectorAll('.otp-digit'));

    // Error Elements
    const emailError = document.getElementById('email-error');
    const otpError = document.getElementById('otp-error');
    const newPasswordError = document.getElementById('new-password-error');
    const confirmPasswordError = document.getElementById('confirm-password-error');

    // Timer & Resend
    const otpTimer = document.getElementById('otp-timer');
    const resendOtpBtn = document.getElementById('resend-otp-btn');
    let timerInterval;
    let resendCount = 0;
    const MAX_RESEND = 3;

    function showFieldError(input, errorEl, msg) {
        if (!errorEl) return;
        errorEl.textContent = msg;
        errorEl.classList.remove('hidden');
        if (Array.isArray(input)) {
            input.forEach(i => {
                i.classList.remove('border-gray-300', 'focus:ring-blue-500', 'focus:border-blue-500');
                i.classList.add('border-red-500', 'focus:ring-red-500', 'focus:border-red-500');
            });
        } else if (input) {
            input.classList.remove('border-gray-300', 'focus:ring-blue-500', 'focus:border-blue-500');
            input.classList.add('border-red-500', 'focus:ring-red-500', 'focus:border-red-500');
        }
    }

    function hideFieldError(input, errorEl) {
        if (!errorEl) return;
        errorEl.classList.add('hidden');
        if (Array.isArray(input)) {
            input.forEach(i => {
                i.classList.remove('border-red-500', 'focus:ring-red-500', 'focus:border-red-500');
                i.classList.add('border-gray-300', 'focus:ring-blue-500', 'focus:border-blue-500');
            });
        } else if (input) {
            input.classList.remove('border-red-500', 'focus:ring-red-500', 'focus:border-red-500');
            input.classList.add('border-gray-300', 'focus:ring-blue-500', 'focus:border-blue-500');
        }
    }

    function hideAllErrors() {
        hideFieldError(emailInput, emailError);
        hideFieldError(otpInputs, otpError);
        hideFieldError(passwordInput, newPasswordError);
        hideFieldError(confirmPasswordInput, confirmPasswordError);
    }

    function startTimer() {
        let timeLeft = 300; // 5 minutes
        if (resendOtpBtn) resendOtpBtn.disabled = true;
        if (otpTimer) otpTimer.textContent = '05:00';
        
        if (timerInterval) clearInterval(timerInterval);
        
        timerInterval = setInterval(() => {
            timeLeft--;
            let minutes = Math.floor(timeLeft / 60);
            let seconds = timeLeft % 60;
            let formattedTime = `${minutes < 10 ? '0' : ''}${minutes}:${seconds < 10 ? '0' : ''}${seconds}`;
            if (otpTimer) otpTimer.textContent = formattedTime;
            
            if (timeLeft <= 0) {
                clearInterval(timerInterval);
                if (otpTimer) otpTimer.textContent = '00:00';
                if (resendOtpBtn && resendCount < MAX_RESEND) {
                    resendOtpBtn.disabled = false;
                } else if (resendOtpBtn && resendCount >= MAX_RESEND) {
                    if (otpTimer) otpTimer.textContent = 'Max attempts';
                }
            }
        }, 1000);
    }

    // Validation Elements
    const checklist = document.getElementById('validation-checklist');
    const lengthCheck = document.getElementById('length-check');
    const caseCheck = document.getElementById('case-check');
    const numCheck = document.getElementById('num-check');
    const matchCheck = document.getElementById('match-check');
    const specialCheck = document.getElementById('special-check');

    const hasLower = new RegExp(/[a-z]/);
    const hasUpper = new RegExp(/[A-Z]/);
    const hasNumber = new RegExp(/[0-9]/);
    const hasSpecial = new RegExp(/[^A-Za-z0-9]/);

    let currentEmail = "";
    let currentOtp = "";

    function updateChecklistItem(checkElement, isValid) {
        if (!checkElement) return;
        const icon = checkElement.querySelector('i');
        if (isValid) {
            checkElement.classList.remove('text-gray-500', 'bg-gray-100');
            checkElement.classList.add('text-green-700', 'bg-green-100');
            icon.classList.remove('fa-circle');
            icon.classList.add('fa-check');
        } else {
            checkElement.classList.remove('text-green-700', 'bg-green-100');
            checkElement.classList.add('text-gray-500', 'bg-gray-100');
            icon.classList.remove('fa-check');
            icon.classList.add('fa-circle');
        }
    }

    function validateForm() {
        if (!passwordInput || !confirmPasswordInput) return;

        const password = passwordInput.value;
        const confirmPassword = confirmPasswordInput.value;

        const isLengthValid = password.length >= 8;
        const isCaseValid = hasLower.test(password) && hasUpper.test(password);
        const isNumValid = hasNumber.test(password);
        const isSpecialValid = hasSpecial.test(password);
        const passwordsMatch = password === confirmPassword && password.length > 0;

        updateChecklistItem(lengthCheck, isLengthValid);
        updateChecklistItem(caseCheck, isCaseValid);
        updateChecklistItem(numCheck, isNumValid);
        updateChecklistItem(specialCheck, isSpecialValid);
        updateChecklistItem(matchCheck, passwordsMatch);

        const allValid = isLengthValid && isCaseValid && isNumValid && isSpecialValid && passwordsMatch;

        if (submitButton) {
            if (allValid) {
                submitButton.disabled = false;
                hideAllErrors();
            } else {
                submitButton.disabled = true;
            }
        }
    }

    if (passwordInput && confirmPasswordInput) {
        passwordInput.addEventListener('input', () => { hideFieldError(passwordInput, newPasswordError); validateForm(); });
        confirmPasswordInput.addEventListener('input', () => { hideFieldError(confirmPasswordInput, confirmPasswordError); validateForm(); });
    }

    if (emailInput) {
        emailInput.addEventListener('input', () => hideFieldError(emailInput, emailError));
    }

    otpInputs.forEach((input, index) => {
        input.addEventListener('input', (e) => {
            input.value = input.value.replace(/[^0-9]/g, '');
            hideFieldError(otpInputs, otpError);
            if (input.value && index < otpInputs.length - 1) {
                otpInputs[index + 1].focus();
            }
        });
        
        input.addEventListener('keydown', (e) => {
            if (e.key === 'Backspace' && !input.value && index > 0) {
                otpInputs[index - 1].focus();
            } else if (e.key === 'Enter') {
                e.preventDefault();
                if (confirmOtpBtn) confirmOtpBtn.click();
            }
        });

        input.addEventListener('paste', (e) => {
            e.preventDefault();
            const pastedData = e.clipboardData.getData('text').replace(/[^0-9]/g, '').slice(0, 6);
            for (let i = 0; i < pastedData.length; i++) {
                if (otpInputs[i]) {
                    otpInputs[i].value = pastedData[i];
                    if (i < 5) otpInputs[i + 1].focus();
                }
            }
            hideFieldError(otpInputs, otpError);
        });
    });

    if (forgotForm) {
        forgotForm.addEventListener('submit', (e) => e.preventDefault());
        
        if (emailInput) {
            emailInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    if (sendLinkBtn) sendLinkBtn.click();
                }
            });
        }
        
        const handlePasswordEnter = (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                if (submitButton && !submitButton.disabled) submitButton.click();
            }
        };
        
        if (passwordInput) passwordInput.addEventListener('keypress', handlePasswordEnter);
        if (confirmPasswordInput) confirmPasswordInput.addEventListener('keypress', handlePasswordEnter);
    }

    if (forgotForm && sendLinkBtn && confirmOtpBtn && emailStep && otpStep) {

        // --- STEP 1: Handle Email submission ---
        sendLinkBtn.addEventListener('click', async function (event) {
            event.preventDefault();
            hideAllErrors();

            const emailVal = emailInput.value;
            const emailRegex = /^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$/;

            if (!emailVal || !emailRegex.test(emailVal)) {
                // validateEmailDynamically will handle the UI
                if (emailInput) emailInput.dispatchEvent(new Event('input'));
                return;
            }

            sendLinkBtn.disabled = true;
            sendLinkBtn.textContent = "Sending...";

            try {
                const response = await fetch('/v1/forgot-password/send-otp', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ email: emailVal })
                });

                if (response.ok) {
                    currentEmail = emailVal;
                    emailStep.classList.add('hidden');
                    otpStep.classList.remove('hidden');

                    formTitle.textContent = "Check your email";
                    formSubtitle.textContent = "We sent a 6-digit code to your email.";
                    
                    // Reset resend count and start timer
                    resendCount = 0;
                    startTimer();
                } else {
                    const error = await response.text();
                    showFieldError(emailInput, emailError, "Error: " + error);
                }
            } catch (error) {
                showFieldError(emailInput, emailError, "An error occurred. Please try again.");
            } finally {
                sendLinkBtn.disabled = false;
                sendLinkBtn.textContent = "Send Reset Link";
            }
        });

        // --- STEP 2: Handle OTP submission ---
        confirmOtpBtn.addEventListener('click', async function (event) {
            event.preventDefault();
            hideAllErrors();

            const otpVal = otpInputs.map(i => i.value).join('');

            if (!otpVal || otpVal.length !== 6) {
                showFieldError(otpInputs, otpError, "Please enter a valid 6-digit OTP");
                return;
            }

            confirmOtpBtn.disabled = true;
            confirmOtpBtn.textContent = "Verifying...";

            try {
                const response = await fetch('/v1/forgot-password/verify-otp', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ email: currentEmail, otp: otpVal })
                });

                if (response.ok) {
                    clearInterval(timerInterval);
                    currentOtp = otpVal;
                    otpStep.classList.add('hidden');
                    if (passwordStep) passwordStep.classList.remove('hidden');

                    formTitle.textContent = "Set New Password";
                    formSubtitle.textContent = "Please create a secure password.";
                } else {
                    const error = await response.text();
                    showFieldError(otpInputs, otpError, "Invalid OTP: " + error);
                }
            } catch (error) {
                showFieldError(otpInputs, otpError, "An error occurred. Please try again.");
            } finally {
                confirmOtpBtn.disabled = false;
                confirmOtpBtn.textContent = "Confirm Code & Continue";
            }
        });

        // --- Handle Resend OTP ---
        if (resendOtpBtn) {
            resendOtpBtn.addEventListener('click', async function(e) {
                e.preventDefault();
                hideAllErrors();

                if (resendCount >= MAX_RESEND) {
                    showFieldError(otpInputs, otpError, "Maximum resend attempts reached.");
                    return;
                }
                
                resendOtpBtn.disabled = true;
                resendOtpBtn.textContent = "Sending...";
                
                try {
                    const response = await fetch('/v1/forgot-password/send-otp', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ email: currentEmail })
                    });
                    
                    if (response.ok) {
                        resendCount++;
                        resendOtpBtn.textContent = "Resend";
                        startTimer();
                        otpInputs.forEach(i => i.value = '');
                        otpInputs[0].focus();
                    } else {
                        const error = await response.text();
                        showFieldError(otpInputs, otpError, "Error: " + error);
                        resendOtpBtn.textContent = "Resend";
                        resendOtpBtn.disabled = false;
                    }
                } catch (error) {
                    showFieldError(otpInputs, otpError, "An error occurred. Please try again.");
                    resendOtpBtn.textContent = "Resend";
                    resendOtpBtn.disabled = false;
                }
            });
        }

        // --- STEP 3: Handle Password Reset submission ---
        if (submitButton) {
            submitButton.addEventListener('click', async function (event) {
                event.preventDefault();
                hideAllErrors();

                validateForm();

                if (submitButton.disabled) {
                    showFieldError(passwordInput, newPasswordError, 'Please fix the errors in the password checklist.');
                } else {
                    if (!currentEmail || !currentOtp) {
                        showFieldError(passwordInput, newPasswordError, 'Missing email or OTP. Please refresh and try again.');
                        return;
                    }

                    submitButton.disabled = true;
                    submitButton.textContent = "Updating...";

                    try {
                        const response = await fetch('/v1/reset-password', {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify({
                                email: currentEmail,
                                otp: currentOtp,
                                new_password: passwordInput.value
                            })
                        });

                        if (response.ok) {
                            hideAllErrors();
                            formTitle.textContent = "Success!";
                            formSubtitle.textContent = "Password reset successfully. Redirecting to login...";
                            if (passwordStep) passwordStep.classList.add('hidden');
                            
                            setTimeout(() => {
                                window.location.href = "/login";
                            }, 2000);
                        } else {
                            const error = await response.text();
                            showFieldError(passwordInput, newPasswordError, "Error: " + error);
                            submitButton.disabled = false;
                            submitButton.textContent = "Set New Password";
                        }
                    } catch (error) {
                        showFieldError(passwordInput, newPasswordError, "An error occurred. Please try again.");
                        submitButton.disabled = false;
                        submitButton.textContent = "Set New Password";
                    }
                }
            });
        }
    }
});
