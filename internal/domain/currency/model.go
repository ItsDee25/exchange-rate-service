package domain

type RateKeyRequest struct {
	From string 
	To   string
	Date string
}

type RateKey struct {
	RateKeyRequest
	Rate float64
}