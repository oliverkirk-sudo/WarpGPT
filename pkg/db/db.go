package db

type DB[T any] interface {
	GetClient() T
}
