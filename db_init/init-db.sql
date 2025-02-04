-- users
-- transactions
    -- enforce_buyer_seller_roles
    -- trigger : transactions_role_check       
-- escrow_accounts
-- transaction_logs
-- files

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

CREATE OR REPLACE FUNCTION enforce_buyer_seller_roles()
RETURNS TRIGGER AS $$
BEGIN
    -- Ensure buyer has 'buyer' role
    IF NOT EXISTS (SELECT 1 FROM users WHERE user_id = NEW.buyer_id AND role = 'buyer') THEN
        RAISE EXCEPTION 'Invalid buyer_id: User is not a buyer';
    END IF;

    -- Ensure seller has 'seller' role
    IF NOT EXISTS (SELECT 1 FROM users WHERE user_id = NEW.seller_id AND role = 'seller') THEN
        RAISE EXCEPTION 'Invalid seller_id: User is not a seller';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

--attach trigger to transactions table
CREATE TRIGGER transactions_role_check
BEFORE INSERT OR UPDATE ON transactions
FOR EACH ROW
EXECUTE FUNCTION enforce_buyer_seller_roles();


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

CREATE TABLE files(
	id SERIAL PRIMARY KEY,
    transaction_id INT REFERENCES transactions(transaction_id),
	file_name TEXT NOT NULL,
    file_path TEXT NOT NULL, --  "transactions/{transactionID}/{filename}"
	uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)