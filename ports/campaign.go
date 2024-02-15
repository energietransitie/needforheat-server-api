package ports

import (
	"time"

	"github.com/energietransitie/twomes-backoffice-api/twomes/app"
	"github.com/energietransitie/twomes-backoffice-api/twomes/campaign"
	"github.com/energietransitie/twomes-backoffice-api/twomes/cloudfeed"
)

// CampaignService exposes all operations that can be performed on a [campaign.Campaign].
type CampaignService interface {
	Create(name string, app app.App, infoURL string, cloudFeeds []cloudfeed.CloudFeed, startTime, endTime *time.Time) (campaign.Campaign, error)
	Find(campaign campaign.Campaign) (campaign.Campaign, error)
	GetByID(id uint) (campaign.Campaign, error)
}
