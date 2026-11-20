// Package talixpay is the server-side Go SDK for TalixPay.
// Keep ox_ API keys on the server — never send them to the browser.
package talixpay

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func New(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type CreatePaymentInput struct {
	Amount        float64 `json:"amount"`
	ExternalID    string  `json:"external_id,omitempty"`
	CustomerID    string  `json:"customer_id,omitempty"`
	PaymentMethod string  `json:"payment_method,omitempty"`
	CryptoMethod  string  `json:"crypto_method,omitempty"`
	CallbackURL   string  `json:"callback_url,omitempty"`
	RedirectURL   string  `json:"redirect_url,omitempty"`
	CancelURL     string  `json:"cancel_url,omitempty"`
}

type CryptoCoinInfo struct {
	ID                string  `json:"id"`
	Currency          string  `json:"currency"`
	Name              string  `json:"name,omitempty"`
	Network           string  `json:"network,omitempty"`
	NetworkLabel      string  `json:"network_label,omitempty"`
	Available         bool    `json:"available"`
	MinAmountUSD      float64 `json:"min_amount_usd,omitempty"`
	UnavailableReason string  `json:"unavailable_reason,omitempty"`
}

type PaymentCreated struct {
	OrderID       string  `json:"order_id"`
	PaymentURL    string  `json:"payment_url"`
	StatusURL     string  `json:"status_url"`
	PaymentStatus string  `json:"payment_status"`
	PaymentMethod string  `json:"payment_method,omitempty"`
	Currency      string  `json:"currency,omitempty"`
	Channel       string  `json:"channel,omitempty"`
	Amount        float64 `json:"amount"`
	ExternalID    *string `json:"external_id,omitempty"`
	CustomerID    string  `json:"customer_id,omitempty"`
}

type PaymentMethodInfo struct {
	ID        string   `json:"id"`
	Aliases   []string `json:"aliases,omitempty"`
	Label     string   `json:"label"`
	Currency  string   `json:"currency"`
	Channel   string   `json:"channel"`
	Enabled   bool     `json:"enabled"`
	MinAmount float64  `json:"min_amount,omitempty"`
	CoinsURL  string   `json:"coins_url,omitempty"`
}

type CreateCashierSessionInput struct {
	Mode           string   `json:"mode"`
	Amount         float64  `json:"amount"`
	Currency       string   `json:"currency,omitempty"`
	ExternalID     string   `json:"external_id,omitempty"`
	CustomerID     string   `json:"customer_id,omitempty"`
	ReturnURL      string   `json:"return_url,omitempty"`
	CancelURL      string   `json:"cancel_url,omitempty"`
	Theme          string   `json:"theme,omitempty"`
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
}

type CashierSession struct {
	SessionID    string    `json:"session_id"`
	SessionToken string    `json:"session_token"`
	EmbedURL     string    `json:"embed_url"`
	Mode         string    `json:"mode"`
	Status       string    `json:"status"`
	Amount       float64   `json:"amount"`
	Currency     string    `json:"currency"`
	ExpiresAt    time.Time `json:"expires_at"`
	PaymentURL   string    `json:"payment_url,omitempty"`
	PaymentID    string    `json:"payment_id,omitempty"`
	PayoutID     string    `json:"payout_id,omitempty"`
}

func (c *Client) doJSON(ctx context.Context, method, path string, in any, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", c.APIKey)
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 300 {
		var errBody struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(raw, &errBody)
		if errBody.Error != "" {
			return fmt.Errorf("talixpay: %s (HTTP %d)", errBody.Error, res.StatusCode)
		}
		return fmt.Errorf("talixpay: HTTP %d: %s", res.StatusCode, string(raw))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(raw, out)
}

func (c *Client) CreatePayment(ctx context.Context, in CreatePaymentInput) (*PaymentCreated, error) {
	var out PaymentCreated
	if err := c.doJSON(ctx, http.MethodPost, "/v1/payments", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetPayment(ctx context.Context, orderID string) (json.RawMessage, error) {
	var out json.RawMessage
	if err := c.doJSON(ctx, http.MethodGet, "/v1/payments/"+orderID, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListPaymentMethods(ctx context.Context) ([]PaymentMethodInfo, error) {
	var out struct {
		Methods []PaymentMethodInfo `json:"methods"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/payment-methods", nil, &out); err != nil {
		return nil, err
	}
	return out.Methods, nil
}

func (c *Client) ListCryptoMethods(ctx context.Context) ([]CryptoCoinInfo, error) {
	var out struct {
		Methods []CryptoCoinInfo `json:"methods"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/crypto/methods", nil, &out); err != nil {
		return nil, err
	}
	return out.Methods, nil
}

func (c *Client) SelectPaymentMethod(ctx context.Context, orderID, method string) (json.RawMessage, error) {
	var out json.RawMessage
	body := map[string]string{"method": method}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/payments/"+orderID+"/select-method", body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateCashierSession(ctx context.Context, in CreateCashierSessionInput) (*CashierSession, error) {
	var out CashierSession
	if err := c.doJSON(ctx, http.MethodPost, "/v1/cashier/sessions", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetCashierSession(ctx context.Context, sessionID string) (*CashierSession, error) {
	var out CashierSession
	if err := c.doJSON(ctx, http.MethodGet, "/v1/cashier/sessions/"+sessionID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// VerifyWebhook checks header `sha256=<hex>` where hex = sha256(rawBody + secret).
func VerifyWebhook(rawBody []byte, secret, signatureHeader string) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(signatureHeader, prefix) {
		return false
	}
	got := strings.ToLower(strings.TrimPrefix(signatureHeader, prefix))
	sum := sha256.Sum256(append(append([]byte{}, rawBody...), []byte(secret)...))
	return hex.EncodeToString(sum[:]) == got
}
