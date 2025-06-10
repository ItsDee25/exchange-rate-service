package usecase

import "github.com/ItsDee25/exchange-rate-service/internal/domain"

type CurrencyUsecase struct {
	currencyRepo *domain.ICurrencyRepository
}

func NewCurrencyUsecase(r *domain.ICurrencyRepository) *CurrencyUsecase {
	return &CurrencyUsecase{
		currencyRepo: r,
	}
}

func (u *CurrencyUsecase) GetConvertedCurrency(from, to string, amount float64) (float64, error) {
	return 0, nil
}

func (u *CurrencyUsecase) GetExchangeRate(from, to string) (float64, error) {
	return 0, nil
}