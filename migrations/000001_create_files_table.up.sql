-- migrations/0001_create_files_table.up.sql
CREATE TABLE files(
	id SERIAL PRIMARY KEY,
  transaction_id INT REFERENCES transactions(transaction_id),
	file_name TEXT NOT NULL,
  file_path TEXT NOT NULL, --  "transactions/{transactionID}/{filename}"
	uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)

