# Ensure the script uses the default credentials you just reverted to
cqlsh -u cassandra -p cassandra <<EOF
CREATE KEYSPACE IF NOT EXISTS stats_keyspace 
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
USE stats_keyspace;
CREATE TABLE IF NOT EXISTS user_stats (
    user_id text,
    event_type text,
    val counter,
    PRIMARY KEY (user_id, event_type)
);
EOF