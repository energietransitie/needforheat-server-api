// Package ports exposes ports for interacting with business logic.
package ports

import "github.com/energietransitie/twomes-backoffice-api/twomes"

// An AccountRepository can load, store and delete accounts.
type AccountRepository interface {
	Find(account twomes.Account) (twomes.Account, error)
	GetAll() ([]twomes.Account, error)
	Create(twomes.Account) (twomes.Account, error)
	Update(twomes.Account) (twomes.Account, error)
	Delete(twomes.Account) error
}

// An AccountService exposes all operations we can perform on a [twomes.Account]
type AccountService interface {
	Create(campaign twomes.Campaign) (twomes.Account, error)
	Activate(id uint, longitude, latitude float32, tzName string) (twomes.Account, error)
	GetByID(id uint) (twomes.Account, error)
	GetCloudFeedAuthStatuses(id uint) ([]twomes.CloudFeedAuthStatus, error)
}
