package dataprovider

import (
	"context"
	"fmt"
	"project_sem/internal/model"
	"project_sem/internal/utils/postgres"
	"strings"
)

type PriceProvider struct {
	transactionProvider *postgres.TransactionProvider
}

func NewPriceProvider(transactionProvider *postgres.TransactionProvider) *PriceProvider {
	return &PriceProvider{
		transactionProvider: transactionProvider,
	}
}

func (p *PriceProvider) InsertBatch(ctx context.Context, prices model.PriceModels) error {
	tx, err := p.transactionProvider.GetDBTransaction(ctx)
	if err != nil {
		return err
	}

	var (
		values       []interface{}
		placeholders []string
	)

	for i, price := range prices {
		numBase := i * 4

		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d)", numBase+1, numBase+2, numBase+3, numBase+4))
		values = append(values, price.Name, price.Category, price.Price, price.CreateDate)
	}

	insertQuery := `
       INSERT INTO prices (name, category, price, create_date)
       VALUES` + strings.Join(placeholders, ", ")

	_, err = tx.ExecContext(ctx, insertQuery, values...)
	if err != nil {
		return err
	}

	return nil
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
