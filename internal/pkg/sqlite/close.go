package sqlite

func (s sqliteStore) Close() error {
	if s.Database != nil {
		return s.Database.Close()
	}
	return nil
}
