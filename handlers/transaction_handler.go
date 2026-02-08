package handlers

import (
	"context"
	"encoding/json"
	"kasir-api/models"
	"kasir-api/services"
	"net/http"
	"time"
)

type TransactionHandler struct {
	service *services.TransactionService
}

func NewTransactionHandler(service *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

// multiple item apa aja, quantity nya
func (h *TransactionHandler) HandleCheckout(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.Checkout(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TransactionHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req models.CheckoutRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	transaction, err := h.service.Checkout(req.Items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) SummaryToday(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	summary, err := h.service.GetSummaryToday(ctx)
	if err != nil {
		http.Error(w, "failed to get today summary: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type bestProduct struct {
		Nama       string `json:"nama"`
		QtyTerjual int    `json:"qty_terjual"`
	}

	bpNama := ""
	bpQty := 0
	{
		nama, qty, e := h.service.GetBestSellerToday(ctx)
		if e == nil {
			bpNama = nama
			bpQty = qty
		}
	}

	response := map[string]any{
		"total_revenue":   summary.TotalRevenue,
		"total_transaksi": summary.TotalTransaksi,
		"produk_terlaris": bestProduct{
			Nama:       bpNama,
			QtyTerjual: bpQty,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
