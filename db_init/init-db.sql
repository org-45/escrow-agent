CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_partman;  -- for partitioning large tables
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;  -- query performance monitoring


-- Define ENUM Types
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('buyer', 'seller', 'admin');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'escrow_status') THEN
        CREATE TYPE escrow_status AS ENUM ('pending', 'funded', 'released', 'cancelled');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'transaction_status') THEN
        CREATE TYPE transaction_status AS ENUM ('pending', 'deposited', 'in_progress', 'completed', 'cancelled');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_method') THEN
        CREATE TYPE payment_method AS ENUM ('credit_card', 'bank_transfer', 'crypto');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_status') THEN
        CREATE TYPE payment_status AS ENUM ('pending', 'completed', 'failed', 'refunded');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'dispute_status') THEN
        CREATE TYPE dispute_status AS ENUM ('open', 'resolved', 'rejected');
    END IF;
END $$;



CREATE TABLE IF NOT EXISTS users(
	user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	username VARCHAR(50) UNIQUE NOT NULL,
	password_hash VARCHAR(255) NOT NULL,
    role user_role NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP

    -- indexes
    EXCLUDE USING hash (username WITH =)  -- Prevent even case-insensitive duplicates
);

CREATE INDEX users_role_idx ON users(role);
CREATE INDEX users_created_idx ON users USING BRIN(created_at);

CREATE TABLE transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buyer_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    seller_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    escrow_status escrow_status NOT NULL,
    transaction_status transaction_status NOT NULL,
    dispute_id UUID UNIQUE REFERENCES disputes(dispute_id) ON DELETE SET NULL,
    payment_id UUUID UNIQUE REFERENCES payments(payment_id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()

    -- indexes
    INDEX transactions_buyer_idx (buyer_id),
    INDEX transactions_seller_idx (seller_id),
    INDEX transactions_status_idx (escrow_status, transaction_status),
    INDEX transactions_created_idx USING BRIN(created_at)
) PARTITION BY RANGE (created_at);

-- Create default partition
CREATE TABLE transactions_default PARTITION OF transactions DEFAULT;



CREATE TABLE escrow_accounts (
    escrow_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUUID UNIQUE REFERENCES transactions(transaction_id) ON DELETE CASCADE,
    escrowed_amount DECIMAL(10, 2) NOT NULL CHECK (escrowed_amount > 0),
    escrow_status escrow_status NOT NULL,
    payment_id UUID UNIQUE REFERENCES payments(payment_id) ON DELETE SET NULL,
    funded_at TIMESTAMPTZ DEFAULT NOW(),
    released_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    expiry_date TIMESTAMPTZ  -- If the transaction is not confirmed, escrow expires

    -- indexes
    INDEX escrow_transaction_idx (transaction_id),
    INDEX escrow_status_idx (escrow_status)
);


CREATE TABLE transaction_logs (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES transactions(transaction_id),
    event_type VARCHAR(50) NOT NULL,
    event_details TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    INDEX logs_transaction_idx (transaction_id),
    INDEX logs_created_idx USING BRIN(created_at)
);

CREATE TABLE files(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES transactions(transaction_id),
	file_name TEXT NOT NULL,
    file_path TEXT NOT NULL, --  "transactions/{transactionID}/{filename}"
    uploaded_at TIMESTAMPTZ DEFAULT NOW(),

    INDEX files_transaction_idx (transaction_id), 
)


--payments
CREATE TABLE IF NOT EXISTS payments (
    payment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES transactions(transaction_id) ON DELETE CASCADE,
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    method payment_method NOT NULL,
    status payment_status DEFAULT 'pending',
    encrypted_details BYTEA NOT NULL, -- Secure encrypted payment details
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()

    INDEX payments_transaction_idx (transaction_id),
    INDEX payments_status_idx (status),
    INDEX payments_created_method_idx (created_at, method)
) PARTITION BY HASH (transaction_id);




--disputes

CREATE TABLE IF NOT EXISTS disputes (
    dispute_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID UNIQUE REFERENCES transactions(transaction_id) ON DELETE CASCADE,
    raised_by UUID REFERENCES users(user_id) ON DELETE CASCADE,  -- Buyer or seller who raised the dispute
    reason TEXT NOT NULL,  -- Description of the dispute
    dispute_status dispute_status NOT NULL,
    resolution TEXT,  -- Admin decision
    resolved_by UUID REFERENCES users(user_id),  -- Admin who resolved the dispute
    created_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    
    INDEX disputes_status_idx (status),
    INDEX disputes_transaction_idx (transaction_id),
    INDEX disputes_created_idx USING BRIN(created_at)

);


CREATE OR REPLACE FUNCTION enforce_buyer_seller_roles()
RETURNS TRIGGER AS $$
BEGIN
    -- Ensure buyer has 'buyer' role
    IF NOT EXISTS (SELECT 1 FROM users WHERE user_id = NEW.buyer_id AND (role = 'buyer' or role = 'admin')) THEN
        RAISE EXCEPTION 'Invalid buyer_id: User is not a buyer';
    END IF;

    -- Ensure seller has 'seller' role
    IF NOT EXISTS (SELECT 1 FROM users WHERE user_id = NEW.seller_id AND (role = 'seller' or role = 'admin')) THEN
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
