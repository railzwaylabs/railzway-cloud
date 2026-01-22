package railzwayclient

import (
	"context"
	"fmt"
	"net/http"
)

type Country struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Currency struct {
	Code   string `json:"code"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

type Timezone struct {
	Name   string `json:"name"`
	Offset string `json:"offset"`
}

// ListCountries lists all available countries
func (c *Client) ListCountries(ctx context.Context) ([]Country, error) {
	var countries []Country
	err := c.doRequest(ctx, http.MethodGet, "/api/countries", nil, &countries)
	if err != nil {
		return nil, fmt.Errorf("failed to list countries: %w", err)
	}
	return countries, nil
}

// ListCurrencies lists all available currencies
func (c *Client) ListCurrencies(ctx context.Context) ([]Currency, error) {
	var currencies []Currency
	err := c.doRequest(ctx, http.MethodGet, "/api/currencies", nil, &currencies)
	if err != nil {
		return nil, fmt.Errorf("failed to list currencies: %w", err)
	}
	return currencies, nil
}

// ListTimezones lists all available timezones
func (c *Client) ListTimezones(ctx context.Context) ([]Timezone, error) {
	var timezones []Timezone
	err := c.doRequest(ctx, http.MethodGet, "/api/timezones", nil, &timezones)
	if err != nil {
		return nil, fmt.Errorf("failed to list timezones: %w", err)
	}
	return timezones, nil
}
