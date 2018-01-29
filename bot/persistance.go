package bot

type Persistance interface {
	Get(tableName string, key string) (value []byte, err error)
	Set(tableName string, key string, value []byte) (err error)
	List(tableName string) (keys []string, err error)
	ListPrefix(tableName string, prefix string) (keys []string, err error)
	Close() (err error)
}

type NoneDB struct{}

func (*NoneDB) Get(_, _ string) ([]byte, error) {
	return nil, nil
}

func (*NoneDB) Set(_, _ string, _ []byte) error {
	return nil
}

func (*NoneDB) List(_ string) ([]string, error) {
	return nil, nil
}

func (*NoneDB) ListPrefix(_, _ string) ([]string, error) {
	return nil, nil
}

func (*NoneDB) Close() error {
	return nil
}
