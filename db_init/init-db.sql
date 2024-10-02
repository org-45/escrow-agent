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

-- TODO
-- CREATE TABLE transactions (
--     transaction_id SERIAL PRIMARY KEY,
--     buyer_id INT REFERENCES users(user_id),
--     seller_id INT REFERENCES users(user_id),
--     amount DECIMAL(10, 2) NOT NULL,
--     status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'deposited', 'in_progress', 'completed', 'cancelled')),
--     created_at TIMESTAMPTZ DEFAULT NOW(),
--     updated_at TIMESTAMPTZ DEFAULT NOW()
-- );



CREATE TABLE IF NOT EXISTS users(
	id SERIAL PRIMARY KEY,
	username VARCHAR(50) UNIQUE NOT NULL,
	password_hash VARCHAR(255) NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- TODO
-- CREATE TABLE users (
--     user_id SERIAL PRIMARY KEY,
--     email VARCHAR(100) UNIQUE NOT NULL,
--     password_hash TEXT NOT NULL,
--     role VARCHAR(20) NOT NULL CHECK (role IN ('buyer', 'seller', 'admin')),
--     created_at TIMESTAMPTZ DEFAULT NOW()
-- );

CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    last_name VARCHAR(100) NOT NULL,
    line1 VARCHAR(255) NOT NULL,
    line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL,
    post_code VARCHAR(20) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- CREATE TABLE escrow_accounts (
--     escrow_id SERIAL PRIMARY KEY,
--     transaction_id INT REFERENCES transactions(transaction_id),
--     escrowed_amount DECIMAL(10, 2) NOT NULL,
--     status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'held', 'released', 'cancelled')),
--     created_at TIMESTAMPTZ DEFAULT NOW()
-- );


-- CREATE TABLE transaction_logs (
--     log_id SERIAL PRIMARY KEY,
--     transaction_id INT REFERENCES transactions(transaction_id),
--     event_type VARCHAR(50) NOT NULL,
--     event_details TEXT,
--     created_at TIMESTAMPTZ DEFAULT NOW()
-- );


