package config

type Input struct {
	Port int
	Cost struct {
		OverdueFine float64
		Discount    struct {
			Teacher float64
			Student float64
		}
	}
	Output struct {
		Enabled bool
		File    string
		Format  string
	}
	DB struct {
		DSN                    string
		Database               string
		BooksCollection        string
		PatronsCollection      string
		TransactionsCollection string
	}
}