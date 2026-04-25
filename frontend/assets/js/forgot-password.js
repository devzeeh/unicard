document.addEventListener("DOMContentLoaded", function () {
    // Get all the elements
    const forgotForm = document.getElementById('forgot-form');
    const emailStep = document.getElementById('email-step');
    const otpStep = document.getElementById('otp-step');
    
    const formTitle = document.getElementById('form-title');
    const formSubtitle = document.getElementById('form-subtitle');

    // Get the two buttons
    const sendLinkBtn = document.getElementById('send-link-btn');
    const confirmOtpBtn = document.getElementById('confirm-otp-btn');

    // This check is in case the script is loaded on a page without these elements
    if (forgotForm && sendLinkBtn && confirmOtpBtn && emailStep && otpStep) {
        
        // --- STEP 1: Handle Email submission ---
        // Listen for a CLICK on the "Send Reset Link" button
        sendLinkBtn.addEventListener('click', function (event) {
            event.preventDefault(); // Stop the form from submitting

            // --- THIS IS A FRONTEND-ONLY DEMO ---
            // In a real app, you would make a fetch() call to your backend here
            // to check if the email exists and to send the OTP.
            
            const email = document.getElementById('email').value;
            
            // For this demo, we'll just simulate a successful email send
            // (assuming the email was valid).
            console.log(`Simulating sending OTP to: ${email}`);
            
            // Hide the email step and show the OTP step
            emailStep.classList.add('hidden');
            otpStep.classList.remove('hidden');

            // Update the titles
            formTitle.textContent = "Check your email";
            formSubtitle.textContent = "We sent a 6-digit code to your email.";
        });


        // --- STEP 2: Handle OTP submission ---
        // Listen for a CLICK on the "Confirm Code" button
        confirmOtpBtn.addEventListener('click', function(event) {
            event.preventDefault(); // Stop the form from submitting

            // --- THIS IS A FRONTEND-ONLY DEMO ---
            // This is where you would verify the OTP with your backend
            const otp = document.getElementById('otp').value;
            
            // For demo purposes, we'll just check for a static OTP
            if(otp === "123456") { // Example: a valid OTP
                alert("OTP Correct! You would now be redirected to reset your password.");
                // redirect to login page
                window.location.href = "../templates/resetpassword.html";

                // In a real app, you would redirect to the password reset page
                // window.location.href = "/reset-password.html"; 
            } else {
                alert("Invalid OTP. Please try again.");
            }
        });
    }
});

