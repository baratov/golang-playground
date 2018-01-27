package store

type Store struct {
	items map[string]interface{}
}

func NewStore() Store {
	return Store{
		items: make(map[string]interface{}),
	}
}

func (store Store) Get(key string) (interface{}) {
	return store.items[key]
}

func (store Store) Set(key string, value interface{}) {
	store.items[key] = value
}

func (store Store) Update(key string, value interface{}) {
	store.items[key] = value
}

func (store Store) Delete(key string) {
	delete(store.items, key)
}