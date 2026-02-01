package repositories

import (
	"database/sql"
	"errors"
	"kasir-api/models"
	"log"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (repo *ProductRepository) GetAll() ([]models.Product, error) {
	query := "SELECT id, name, price, stock, category_id FROM product"
	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		var p models.Product
		var catID sql.NullInt64
		err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &catID)
		if err != nil {
			return nil, err
		}
		if catID.Valid {
			v := int(catID.Int64)
			p.CategoryId = &v
		} else {
			p.CategoryId = nil
		}
		products = append(products, p)
	}

	return products, nil
}

// GetAllWithCategory - LEFT JOIN untuk isi field Category di model
func (repo *ProductRepository) GetAllWithCategory() ([]models.Product, error) {
	query := `
        SELECT p.id, p.name, p.price, p.stock,
               p.category_id,
               c.id, c.name, c.description
        FROM product p
        LEFT JOIN category c ON c.id = p.category_id
        ORDER BY p.id;
    `
	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Product, 0)
	for rows.Next() {
		var p models.Product
		var catID sql.NullInt64
		var cID sql.NullInt64
		var cName, cDesc sql.NullString

		if err := rows.Scan(
			&p.ID, &p.Name, &p.Price, &p.Stock,
			&catID,
			&cID, &cName, &cDesc,
		); err != nil {
			return nil, err
		}

		if catID.Valid {
			v := int(catID.Int64)
			p.CategoryId = &v
		}
		// isi Category hanya jika join menemukan data
		if cID.Valid {
			p.Category = &models.Category{
				ID:          int(cID.Int64),
				Name:        cName.String,
				Description: cDesc.String,
			}
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (repo *ProductRepository) Create(product *models.Product) error {
	query := "INSERT INTO product (name, price, stock, category_id) VALUES ($1, $2, $3, $4) RETURNING id"
	var catID interface{}

	if product.CategoryId == nil {
		catID = nil
	} else {
		catID = *product.CategoryId
	}
	err := repo.db.QueryRow(query, product.Name, product.Price, product.Stock, catID).Scan(&product.ID)
	return err
}

// GetByID - ambil produk by ID
func (repo *ProductRepository) GetByID(id int) (*models.Product, error) {
	query := "SELECT id, name, price, stock, category_id FROM product WHERE id = $1"

	var p models.Product
	var catID sql.NullInt64

	err := repo.db.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &catID)
	if err == sql.ErrNoRows {
		return nil, errors.New("produk tidak ditemukan")
	}
	if err != nil {
		return nil, err
	}

	if catID.Valid {
		v := int(catID.Int64)
		p.CategoryId = &v
	}

	return &p, nil
}

// GetByID - ambil produk by ID + object Category (LEFT JOIN)
func (repo *ProductRepository) GetByIdWithCategory(id int) (*models.Product, error) {
	query := `
        SELECT 
            p.id, p.name, p.price, p.stock, p.category_id,
            c.id, c.name, c.description
        FROM product p
        LEFT JOIN category c ON c.id = p.category_id
        WHERE p.id = $1
    `

	var p models.Product
	var catID sql.NullInt64
	var cID sql.NullInt64
	var cName, cDesc sql.NullString

	err := repo.db.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.Price, &p.Stock, &catID,
		&cID, &cName, &cDesc,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("produk tidak ditemukan")
	}
	if err != nil {
		return nil, err
	}

	// set CategoryID (nullable)
	if catID.Valid {
		v := int(catID.Int64)
		p.CategoryId = &v
	} else {
		log.Println("category_id belum diset")
	}

	// isi object Category hanya jika baris kategori ada (bukan NULL)
	if cID.Valid {
		p.Category = &models.Category{
			ID:          int(cID.Int64),
			Name:        cName.String,
			Description: cDesc.String,
		}
	}

	return &p, nil
}

func (repo *ProductRepository) Update(product *models.Product) error {
	query := "UPDATE product SET name = $1, price = $2, stock = $3, category_id = $4 WHERE id = $5"

	var catID interface{}

	if product.CategoryId == nil {
		catID = nil
	} else {
		catID = *product.CategoryId
	}

	result, err := repo.db.Exec(query, product.Name, product.Price, product.Stock, catID, product.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("produk tidak ditemukan")
	}

	return nil
}

func (repo *ProductRepository) Delete(id int) error {
	query := "DELETE FROM product WHERE id = $1"
	result, err := repo.db.Exec(query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("produk tidak ditemukan")
	}

	return err
}
