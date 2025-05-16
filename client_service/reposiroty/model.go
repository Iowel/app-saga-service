package repository

type Order struct {
	OrderID    int64
	UserID     int64
	Timestamp  int64
	ProductSKU int64
	Status     string
	Reason     string
}
