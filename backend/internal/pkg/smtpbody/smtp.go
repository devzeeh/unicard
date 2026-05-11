package smtp

import (
	"fmt"
	"time"
)

func OTPCode() string {
	return `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Password Reset OTP</title>
	<style>
		body { font-family: 'Inter', Arial, sans-serif; background-color: #f4f4f5; margin: 0; padding: 0; }
		.container { max-width: 600px; margin: 40px auto; background-color: #ffffff; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); overflow: hidden; }
		.header { background-color: #0f172a; color: #ffffff; padding: 20px; text-align: center; }
		.header h1 { margin: 0; font-size: 24px; font-weight: 600; }
		.content { padding: 30px; color: #334155; line-height: 1.6; }
		.content p { margin: 0 0 15px; }
		.otp-container { text-align: center; margin: 30px 0; }
		.otp-code { display: inline-block; font-size: 32px; font-weight: 700; letter-spacing: 5px; color: #2563eb; background-color: #eff6ff; padding: 15px 25px; border-radius: 6px; border: 1px dashed #bfdbfe; }
		.footer { background-color: #f8fafc; padding: 15px; text-align: center; font-size: 12px; color: #64748b; border-top: 1px solid #e2e8f0; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Unicard</h1>
		</div>
		<div class="content">
			<p>Hello %s,</p>
			<p>We received a request to reset your password. Please use the following One-Time Password (OTP) to proceed. This OTP is valid for <strong>5 minutes</strong>.</p>
			<div class="otp-container">
				<div class="otp-code">%s</div>
			</div>
			<p>If you did not request a password reset, please ignore this email or contact support if you have concerns.</p>
			<p>Thank you,<br>The Unicard Team</p>
		</div>
		<div class="footer">
			&copy; ` + Year() + ` Unicard. All rights reserved.
		</div>
	</div>
</body>
</html>`
}

func Year() string {
	return fmt.Sprintf("%d", time.Now().Year())
}
