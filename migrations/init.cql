CREATE KEYSPACE IF NOT EXISTS shortener
    WITH replication = {'class': 'NetworkTopologyStrategy', 'replication_factor': 3};

USE shortener;

CREATE TABLE IF NOT EXISTS urls (
    code       TEXT,
    long_url   TEXT,
    is_active  BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    PRIMARY KEY (code)
);

CREATE TABLE IF NOT EXISTS url_clicks (
    code        TEXT PRIMARY KEY,
    click_count COUNTER
);
