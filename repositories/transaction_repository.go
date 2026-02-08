package repositories

import (
	"database/sql"
	"fmt"
	"kasir-api/models"
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

	// insert transaction details
	for i, detail := range details {
		details[i].TransactionID = transactionID
		_, err := tx.Exec("INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES ($1, $2,$3,$4)", transactionID, detail.ProductID, detail.Quantity, detail.Subtotal)
		if err != nil {
			return nil, err
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
