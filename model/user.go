package model

type User struct {
	ID       int64
	Password string
	Email    string
	Role     string
}

type LoginAttept struct {
	Email    string
	Password string
}
