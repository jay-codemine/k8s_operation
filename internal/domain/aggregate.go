package domain

// AggregateRoot 聚合根标记接口
// 聚合根是领域模型的入口点，负责维护其边界内所有实体的不变性。
// 对聚合内实体的所有修改都必须通过聚合根进行。
type AggregateRoot interface {
	// AggregateID 返回聚合根的唯一标识
	AggregateID() int64
}

// DomainEvent 领域事件标记接口
type DomainEvent interface {
	EventName() string
}
