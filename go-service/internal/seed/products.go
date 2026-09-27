// Package seed містить початкові дані для in-memory сховища.
package seed

import "tech-store-go/internal/model"

func Products() []model.Product {
	return []model.Product{
		{Name: "iPhone 15 128GB", Category: "Smartphones", Brand: "Apple", Price: 32999, Stock: 12,
			Description: "Смартфон Apple з чипом A16 Bionic"},
		{Name: "Galaxy S24", Category: "Smartphones", Brand: "Samsung", Price: 27999, Stock: 8,
			Description: "Флагман Samsung з екраном 6.2\" AMOLED"},
		{Name: "ThinkPad E14", Category: "Laptops", Brand: "Lenovo", Price: 29999, Stock: 5,
			Description: "Ноутбук 14\", Intel Core i5, 16 GB RAM, 512 GB SSD"},
		{Name: "WH-1000XM5", Category: "Headphones", Brand: "Sony", Price: 11999, Stock: 20,
			Description: "Бездротові навушники з шумопригніченням"},
	}
}
