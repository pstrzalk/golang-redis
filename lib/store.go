package lib

type Store interface {
	Get(key string) (value string, found bool, err error)
	Set(key string, value string) error
	Delete(key string) error
}
