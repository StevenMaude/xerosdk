package accounting

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/quickaco/xerosdk/helpers"
)

const (
	accountsURL = "https://api.xero.com/api.xro/2.0/Accounts"
)

//Account represents individual accounts in a Xero organisation
type Account struct {

	// Customer defined alpha numeric account code e.g 200 or SALES (max length = 10)
	Code string `json:"Code,omitempty"`

	// Name of account (max length = 150)
	Name string `json:"Name,omitempty"`

	// See Account Types
	Type string `json:"Type,omitempty"`

	// For bank accounts only (Account Type BANK)
	BankAccountNumber string `json:"BankAccountNumber,omitempty"`

	// Accounts with a status of ACTIVE can be updated to ARCHIVED. See Account Status Codes
	Status string `json:"Status,omitempty"`

	// Description of the Account. Valid for all types of accounts except bank accounts (max length = 4000)
	Description string `json:"Description,omitempty"`

	// For bank accounts only. See Bank Account types
	BankAccountType string `json:"BankAccountType,omitempty"`

	// For bank accounts only
	CurrencyCode string `json:"CurrencyCode,omitempty"`

	// See Tax Types
	TaxType string `json:"TaxType,omitempty"`

	// Boolean – describes whether account can have payments applied to it
	EnablePaymentsToAccount bool `json:"EnablePaymentsToAccount,omitempty"`

	// Boolean – describes whether account code is available for use with expense claims
	ShowInExpenseClaims bool `json:"ShowInExpenseClaims,omitempty"`

	// The Xero identifier for an account – specified as a string following the endpoint name e.g. /297c2dc5-cc47-4afd-8ec8-74990b8761e9
	AccountID string `json:"AccountID,omitempty"`

	// See Account Class Types
	Class string `json:"Class,omitempty"`

	// If this is a system account then this element is returned. See System Account types. Note that non-system accounts may have this element set as either “” or null.
	SystemAccount string `json:"SystemAccount,omitempty"`

	// Shown if set
	ReportingCode string `json:"ReportingCode,omitempty"`

	// Shown if set
	ReportingCodeName string `json:"ReportingCodeName,omitempty"`

	// boolean to indicate if an account has an attachment (read only)
	HasAttachments bool `json:"HasAttachments,omitempty"`

	// Last modified date UTC format
	UpdatedDateUTC string `json:"UpdatedDateUTC,omitempty"`
}

//Accounts contains a collection of Accounts
type Accounts struct {
	Accounts []Account `json:"Accounts,omitempty"`
}

//The Xero API returns Dates based on the .Net JSON date format available at the time of development
//We need to convert these to a more usable format - RFC3339 for consistency with what the API expects to recieve
func (a *Accounts) convertDates() error {
	var err error
	for n := len(a.Accounts) - 1; n >= 0; n-- {
		a.Accounts[n].UpdatedDateUTC, err = helpers.DotNetJSONTimeToRFC3339(a.Accounts[n].UpdatedDateUTC, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func unmarshalAccount(accountResponseBytes []byte) (*Accounts, error) {
	var accountResponse *Accounts
	err := json.Unmarshal(accountResponseBytes, &accountResponse)
	if err != nil {
		return nil, err
	}

	err = accountResponse.convertDates()
	if err != nil {
		return nil, err
	}

	return accountResponse, err
}

//FindAccountsModifiedSince will get all accounts modified after a specified date.
//additional querystringParameters such as where and order can be added as a map
func FindAccountsModifiedSince(cl *http.Client, modifiedSince time.Time, queryParameters map[string]string) (*Accounts, error) {
	additionalHeaders := map[string]string{}
	additionalHeaders["If-Modified-Since"] = modifiedSince.Format(time.RFC3339)

	accountResponseBytes, err := helpers.Find(cl, accountsURL, additionalHeaders, queryParameters)
	if err != nil {
		return nil, err
	}

	return unmarshalAccount(accountResponseBytes)
}

//FindAccounts will get all accounts. These account will not have details like line items.
//additional querystringParameters such as where and order can be added as a map
func FindAccounts(cl *http.Client, queryParameters map[string]string) (*Accounts, error) {
	accountResponseBytes, err := helpers.Find(cl, accountsURL, nil, queryParameters)
	if err != nil {
		return nil, err
	}

	return unmarshalAccount(accountResponseBytes)
}

//FindAccount will get a single account - accountID must be a GUID for an account
func FindAccount(cl *http.Client, accountID uuid.UUID) (*Account, error) {
	accountResponseBytes, err := helpers.Find(cl, accountsURL+"/"+accountID.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	a, err := unmarshalAccount(accountResponseBytes)
	if err != nil {
		return nil, err
	}
	if len(a.Accounts) > 0 {
		return &a.Accounts[0], nil
	}
	return nil, nil
}

// RemoveAccount will get a single account - accountID must be a GUID for an account
func RemoveAccount(cl *http.Client, accountID uuid.UUID) (*Accounts, error) {
	accountResponseBytes, err := helpers.Remove(cl, accountsURL+"/"+accountID.String())
	if err != nil {
		return nil, err
	}

	return unmarshalAccount(accountResponseBytes)
}

// Create will create accounts given an Accounts struct
func (a *Accounts) Create(cl *http.Client) (*Accounts, error) {
	buf, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	accountResponseBytes, err := helpers.Create(cl, accountsURL, buf)
	if err != nil {
		return nil, err
	}

	return unmarshalAccount(accountResponseBytes)
}

// Update will update an account given an Accounts struct
// This will only handle single account - you cannot update multiple accounts in a single call
func (a *Account) Update(cl *http.Client) (*Accounts, error) {
	acc := Accounts{
		Accounts: []Account{*a},
	}
	buf, err := json.Marshal(acc)
	if err != nil {
		return nil, err
	}
	accountResponseBytes, err := helpers.Update(cl, accountsURL+"/"+a.AccountID, buf)
	if err != nil {
		return nil, err
	}

	return unmarshalAccount(accountResponseBytes)
}
