package repository

type CurrencyRepository struct {
	// connections with redis
}

func NewCurrencyRepository() *CurrencyRepository {
	return &CurrencyRepository{}
}

func (r *CurrencyRepository) GetRate(from, to, date string) (float64, error) {
	return 0, nil
}

func (r *CurrencyRepository) SaveRate(from, to, date string, rate float64) error {
	// save the rate to redis
	return nil
}