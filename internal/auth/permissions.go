package auth

const (
	WriteBooksPermission        = "books:write"
	ReadBooksPermission         = "books:read"
	BorrowBookPermission        = "book:borrow"
	ReturnBookPermission        = "book:return"
	ReadPatronsPermission       = "patrons:read"
	WritePatronsPermission      = "patrons:write"
	ReadPatronPermission        = "patron:read"
	WritePatronPermission       = "patron:write"
	ReadTransactionsPermission  = "transactions:read"
	WriteTransactionsPermission = "transactions:write"
)

var AdminPermissions = []string{WriteBooksPermission, ReadBooksPermission, BorrowBookPermission, ReturnBookPermission,
	ReadPatronsPermission, WritePatronsPermission, ReadPatronPermission, WritePatronPermission,
	ReadTransactionsPermission, WriteTransactionsPermission}
