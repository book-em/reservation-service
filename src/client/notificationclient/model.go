package notificationclient

import "time"

type CreateNotificationDTO struct {
	ReceiverID  uint             `json:"receiverId"`
	Type        NotificationType `json:"type"`
	Subject     uint             `json:"subject"`
	Object      uint             `json:"object"`
	StarsNumber int              `json:"starsNumber,omitempty"`
}

type NotificationDTO struct {
	ID          string           `json:"id"`
	ReceiverID  uint             `json:"receiverId"`
	Type        NotificationType `json:"type"`
	Subject     uint             `json:"subject"`
	Object      uint             `json:"object"`
	StarsNumber int              `json:"starsNumber,omitempty"`
	IsRead      bool             `json:"isRead"`
	CreatedAt   time.Time        `json:"createdAt"`
}

type NotificationType string

const (
	ReservationRequested NotificationType = "reservation_requested"
	ReservationCancelled NotificationType = "reservation_cancelled"
	ReservationAccepted  NotificationType = "reservation_accepted"
	ReservationDeclined  NotificationType = "reservation_declined"
)
