-- users
-- transactions
    -- enforce_buyer_seller_roles
    -- trigger : transactions_role_check       
-- escrow_accounts
-- transaction_logs
-- files

CREATE TABLE IF NOT EXISTS users(
	user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	username VARCHAR(50) UNIQUE NOT NULL,
	password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('buyer', 'seller', 'admin')),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buyer_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    seller_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0), -- Ensure positive amount
    escrow_status VARCHAR(20) NOT NULL CHECK (escrow_status IN ('pending', 'funded', 'released', 'cancelled')),
    transaction_status VARCHAR(20) NOT NULL CHECK (transaction_status IN ('pending', 'deposited', 'in_progress', 'completed', 'cancelled')),
    dispute_id UUID UNIQUE REFERENCES disputes(dispute_id) ON DELETE SET NULL,
    payment_id UUUID UNIQUE REFERENCES payments(payment_id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
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

CREATE TABLE escrow_accounts (
    escrow_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUUID UNIQUE REFERENCES transactions(transaction_id) ON DELETE CASCADE,
    escrowed_amount DECIMAL(10, 2) NOT NULL CHECK (escrowed_amount > 0),
    escrow_status VARCHAR(20) NOT NULL CHECK (escrow_status IN ('pending', 'held', 'released', 'cancelled')),
    payment_id UUID UNIQUE REFERENCES payments(payment_id) ON DELETE SET NULL,  -- Links escrow to payments
    funded_at TIMESTAMPTZ DEFAULT NOW(),
    released_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    expiry_date TIMESTAMPTZ  -- If the transaction is not confirmed, escrow expires
);


CREATE TABLE transaction_logs (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES transactions(transaction_id),
    event_type VARCHAR(50) NOT NULL,
    event_details TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE files(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES transactions(transaction_id),
	file_name TEXT NOT NULL,
    file_path TEXT NOT NULL, --  "transactions/{transactionID}/{filename}"
	uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)


--payments

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Define ENUM Types
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_method') THEN
        CREATE TYPE payment_method AS ENUM ('credit_card', 'bank_transfer', 'crypto');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_status') THEN
        CREATE TYPE payment_status AS ENUM ('pending', 'completed', 'failed', 'refunded');
    END IF;
END $$;

-- Create Secure Payments Table
CREATE TABLE IF NOT EXISTS payments (
    payment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES transactions(transaction_id) ON DELETE CASCADE,
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    method payment_method NOT NULL,
    status payment_status DEFAULT 'pending',
    encrypted_details BYTEA NOT NULL, -- Secure encrypted payment details
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for Better Performance
CREATE INDEX IF NOT EXISTS idx_payment_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payment_method ON payments(method);


--disputes

CREATE TABLE IF NOT EXISTS disputes (
    dispute_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID UNIQUE REFERENCES transactions(transaction_id) ON DELETE CASCADE,
    raised_by UUID REFERENCES users(user_id) ON DELETE CASCADE,  -- Buyer or seller who raised the dispute
    reason TEXT NOT NULL,  -- Description of the dispute
    status VARCHAR(20) NOT NULL CHECK (status IN ('open', 'resolved', 'rejected')),
    resolution TEXT,  -- Admin decision
    resolved_by UUID REFERENCES users(user_id),  -- Admin who resolved the dispute
    created_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);
