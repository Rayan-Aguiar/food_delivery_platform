package events

const (
	OrderCreated       = "order.created.v1"
	OrderConfirmed     = "order.confirmed.v1"
	OrderCancelled     = "order.cancelled.v1"
	OrderStatusChanged = "order.status.changed.v1"

	PaymentApproved  = "payment.approved.v1"
	PaymentFailed    = "payment.failed.v1"
	PaymentRefunded  = "payment.refunded.v1"
	PaymentRefundReq = "payment.refund.requested.v1"

	DeliveryRequested = "delivery.requested.v1"
	DeliveryStarted   = "delivery.started.v1"
	DeliveryCompleted = "delivery.completed.v1"
	DeliveryFailed    = "delivery.failed.v1"

	UserAuthRegistered = "user.auth.registered.v1"
	AuthLoginSucceeded = "auth.login.succeeded.v1"

	NotificationSent   = "notification.sent.v1"
	NotificationFailed = "notification.failed.v1"
)
