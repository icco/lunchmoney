package lunchmoney

// PlaidAccountResponse is a list plaid accounts response.
type PlaidAccountResponse struct {
	Error         string          `json:"error"`
	PlaidAccounts []*PlaidAccount `json:"plaid_accounts"`
}

// PlaidAccount is a single LM Plaid account.
type PlaidAccount struct {
	ID                int64     `json:"id"`
	DateLinked        time.Time `json:"date_linked"`
	Name              string    `json:"name"`
	Type              string    `json:"type"`
	Subtype           string    `json:"subtype"`
	Mask              string    `json:"mask"`
	InstitutionName   string    `json:"institution_name"`
	Status            string    `json:"status"`
	LastImport        time.Time `json:"last_import"`
	Balance           string    `json:"balance"`
	Currency          string    `json:"currency"`
	BalanceLastUpdate time.Time `json:"balance_last_update"`
	Limit             int64     `json:"limit"`
}
