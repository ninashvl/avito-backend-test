CREATE TABLE IF NOT EXISTS segments (
    name VARCHAR(255) NOT NULL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);


CREATE TABLE IF NOT EXISTS user_segment (
    user_id INT,
    segment_name VARCHAR(255),
    FOREIGN KEY (segment_name) REFERENCES segments(name) ON UPDATE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE UNIQUE INDEX assigned_user_segments
    ON user_segment(user_id, segment_name)
    WHERE deleted_at IS NULL

