package svc

import "github.com/qmuntal/stateless"

// OrderStateMachine 封装 stateless 状态机
type OrderStateMachine struct {
	sm *stateless.StateMachine
}

func NewOrderStateMachine(currentStatus string, order OrderModel) *OrderStateMachine {
	sm := stateless.NewStateMachine(currentStatus)
	// 初始化状态机
	return &OrderStateMachine{
		sm: sm,
	}
}
