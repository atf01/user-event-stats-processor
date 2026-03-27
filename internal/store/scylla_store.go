package store

import (
	"time"
	"user-event-stats-processor/internal/config"

	"github.com/gocql/gocql"
)

type ScyllaStore struct {
	session *gocql.Session
}

func NewScyllaStore(cfg config.ScyllaConfig) (*ScyllaStore, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace
	cluster.Timeout = 5 * time.Second
	cluster.PoolConfig.HostSelectionPolicy = gocql.RoundRobinHostPolicy()

	// SECURE STEP: Only add authenticator if User/Pass are provided
	if cfg.User != "" && cfg.Pass != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.User,
			Password: cfg.Pass,
		}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return &ScyllaStore{session: session}, nil
}

// Close gracefully shuts down the ScyllaDB session
func (s *ScyllaStore) Close() {
	if s.session != nil {
		s.session.Close()
	}
}

func (s *ScyllaStore) Update(userID string, eventType string, value int64) error {
	query := `UPDATE user_stats SET val = val + ? WHERE user_id = ? AND event_type = ?`

	// Return the error from the Exec() call
	return s.session.Query(query, value, userID, eventType).Exec()
}

func (s *ScyllaStore) GetStats(userID string) map[string]float64 {
	results := make(map[string]float64)
	var eventType string
	var val int64

	iter := s.session.Query("SELECT event_type, val FROM user_stats WHERE user_id = ?", userID).Iter()
	for iter.Scan(&eventType, &val) {
		results[eventType] = float64(val)
	}
	return results
}
