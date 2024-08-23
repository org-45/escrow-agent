-- Create the escrows table
CREATE TABLE IF NOT EXISTS escrows (
    id VARCHAR(50) PRIMARY KEY,
    buyer_id VARCHAR(50),
    seller_id VARCHAR(50),
    amount NUMERIC(10, 2),
    status VARCHAR(20),
    created_at TIMESTAMP,
    released_at TIMESTAMP,
    disputed_at TIMESTAMP,
    description TEXT
);
