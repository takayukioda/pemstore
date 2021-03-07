package pemstore

type PemStore interface {
	List() ([]string, error)

	Exists(key string) (bool, error)

	Get(key string, decryption bool) ([]byte, error)

	Store(key string, data []byte, overwrite bool) error

	Remove(key string) error
}
