package domain

type ICurrencyUsecase interface {
	GetConvertedCurrency(from, to, data string, amount float64) (float64, error)
	GetExchangeRate(from, to, date string) (float64, error)
}

type ICurrencyRepository interface {
	GetRate(from, to, date string) (float64, error)
	SaveRate(from, to, date string, rate float64) error
}
