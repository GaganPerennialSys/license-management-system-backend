package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const BaseURL = "http://localhost:8080"

// Client represents the API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	JWTToken   string
	APIKey     string
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    uint   `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	} `json:"user"`
}

// SDKLoginResponse represents SDK login response
type SDKLoginResponse struct {
	APIKey string `json:"api_key"`
	User   struct {
		ID    uint   `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	} `json:"user"`
}

// SubscriptionRequest represents subscription request
type SubscriptionRequest struct {
	PackSKU string `json:"pack_sku"`
}

// Subscription represents a subscription
type Subscription struct {
	ID         uint   `json:"id"`
	CustomerID uint   `json:"customer_id"`
	PackID     uint   `json:"pack_id"`
	Status     string `json:"status"`
	Pack       struct {
		ID             uint    `json:"id"`
		Name           string  `json:"name"`
		Description    string  `json:"description"`
		SKU            string  `json:"sku"`
		Price          float64 `json:"price"`
		ValidityMonths int     `json:"validity_months"`
	} `json:"pack"`
	RequestedAt   string `json:"requested_at"`
	ApprovedAt    string `json:"approved_at"`
	AssignedAt    string `json:"assigned_at"`
	ExpiresAt     string `json:"expires_at"`
	DeactivatedAt string `json:"deactivated_at"`
}

// AdminLogin authenticates as admin
func (c *Client) AdminLogin(email, password string) error {
	req := LoginRequest{Email: email, Password: password}
	
	resp, err := c.makeRequest("POST", "/api/admin/login", req, false, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return err
	}

	c.JWTToken = loginResp.Token
	fmt.Printf("✅ Admin login successful: %s (%s)\n", loginResp.User.Email, loginResp.User.Role)
	return nil
}

// CustomerLogin authenticates as customer
func (c *Client) CustomerLogin(email, password string) error {
	req := LoginRequest{Email: email, Password: password}
	
	resp, err := c.makeRequest("POST", "/api/customer/login", req, false, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return err
	}

	c.JWTToken = loginResp.Token
	fmt.Printf("✅ Customer login successful: %s (%s)\n", loginResp.User.Email, loginResp.User.Role)
	return nil
}

// CustomerSignup registers a new customer
func (c *Client) CustomerSignup(email, password, name, phone string) error {
	req := map[string]string{
		"email":    email,
		"password": password,
		"name":     name,
		"phone":    phone,
	}
	
	resp, err := c.makeRequest("POST", "/api/customer/signup", req, false, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return err
	}

	c.JWTToken = loginResp.Token
	fmt.Printf("✅ Customer signup successful: %s (%s)\n", loginResp.User.Email, loginResp.User.Role)
	return nil
}

// SDKLogin authenticates for SDK access
func (c *Client) SDKLogin(email, password string) error {
	req := LoginRequest{Email: email, Password: password}
	
	resp, err := c.makeRequest("POST", "/sdk/auth/login", req, false, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var loginResp SDKLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return err
	}

	c.APIKey = loginResp.APIKey
	fmt.Printf("✅ SDK login successful: %s (%s)\n", loginResp.User.Email, loginResp.User.Role)
	fmt.Printf("🔑 API Key: %s\n", loginResp.APIKey)
	return nil
}

// GetCurrentSubscription gets current customer subscription
func (c *Client) GetCurrentSubscription() (*Subscription, error) {
	resp, err := c.makeRequest("GET", "/api/v1/customer/subscription", nil, true, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool         `json:"success"`
		Data    Subscription `json:"data"`
		Message string       `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result.Data, nil
}

// RequestSubscription requests a new subscription
func (c *Client) RequestSubscription(packSKU string) (*Subscription, error) {
	req := SubscriptionRequest{PackSKU: packSKU}
	
	resp, err := c.makeRequest("POST", "/api/v1/customer/subscription/request", req, true, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool         `json:"success"`
		Data    Subscription `json:"data"`
		Message string       `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result.Data, nil
}

// SDKGetCurrentSubscription gets current subscription via SDK
func (c *Client) SDKGetCurrentSubscription() (*Subscription, error) {
	resp, err := c.makeRequest("GET", "/sdk/v1/subscription", nil, false, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool         `json:"success"`
		Data    Subscription `json:"data"`
		Message string       `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result.Data, nil
}

// makeRequest makes an HTTP request
func (c *Client) makeRequest(method, path string, body interface{}, useJWT, useAPIKey bool) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	
	if useJWT && c.JWTToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.JWTToken)
	}
	
	if useAPIKey && c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}

	return c.HTTPClient.Do(req)
}

func main() {
	fmt.Println("🚀 License Management System API Client Example")
	fmt.Println("===============================================")

	client := NewClient(BaseURL)

	// Example 1: Admin login
	fmt.Println("\n1. Admin Login")
	if err := client.AdminLogin("admin@example.com", "admin123"); err != nil {
		fmt.Printf("❌ Admin login failed: %v\n", err)
	}

	// Example 2: Customer signup
	fmt.Println("\n2. Customer Signup")
	if err := client.CustomerSignup("customer@example.com", "password123", "John Doe", "+1234567890"); err != nil {
		fmt.Printf("❌ Customer signup failed: %v\n", err)
	}

	// Example 3: Customer login
	fmt.Println("\n3. Customer Login")
	if err := client.CustomerLogin("customer@example.com", "password123"); err != nil {
		fmt.Printf("❌ Customer login failed: %v\n", err)
	}

	// Example 4: Request subscription
	fmt.Println("\n4. Request Subscription")
	subscription, err := client.RequestSubscription("basic-plan")
	if err != nil {
		fmt.Printf("❌ Subscription request failed: %v\n", err)
	} else {
		fmt.Printf("✅ Subscription requested: %s (Status: %s)\n", subscription.Pack.Name, subscription.Status)
	}

	// Example 5: Get current subscription
	fmt.Println("\n5. Get Current Subscription")
	currentSub, err := client.GetCurrentSubscription()
	if err != nil {
		fmt.Printf("❌ Get subscription failed: %v\n", err)
	} else {
		fmt.Printf("✅ Current subscription: %s (Status: %s)\n", currentSub.Pack.Name, currentSub.Status)
	}

	// Example 6: SDK login
	fmt.Println("\n6. SDK Login")
	if err := client.SDKLogin("customer@example.com", "password123"); err != nil {
		fmt.Printf("❌ SDK login failed: %v\n", err)
	}

	// Example 7: SDK get subscription
	fmt.Println("\n7. SDK Get Current Subscription")
	sdkSub, err := client.SDKGetCurrentSubscription()
	if err != nil {
		fmt.Printf("❌ SDK get subscription failed: %v\n", err)
	} else {
		fmt.Printf("✅ SDK subscription: %s (Status: %s)\n", sdkSub.Pack.Name, sdkSub.Status)
	}

	fmt.Println("\n🎉 API Client Example Completed!")
	fmt.Println("\nTo run this example:")
	fmt.Println("1. Start the server: go run main.go")
	fmt.Println("2. Run this client: go run examples/client.go")
}
