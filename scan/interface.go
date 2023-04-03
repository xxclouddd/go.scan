package scan

type RowsScanner interface {
	Scan(dest ...interface{}) error
	Err() error
	Next() bool
}

type RowScanner interface {
	Scan(dest ...interface{}) error
	Err() error
}
