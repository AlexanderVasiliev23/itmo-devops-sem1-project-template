package dataprovider

import (
	"context"
	"database/sql"
	"project_sem/internal/model"
)

type PriceProvider struct {
	db *sql.DB
}

func NewPriceProvider(db *sql.DB) *PriceProvider {
	return &PriceProvider{
		db: db,
	}
}

func (p *PriceProvider) Insert(ctx context.Context, price model.PriceModel) error {
	insertQuery := `
  INSERT INTO prices (id, name, category, price, create_date)
  VALUES ($1, $2, $3, $4, $5)
 `

	_, err := p.db.ExecContext(ctx, insertQuery,
		price.ID,
		price.Name,
		price.Category,
		price.Price,
		price.CreateDate,
	)

	return err
}

func (p *PriceProvider) List(ctx context.Context) (model.PriceModels, error) {
	query := "SELECT id, name, category, price, create_date FROM prices"

	rows, err := p.db.QueryContext(ctx, query)
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
