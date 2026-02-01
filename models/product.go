package models

type Product struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Price      int       `json:"price"`
	Stock      int       `json:"stock"`
	CategoryId *int      `json:"category_id,omitempty"`
	Category   *Category `json:"category,omitempty"`
}
