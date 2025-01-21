package model

import "time"

type PriceModels []*PriceModel

type PriceModel struct {
	ID         int
	Name       string
	Category   string
	Price      float64
	CreateDate time.Time
}
