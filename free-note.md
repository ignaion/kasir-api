Akarprima@1234


postgresql://postgres.riieosieddltxrcwpaaa:[YOUR-PASSWORD]@aws-1-ap-south-1.pooler.supabase.com:6543/postgres


https://github.com/spf13/viper

Akarprima1234


tolong buatkan aplikai CLI dengan bahasa Go-lang.
fitur:
- bisa input, edit, delete data product
- bisa input, edit, delete data category
- terdapat relasi antara product dan category
- database menggunakan postgres



CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    total_amount INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transaction_details (
    id SERIAL PRIMARY KEY,
    transaction_id INT REFERENCES transactions(id) ON DELETE CASCADE,
    product_id INT REFERENCES products(id),
    quantity INT NOT NULL,
    subtotal INT NOT NULL
);


func (repo *TransactionRepository) summaryToday() (*models.SummaryToday, error) {
	queryTotalTransaksi := `
			SELECT COUNT(*) as total 
			FROM transactions t
			WHERE t.created_at::date = CURRENT_DATE;
	`

	
}