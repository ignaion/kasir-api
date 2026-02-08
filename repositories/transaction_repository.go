package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"kasir-api/models"
	"strings"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (repo *TransactionRepository) CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error) {
	var (
		res *models.Transaction
	)

	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// inisialisasi subtotal -> jumlah total transaksi keseluruhan
	totalAmount := 0
	// inisialisasi modeling transactionDetails -> nanti kita insert ke db
	details := make([]models.TransactionDetail, 0)
	// loop setiap item
	for _, item := range items {
		var productName string
		var productID, price, stock int
		// get product dapet pricing
		err := tx.QueryRow("SELECT id, name, price, stock FROM product WHERE id=$1", item.ProductID).Scan(&productID, &productName, &price, &stock)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product id %d not found", item.ProductID)
		}

		if err != nil {
			return nil, err
		}

		subtotal := item.Quantity * price
		totalAmount += subtotal

		_, err = tx.Exec("UPDATE product SET stock = stock - $1 WHERE id = $2", item.Quantity, productID)
		if err != nil {
			return nil, err
		}

		// item nya dimasukkin ke transactionDetails
		details = append(details, models.TransactionDetail{
			ProductID:   productID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	// insert transaction
	var transactionID int
	err = tx.QueryRow("INSERT INTO transactions (total_amount) VALUES ($1) RETURNING ID", totalAmount).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	// insert transaction details (bulk)
	if len(details) > 0 {
		for i := range details {
			details[i].TransactionID = transactionID
		}

		var (
			sb   strings.Builder
			args []any
		)

		sb.WriteString("INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES ")

		// total kolom per row
		const cols = 4
		for i, d := range details {
			if i > 0 {
				sb.WriteString(",")
			}
			base := i*cols + 1
			// ($base, $base+1, $base+2, $base+3)
			sb.WriteString(fmt.Sprintf("($%d,$%d,$%d,$%d)", base, base+1, base+2, base+3))

			args = append(args,
				transactionID,
				d.ProductID,
				d.Quantity,
				d.Subtotal,
			)
		}

		if _, err := tx.Exec(sb.String(), args...); err != nil {
			return nil, fmt.Errorf("bulk insert transaction_details: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	res = &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
		Details:     details,
	}

	return res, nil
}

func (repo *TransactionRepository) GetSummaryToday(ctx context.Context) (*models.SummaryToday, error) {
	var (
		totalRevenue   sql.NullInt64
		totalTransaksi int
	)

	err := repo.db.QueryRowContext(ctx, `
        SELECT
            COALESCE(SUM(t.total_amount), 0) AS total_revenue,
            COUNT(*) AS total_transaksi
        FROM transactions t
        WHERE t.created_at >= CURRENT_DATE
          AND t.created_at < CURRENT_DATE + INTERVAL '1 day'
    `).Scan(&totalRevenue, &totalTransaksi)
	if err != nil {
		return nil, fmt.Errorf("query summary today: %w", err)
	}

	var (
		bestName sql.NullString
		bestQty  sql.NullInt64
	)

	err = repo.db.QueryRowContext(ctx, `
        SELECT p.name AS nama, SUM(td.quantity) AS qty_terjual
        FROM transaction_details td
        JOIN transactions t ON t.id = td.transaction_id
        JOIN product p ON p.id = td.product_id
        WHERE t.created_at >= CURRENT_DATE
          AND t.created_at < CURRENT_DATE + INTERVAL '1 day'
        GROUP BY p.name
        ORDER BY qty_terjual DESC, p.name ASC
        LIMIT 1
    `).Scan(&bestName, &bestQty)

	produkTerlaris := models.Product{}
	if err == sql.ErrNoRows {
		produkTerlaris = models.Product{
			Name: "", // bisa dibiarkan kosong
		}
		bestQty = sql.NullInt64{Int64: 0, Valid: true}
	} else if err != nil {
		return nil, fmt.Errorf("query produk terlaris today: %w", err)
	} else {
		produkTerlaris = models.Product{
			Name: bestName.String,
		}
	}

	return &models.SummaryToday{
		TotalRevenue:   int(totalRevenue.Int64),
		TotalTransaksi: totalTransaksi,
		ProdukTerlaris: produkTerlaris,
	}, nil
}

func (repo *TransactionRepository) GetBestSellerToday(ctx context.Context) (string, int, error) {
	var (
		name sql.NullString
		qty  sql.NullInt64
	)
	err := repo.db.QueryRowContext(ctx, `
			SELECT p.name AS nama, SUM(td.quantity) AS qty_terjual
			FROM transaction_details td
			JOIN transactions t ON t.id = td.transaction_id
			JOIN product p ON p.id = td.product_id -- ganti jadi "products" jika skema kamu jamak
			WHERE t.created_at >= CURRENT_DATE
			AND t.created_at < CURRENT_DATE + INTERVAL '1 day'
			GROUP BY p.name
			ORDER BY qty_terjual DESC, p.name ASC
			LIMIT 1;
    `).Scan(&name, &qty)
	if err == sql.ErrNoRows {
		return "", 0, nil
	}
	if err != nil {
		return "", 0, err
	}
	return name.String, int(qty.Int64), nil
}
