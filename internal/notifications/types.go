package notifications

// NotificationType represents the type of notification to send
type NotificationType string

const (
	NotificationTypeWelcome           NotificationType = "welcome"
	NotificationTypePasswordReset     NotificationType = "password_reset"
	NotificationTypeOrderConfirmation NotificationType = "order_confirmation"
	NotificationTypeLoginNotification NotificationType = "login_notification"
	NotificationTypeUserLoggedIn      NotificationType = "user_logged_in"
)

// Notification represents a notification message from the queue
type Notification struct {
	Type NotificationType `json:"type"`

	// Common fields
	Email    string `json:"email"`
	Username string `json:"username,omitempty"`
	UserID   int64  `json:"user_id,omitempty"`

	// Password reset fields
	ResetToken string `json:"reset_token,omitempty"`

	// Order confirmation fields
	OrderID string  `json:"order_id,omitempty"`
	Total   float64 `json:"total,omitempty"`

	// Login notification fields
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	LoginTime string `json:"login_time,omitempty"`
}
