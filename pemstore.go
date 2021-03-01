package pemstore

type PemStore interface {
	List() ([]string, error)

	Exists(key string) (bool, error)

	Download(key string, decryption bool) (string, error)

	Store(key string, data []byte, overwrite bool) error

	Remove(key string) error
}
