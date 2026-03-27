package store

type Storer interface {
	Update(userID string, eventType string, value int64) error
	GetStats(userID string) map[string]float64
}
