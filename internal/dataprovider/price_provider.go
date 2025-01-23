package dataprovider

import (
	"context"
	"project_sem/internal/model"
	"project_sem/internal/utils/postgres"
)

type PriceProvider struct {
	transactionProvider *postgres.TransactionProvider
}

func NewPriceProvider(transactionProvider *postgres.TransactionProvider) *PriceProvider {
	return &PriceProvider{
		transactionProvider: transactionProvider,
	}
}

func (p *PriceProvider) Insert(ctx context.Context, price model.PriceModel) error {
	tx, err := p.transactionProvider.GetDBTransaction(ctx)
	if err != nil {
		return err
	}

	insertQuery := `
  INSERT INTO prices (name, category, price, create_date)
  VALUES ($1, $2, $3, $4)
 `

	_, err = tx.ExecContext(ctx, insertQuery,
		price.Name,
		price.Category,
		price.Price,
		price.CreateDate,
	)

	return err
}

func (p *PriceProvider) List(ctx context.Context) (model.PriceModels, error) {
	tx, err := p.transactionProvider.GetDBTransaction(ctx)
	if err != nil {
		return model.PriceModels{}, err
	}

	query := "SELECT id, name, category, price, create_date FROM prices"

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []*model.PriceModel

	for rows.Next() {
		var price model.PriceModel
		err := rows.Scan(&price.ID, &price.Name, &price.Category, &price.Price, &price.CreateDate)
		if err != nil {
			return nil, err
		}
		prices = append(prices, &price)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return prices, nil
}
