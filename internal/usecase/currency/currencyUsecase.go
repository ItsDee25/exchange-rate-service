package usecase

import (
	"context"
	"time"

	domain "github.com/ItsDee25/exchange-rate-service/internal/domain/currency"
	"github.com/ItsDee25/exchange-rate-service/pkg/constants"
)

type CurrencyUsecase struct {
	currencyRepo domain.ICurrencyRepository
}

func NewCurrencyUsecase(r domain.ICurrencyRepository) *CurrencyUsecase {
	return &CurrencyUsecase{
		currencyRepo: r,
	}
}

func (u *CurrencyUsecase) GetConvertedCurrency(ctx context.Context, from, to, date string, amount float64) (float64, error) {
	if date == "" {
		date = time.Now().Format(constants.DateLayout)
	}
	exchangeRate, err := u.GetExchangeRate(ctx, from, to, date)
	if err != nil {
		return 0, err
	}
	return amount * exchangeRate, nil
}

func (u *CurrencyUsecase) GetExchangeRate(ctx context.Context, from, to, date string) (float64, error) {
	if date == "" {
		date = time.Now().Format(constants.DateLayout)
	}
	if from == to {
		return 1, nil
	}

	return u.currencyRepo.GetRate(ctx, from, to, date)

}
