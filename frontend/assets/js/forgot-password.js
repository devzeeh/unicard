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
    const errorMessage = document.getElementById('error-message');

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
            checkElement.classList.add('valid');
            icon.classList.remove('fa-times-circle', 'text-red-500');
            icon.classList.add('fa-check-circle', 'text-green-600');
        } else {
            checkElement.classList.remove('valid');
            icon.classList.remove('fa-check-circle', 'text-green-600');
            icon.classList.add('fa-times-circle', 'text-red-500');
        }
    }

    function validateForm() {
        if (!passwordInput || !confirmPasswordInput) return;

        const password = passwordInput.value;
        const confirmPassword = confirmPasswordInput.value;

        if (password.length > 0 || confirmPassword.length > 0) {
            if (checklist) checklist.classList.remove('hidden');
        } else {
            if (checklist) checklist.classList.add('hidden');
        }

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
                if (errorMessage) errorMessage.classList.add('hidden');
            } else {
                submitButton.disabled = true;
            }
        }
    }

    if (passwordInput && confirmPasswordInput) {
        passwordInput.addEventListener('input', validateForm);
        confirmPasswordInput.addEventListener('input', validateForm);
    }

    if (forgotForm && sendLinkBtn && confirmOtpBtn && emailStep && otpStep) {

        // --- STEP 1: Handle Email submission ---
        sendLinkBtn.addEventListener('click', async function (event) {
            event.preventDefault();

            const emailVal = document.getElementById('email').value;
            if (!emailVal) {
                alert("Please enter your email");
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
                } else {
                    const error = await response.text();
                    alert("Error: " + error);
                }
            } catch (error) {
                alert("An error occurred. Please try again.");
            } finally {
                sendLinkBtn.disabled = false;
                sendLinkBtn.textContent = "Send Reset Link";
            }
        });

        // --- STEP 2: Handle OTP submission ---
        confirmOtpBtn.addEventListener('click', async function (event) {
            event.preventDefault();

            const otpVal = document.getElementById('otp').value;

            if (!otpVal) {
                alert("Please enter the OTP");
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
                    currentOtp = otpVal;
                    otpStep.classList.add('hidden');
                    if (passwordStep) passwordStep.classList.remove('hidden');

                    formTitle.textContent = "Set New Password";
                    formSubtitle.textContent = "Please create a secure password.";
                } else {
                    const error = await response.text();
                    alert("Invalid OTP: " + error);
                }
            } catch (error) {
                alert("An error occurred. Please try again.");
            } finally {
                confirmOtpBtn.disabled = false;
                confirmOtpBtn.textContent = "Confirm Code & Continue";
            }
        });

        // --- STEP 3: Handle Password Reset submission ---
        if (submitButton) {
            submitButton.addEventListener('click', async function (event) {
                event.preventDefault();

                validateForm();

                if (submitButton.disabled) {
                    if (errorMessage) {
                        errorMessage.textContent = 'Please fix the errors in the password checklist.';
                        errorMessage.classList.remove('hidden');
                    }
                } else {
                    if (!currentEmail || !currentOtp) {
                        if (errorMessage) {
                            errorMessage.textContent = 'Missing email or OTP. Please refresh and try again.';
                            errorMessage.classList.remove('hidden');
                        }
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
                            if (errorMessage) errorMessage.classList.add('hidden');
                            alert('Password reset successfully! Redirecting to login...');
                            window.location.href = "/login";
                        } else {
                            const error = await response.text();
                            if (errorMessage) {
                                errorMessage.textContent = "Error: " + error;
                                errorMessage.classList.remove('hidden');
                            }
                            submitButton.disabled = false;
                            submitButton.textContent = "Set New Password";
                        }
                    } catch (error) {
                        if (errorMessage) {
                            errorMessage.textContent = "An error occurred. Please try again.";
                            errorMessage.classList.remove('hidden');
                        }
                        submitButton.disabled = false;
                        submitButton.textContent = "Set New Password";
                    }
                }
            });
        }
    }
});
