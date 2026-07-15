package main

const (
	StatusPending   = "PENDING"
	StatusDelivered = "DELIVERED"
	StatusRead      = "READ"
	StatusDismissed = "DISMISSED"
)

var validTransitions = map[string][]string{
	StatusPending:   {StatusDelivered, StatusDismissed},
	StatusDelivered: {StatusRead, StatusDismissed},
	StatusRead:      {StatusDismissed},
	StatusDismissed: {},
}

func isValidStatus(s string) bool {
	_, ok := validTransitions[s]
	return ok
}

func canTransition(from, to string) bool {
	for _, allowed := range validTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}
