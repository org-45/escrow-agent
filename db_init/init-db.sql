CREATE EXTENSION IF NOT EXISTS pgcrypto;


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
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX users_username_lower_idx ON users(LOWER(username));
CREATE INDEX users_role_idx ON users(role);
CREATE INDEX users_created_idx ON users USING BRIN(created_at);

-- Create the base table
CREATE TABLE transactions (
    transaction_id UUID NOT NULL DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    buyer_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    seller_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    escrow_status escrow_status NOT NULL,
    transaction_status transaction_status NOT NULL,
    dispute_id UUID,
    payment_id UUID,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (transaction_id, created_at)
);

CREATE INDEX transactions_buyer_idx ON transactions(buyer_id);
CREATE INDEX transactions_seller_idx ON transactions(seller_id);
CREATE INDEX transactions_status_idx ON transactions(escrow_status, transaction_status);
CREATE INDEX transactions_created_idx ON transactions USING BRIN (created_at);

CREATE UNIQUE INDEX transactions_unique_id_idx ON transactions(transaction_id);


CREATE TABLE transaction_logs (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES transactions(transaction_id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    event_details TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX logs_transaction_idx ON transaction_logs(transaction_id);
CREATE INDEX logs_created_idx ON transaction_logs USING BRIN(created_at);

CREATE TABLE files(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES transactions(transaction_id),
	file_name TEXT NOT NULL,
    file_path TEXT NOT NULL, --  "transactions/{transactionID}/{filename}"
    uploaded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX files_transaction_idx ON files(transaction_id);


--payments
CREATE TABLE IF NOT EXISTS payments (
    payment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID, -- FK to transactions, altered later on
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    method payment_method NOT NULL,
    payment_status payment_status DEFAULT 'pending',
    encrypted_details BYTEA NOT NULL, -- Secure encrypted payment details
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
) PARTITION BY HASH (payment_id);

CREATE TABLE payments_p1 PARTITION OF payments FOR VALUES WITH (MODULUS 2, REMAINDER 0);
CREATE TABLE payments_p2 PARTITION OF payments FOR VALUES WITH (MODULUS 2, REMAINDER 1);

CREATE INDEX payments_transaction_idx ON payments(transaction_id);
CREATE INDEX payments_status_idx ON payments(payment_status);
CREATE INDEX payments_created_method_idx ON payments (created_at, method);


--disputes

CREATE TABLE IF NOT EXISTS disputes (
    dispute_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID, -- FK to transactions, altered later on
    raised_by UUID REFERENCES users(user_id) ON DELETE CASCADE,  -- Buyer or seller who raised the dispute
    reason TEXT NOT NULL,  -- Description of the dispute
    dispute_status dispute_status NOT NULL,
    resolution TEXT,  -- Admin decision
    resolved_by UUID REFERENCES users(user_id),  -- Admin who resolved the dispute
    created_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX disputes_status_idx ON disputes(dispute_status);
CREATE INDEX disputes_transaction_idx ON disputes(transaction_id);
CREATE INDEX disputes_created_idx ON disputes USING BRIN(created_at);


CREATE TABLE escrow_accounts (
    escrow_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID UNIQUE REFERENCES transactions(transaction_id) ON DELETE CASCADE,
    escrowed_amount DECIMAL(10, 2) NOT NULL CHECK (escrowed_amount > 0),
    escrow_status escrow_status NOT NULL,
    payment_id UUID UNIQUE REFERENCES payments(payment_id) ON DELETE SET NULL,
    funded_at TIMESTAMPTZ DEFAULT NOW(),
    released_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    expiry_date TIMESTAMPTZ  -- If the transaction is not confirmed, escrow expires
);

CREATE INDEX escrow_transaction_idx ON escrow_accounts(transaction_id);
CREATE INDEX escrow_status_idx ON escrow_accounts(escrow_status);


--foreign keys mapping

ALTER TABLE transactions 
    ADD CONSTRAINT fk_dispute_id FOREIGN KEY (dispute_id) REFERENCES disputes(dispute_id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_payment_id FOREIGN KEY (payment_id) REFERENCES payments(payment_id) ON DELETE SET NULL;

ALTER TABLE payments 
    ADD CONSTRAINT fk_transaction_id FOREIGN KEY (transaction_id) REFERENCES transactions(transaction_id) ON DELETE CASCADE;

ALTER TABLE disputes
    ADD CONSTRAINT fk_transaction_id FOREIGN KEY (transaction_id) REFERENCES transactions(transaction_id) ON DELETE CASCADE;



--create index for disputes and payments on transactions table
CREATE UNIQUE INDEX transactions_dispute_id_idx ON transactions (dispute_id, created_at) 
WHERE dispute_id IS NOT NULL;

CREATE UNIQUE INDEX transactions_payment_id_idx ON transactions (payment_id, created_at) 
WHERE payment_id IS NOT NULL;

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
