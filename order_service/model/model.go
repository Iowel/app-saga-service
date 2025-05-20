package model

import "time"

type Product struct {
	SKU   int64
	Price int64
	Cnt   int64
}

type Profile struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Avatar    string    `json:"avatar"`
	About     string    `json:"about"`
	Friends   []int     `json:"friends"`
	Wallet    int       `json:"wallet"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Name      string `json:"name"`
	Token     string
	Role      string    `json:"role"`
	Avatar    string    `json:"avatar"`
	Status    string    `json:"status"`
	Wallet    *int      `json:"wallet"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserCache struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Name      string `json:"name"`
	Token     string
	Role      string    `json:"role"`
	Avatar    string    `json:"avatar"`
	Status    string    `json:"status"`
	Wallet    int       `json:"wallet"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
