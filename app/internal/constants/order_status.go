package constants

// 订单状态
const (
	OrderStatusCreated   = "created"
	OrderStatusPaid      = "paid"
	OrderStatusPreparing = "preparing"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
)

// 状态触发器（事件）
const (
	OrderTriggerPay      = "pay"
	OrderTriggerStart    = "start"
	OrderTriggerComplete = "complete"
	OrderTriggerCancel   = "cancel"
)
