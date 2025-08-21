package postgres

import "context"

type ServiceStore struct {
	storage *Store
}

func (s *ServiceStore) Execute(ctx context.Context, query string, args ...interface{}) (result interface{}, err error) {
	conn, dbErr := s.storage.Database()
	if dbErr != nil {
		return nil, dbErr
	}

	result, err = conn.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *ServiceStore) Array(ctx context.Context, query string, args ...interface{}) ([]interface{}, error) {
	conn, err := s.storage.conn.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []interface{}
	cols := rows.FieldDescriptions()

	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, col := range cols {
			name := string(col.Name)
			rowMap[name] = values[i]
		}
		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func NewServiceStore(storage *Store) *ServiceStore {
	return &ServiceStore{storage: storage}
}
