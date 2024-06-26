// Package services exposes a services as entrypoints for business logic.
package services

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/energietransitie/needforheat-server-api/internal/helpers"
	"github.com/energietransitie/needforheat-server-api/needforheat/account"
	"github.com/energietransitie/needforheat-server-api/needforheat/authorization"
	"github.com/energietransitie/needforheat-server-api/needforheat/campaign"
	"github.com/energietransitie/needforheat-server-api/needforheat/cloudfeed"
	"github.com/energietransitie/needforheat-server-api/needforheat/cloudfeedstatus"
	"github.com/energietransitie/needforheat-server-api/needforheat/cloudfeedtype"
	"github.com/sirupsen/logrus"
)

var (
	ErrTokenSigningMethodInvalid = errors.New("unexpected signing method")
	ErrTokenInvalid              = errors.New("token is invalid")
)

type AccountService struct {
	repository account.AccountRepository

	// Services used when activating an account.
	authService     *AuthorizationService
	appService      *AppService
	campaignService *CampaignService

	// Services used for getting cloud feed auth statuses.
	dataSourceTypeService *DataSourceTypeService
	cloudFeedService      *CloudFeedService

	// Regular expression used for pattern matching in a provisioning_url_template.
	activationTokenRegex *regexp.Regexp
}

// Create a new AccountService
func NewAccountService(
	repository account.AccountRepository,
	authService *AuthorizationService,
	appService *AppService,
	campaignService *CampaignService,
	cloudFeedService *CloudFeedService,
	dataSourceTypeService *DataSourceTypeService,
) *AccountService {
	activationTokenRegex, err := regexp.Compile(`<account_activation_token>`)
	if err != nil {
		logrus.WithField("error", err).Fatal("account activation token regex did not compile")
	}

	return &AccountService{
		repository:            repository,
		authService:           authService,
		appService:            appService,
		campaignService:       campaignService,
		cloudFeedService:      cloudFeedService,
		dataSourceTypeService: dataSourceTypeService,
		activationTokenRegex:  activationTokenRegex,
	}
}

// Create a new account.
func (s *AccountService) Create(campaign campaign.Campaign) (account.Account, error) {
	campaign, err := s.campaignService.Find(campaign)
	if err != nil {
		return account.Account{}, err
	}

	a := account.MakeAccount(campaign)
	a, err = s.repository.Create(a)

	if err != nil {
		return account.Account{}, err
	}

	a.InvitationToken, err = s.authService.CreateToken(authorization.AccountActivationToken, a.ID, time.Time{})
	if err != nil {
		return account.Account{}, err
	}

	a.InvitationURL = s.activationTokenRegex.ReplaceAllString(campaign.App.ProvisioningURLTemplate, url.PathEscape(a.InvitationToken))

	return a, nil
}

// Activate an account.
func (s *AccountService) Activate(id uint) (account.Account, error) {
	a, err := s.repository.Find(account.Account{ID: id})
	if err != nil {
		return account.Account{}, err
	}

	err = a.Activate()
	if err != nil {
		return a, err
	}

	a, err = s.repository.Update(a)
	if err != nil {
		return account.Account{}, err
	}

	a.AuthorizationToken, err = s.authService.CreateToken(authorization.AccountToken, a.ID, time.Time{})
	if err != nil {
		return account.Account{}, err
	}

	return a, nil
}

// Get an account by ID.
func (s *AccountService) GetByID(id uint) (account.Account, error) {
	return s.repository.Find(account.Account{ID: id})
}

// Get cloud feed auth statuses.
func (s *AccountService) GetCloudFeedAuthStatuses(id uint) ([]cloudfeedstatus.CloudFeedStatus, error) {
	var cloudFeedAuthStatuses []cloudfeedstatus.CloudFeedStatus

	a, err := s.GetByID(id)
	if err != nil {
		return cloudFeedAuthStatuses, err
	}

	var cloudFeedTypes []cloudfeedtype.CloudFeedType
	for _, dataSourceType := range a.Campaign.DataSourceList.Items {
		item, _, err := s.dataSourceTypeService.GetSourceByIDAndTable(dataSourceType.ID, "cloud_feed_type")
		if err != nil {
			fmt.Printf("Error fetching source for ID %d: %v\n", dataSourceType.ID, err)
			continue
		}

		// Assert the retrieved source to the appropriate type (CloudFeedType)
		cloudFeedType, ok := item.(cloudfeedtype.CloudFeedType)
		if !ok {
			fmt.Printf("Unexpected type for source ID %d\n", dataSourceType.ID)
			continue
		}

		cloudFeedTypes = append(cloudFeedTypes, cloudFeedType)
	}

	for _, cloudFeed := range cloudFeedTypes {
		cloudFeedAuth, err := s.cloudFeedService.Find(cloudfeed.CloudFeed{AccountID: id, CloudFeedTypeID: cloudFeed.ID})
		if err != nil && !helpers.IsMySQLRecordNotFoundError(err) {
			return cloudFeedAuthStatuses, err
		}

		cloudFeedAuthStatuses = append(cloudFeedAuthStatuses, cloudfeedstatus.MakeCloudFeedStatus(cloudFeed, cloudFeedAuth))
	}

	return cloudFeedAuthStatuses, nil
}
