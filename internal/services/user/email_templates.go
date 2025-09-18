package user

import "fmt"

// EmailTemplates contains all email template functions
type EmailTemplates struct {
	baseURL string
}

// NewEmailTemplates creates a new email templates instance
func NewEmailTemplates(baseURL string) *EmailTemplates {
	return &EmailTemplates{
		baseURL: baseURL,
	}
}

// GetVerificationEmailTemplate returns the HTML template for email verification
func (t *EmailTemplates) GetVerificationEmailTemplate(displayName, token string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px; background: #f9f9f9; }
        .button { display: inline-block; padding: 12px 30px; background: #4F46E5; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎵 NameThatSong</h1>
        </div>
        <div class="content">
            <h2>Welcome to NameThatSong!</h2>
            <p>Hi %s,</p>
            <p>Thanks for signing up! Please verify your email address by clicking the button below:</p>
            <p><a href="%s/verify-email?token=%s" class="button">Verify Email Address</a></p>
            <p>Or copy and paste this link into your browser:</p>
            <p><code>%s/verify-email?token=%s</code></p>
            <p><strong>This link will expire in 10 minutes.</strong></p>
            <p>If you didn't create this account, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>Best regards,<br>The NameThatSong Team</p>
        </div>
    </div>
</body>
</html>
`, displayName, t.baseURL, token, t.baseURL, token)
}

// GetPasswordResetEmailTemplate returns the HTML template for password reset
func (t *EmailTemplates) GetPasswordResetEmailTemplate(displayName, token string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Your Password</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #DC2626; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px; background: #f9f9f9; }
        .button { display: inline-block; padding: 12px 30px; background: #DC2626; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .warning { background: #FEF3C7; border-left: 4px solid #F59E0B; padding: 15px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 Password Reset</h1>
        </div>
        <div class="content">
            <h2>Password Reset Request</h2>
            <p>Hi %s,</p>
            <p>We received a request to reset your password. Click the button below to set a new password:</p>
            <p><a href="%s/reset-password?token=%s" class="button">Reset Password</a></p>
            <p>Or copy and paste this link into your browser:</p>
            <p><code>%s/reset-password?token=%s</code></p>
            <div class="warning">
                <p><strong>⚠️ Security Notice:</strong></p>
                <p>This link will expire in 15 minutes for your security.</p>
                <p>If you didn't request this reset, please ignore this email.</p>
            </div>
        </div>
        <div class="footer">
            <p>Best regards,<br>The NameThatSong Team</p>
        </div>
    </div>
</body>
</html>
`, displayName, t.baseURL, token, t.baseURL, token)
}

// GetWelcomeEmailTemplate returns the HTML template for welcome email
func (t *EmailTemplates) GetWelcomeEmailTemplate(displayName string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to NameThatSong</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #10B981; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px; background: #f9f9f9; }
        .button { display: inline-block; padding: 12px 30px; background: #10B981; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .features { background: white; padding: 20px; border-radius: 8px; margin: 20px 0; }
        .feature { margin: 15px 0; padding: 10px; border-left: 4px solid #10B981; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Welcome to NameThatSong!</h1>
        </div>
        <div class="content">
            <h2>You're all set, %s!</h2>
            <p>Your email has been successfully verified! You're now ready to start playing music guessing games.</p>
            
            <div class="features">
                <h3>Here's what you can do:</h3>
                <div class="feature">
                    <strong>🎵 Play Solo Games</strong><br>
                    Test your music knowledge with various difficulty levels
                </div>
                <div class="feature">
                    <strong>🏆 Compete with Friends</strong><br>
                    Challenge friends in multiplayer games and climb the leaderboard
                </div>
                <div class="feature">
                    <strong>📊 Track Your Progress</strong><br>
                    Monitor your scores, streaks, and improvement over time
                </div>
                <div class="feature">
                    <strong>🎶 Connect Spotify</strong><br>
                    Link your Spotify account for personalized playlists and recommendations
                </div>
            </div>
            
            <p><a href="%s/dashboard" class="button">Start Playing Now!</a></p>
            
            <p>Need help getting started? Check out our <a href="%s/help">help guide</a> or contact support.</p>
        </div>
        <div class="footer">
            <p>Happy gaming!<br>The NameThatSong Team</p>
        </div>
    </div>
</body>
</html>
`, displayName, t.baseURL, t.baseURL)
}

// GetPasswordResetConfirmationTemplate returns the HTML template for password reset confirmation
func (t *EmailTemplates) GetPasswordResetConfirmationTemplate(displayName string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset Successful</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #10B981; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px; background: #f9f9f9; }
        .button { display: inline-block; padding: 12px 30px; background: #10B981; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .success { background: #D1FAE5; border-left: 4px solid #10B981; padding: 15px; margin: 20px 0; }
        .security { background: #FEF3C7; border-left: 4px solid #F59E0B; padding: 15px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ Password Reset Successful</h1>
        </div>
        <div class="content">
            <h2>Hi %s,</h2>
            
            <div class="success">
                <p><strong>✅ Success!</strong></p>
                <p>Your password has been successfully reset. You can now log in with your new password.</p>
            </div>
            
            <p><a href="%s/login" class="button">Login Now</a></p>
            
            <div class="security">
                <p><strong>🔒 Security Reminder:</strong></p>
                <p>If you didn't make this change, please contact us immediately at support@namethatSong.com</p>
            </div>
            
            <p>For your security, we recommend:</p>
            <ul>
                <li>Using a strong, unique password</li>
                <li>Enabling two-factor authentication (coming soon!)</li>
                <li>Logging out from public devices</li>
            </ul>
        </div>
        <div class="footer">
            <p>Best regards,<br>The NameThatSong Team</p>
        </div>
    </div>
</body>
</html>
`, displayName, t.baseURL)
}
