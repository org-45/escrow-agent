-- users
-- transactions
-- escrow_accounts
-- transaction_logs

CREATE TABLE IF NOT EXISTS users(
	user_id SERIAL PRIMARY KEY,
	username VARCHAR(50) UNIQUE NOT NULL,
	password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('buyer', 'seller', 'admin')),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    transaction_id SERIAL PRIMARY KEY,
    buyer_id INT REFERENCES users(user_id),
    seller_id INT REFERENCES users(user_id),
    amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'deposited', 'in_progress', 'completed', 'cancelled')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE escrow_accounts (
    escrow_id SERIAL PRIMARY KEY,
    transaction_id INT REFERENCES transactions(transaction_id),
    escrowed_amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'held', 'released', 'cancelled')),
    created_at TIMESTAMPTZ DEFAULT NOW()
);


CREATE TABLE transaction_logs (
    log_id SERIAL PRIMARY KEY,
    transaction_id INT REFERENCES transactions(transaction_id),
    event_type VARCHAR(50) NOT NULL,
    event_details TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);





