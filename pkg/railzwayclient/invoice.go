package railzwayclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Invoice struct {
	ID            string `json:"id"`
	InvoiceNumber string `json:"invoice_number"`
	CustomerID    string `json:"customer_id"`
	Status        string `json:"status"`
	Total         int64  `json:"total"`
	Currency      string `json:"currency"`
	DueDate       string `json:"due_date"`
	FinalizedAt   string `json:"finalized_at"`
}

type RenderInvoiceResponse struct {
	HTML string `json:"html"`
	PDF  string `json:"pdf"`
}

type ListInvoicesParams struct {
	Status        string
	InvoiceNumber string
	CustomerID    string
	CreatedFrom   string
	CreatedTo     string
	DueFrom       string
	DueTo         string
	FinalizedFrom string
	FinalizedTo   string
	TotalMin      int64
	TotalMax      int64
	PageToken     string
	PageSize      int
}

// ListInvoices lists invoices with optional filtering
func (c *Client) ListInvoices(ctx context.Context, params ListInvoicesParams) ([]Invoice, error) {
	path := "/api/invoices"

	// Build query string
	query := buildListInvoicesQuery(params)
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var invoices []Invoice
	err := c.doRequest(ctx, http.MethodGet, path, nil, &invoices)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	return invoices, nil
}

// GetInvoice retrieves an invoice by ID
func (c *Client) GetInvoice(ctx context.Context, id string) (*Invoice, error) {
	path := fmt.Sprintf("/api/invoices/%s", id)
	var invoice Invoice
	err := c.doRequest(ctx, http.MethodGet, path, nil, &invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return &invoice, nil
}

// RenderInvoice renders an invoice as HTML/PDF
func (c *Client) RenderInvoice(ctx context.Context, id string) (*RenderInvoiceResponse, error) {
	path := fmt.Sprintf("/api/invoices/%s/render", id)
	var response RenderInvoiceResponse
	err := c.doRequest(ctx, http.MethodGet, path, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to render invoice: %w", err)
	}
	return &response, nil
}

func buildListInvoicesQuery(params ListInvoicesParams) url.Values {
	query := url.Values{}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.InvoiceNumber != "" {
		query.Set("invoice_number", params.InvoiceNumber)
	}
	if params.CustomerID != "" {
		query.Set("customer_id", params.CustomerID)
	}
	if params.CreatedFrom != "" {
		query.Set("created_from", params.CreatedFrom)
	}
	if params.CreatedTo != "" {
		query.Set("created_to", params.CreatedTo)
	}
	if params.DueFrom != "" {
		query.Set("due_from", params.DueFrom)
	}
	if params.DueTo != "" {
		query.Set("due_to", params.DueTo)
	}
	if params.FinalizedFrom != "" {
		query.Set("finalized_from", params.FinalizedFrom)
	}
	if params.FinalizedTo != "" {
		query.Set("finalized_to", params.FinalizedTo)
	}
	if params.TotalMin > 0 {
		query.Set("total_min", strconv.FormatInt(params.TotalMin, 10))
	}
	if params.TotalMax > 0 {
		query.Set("total_max", strconv.FormatInt(params.TotalMax, 10))
	}
	if params.PageToken != "" {
		query.Set("page_token", params.PageToken)
	}
	if params.PageSize > 0 {
		query.Set("page_size", strconv.Itoa(params.PageSize))
	}
	return query
}
