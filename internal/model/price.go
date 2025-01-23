package model

import "time"

type PriceModel struct {
	ID         int
	Name       string
	Category   string
	Price      float64
	CreateDate time.Time
}

type PriceModels []*PriceModel

func (p PriceModels) TotalPrice() float64 {
	out := 0.0

	for _, model := range p {
		out += model.Price
	}

	return out
}

func (p PriceModels) UniqueCategoriesCount() int {
	set := map[string]struct{}{}
	out := 0

	for _, model := range p {
		if _, ok := set[model.Category]; !ok {
			set[model.Category] = struct{}{}
			out++
		}
	}

	return out
}
