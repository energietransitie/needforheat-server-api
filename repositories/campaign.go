package repositories

import (
	"time"

	"github.com/energietransitie/twomes-backoffice-api/twomes"
	"gorm.io/gorm"
)

type CampaignRepository struct {
	db *gorm.DB
}

func NewCampaignRepository(db *gorm.DB) *CampaignRepository {
	return &CampaignRepository{
		db: db,
	}
}

// Database representation of [twomes.Campaign].
type CampaignModel struct {
	gorm.Model
	Name       string `gorm:"unique;not null"`
	AppModelID uint   `gorm:"column:app_id"`
	App        AppModel
	InfoURL    string           `gorm:"unique;not null"`
	CloudFeeds []CloudFeedModel `gorm:"many2many:campaign_cloud_feed"`
	StartTime  *time.Time
	EndTime    *time.Time
}

// Set the name of the table in the database.
func (CampaignModel) TableName() string {
	return "campaign"
}

// Create a new CampaignModel from a [twomes.campaign].
func MakeCampaignModel(campaign twomes.Campaign) CampaignModel {
	var cloudFeedModels []CloudFeedModel

	for _, cloudFeed := range campaign.CloudFeeds {
		cloudFeedModels = append(cloudFeedModels, MakeCloudFeedModel(cloudFeed))
	}

	return CampaignModel{
		Model: gorm.Model{
			ID: campaign.ID,
		},
		Name:       campaign.Name,
		AppModelID: campaign.App.ID,
		App:        MakeAppModel(campaign.App),
		InfoURL:    campaign.InfoURL,
		CloudFeeds: cloudFeedModels,
		StartTime:  campaign.StartTime,
		EndTime:    campaign.EndTime,
	}
}

// Create a [twomes.Campaign] from an CampaignModel.
func (m *CampaignModel) fromModel() twomes.Campaign {
	var cloudFeeds []twomes.CloudFeed

	for _, cloudFeedModel := range m.CloudFeeds {
		cloudFeeds = append(cloudFeeds, cloudFeedModel.fromModel())
	}

	return twomes.Campaign{
		ID:         m.ID,
		Name:       m.Name,
		App:        m.App.fromModel(),
		InfoURL:    m.InfoURL,
		CloudFeeds: cloudFeeds,
		StartTime:  m.StartTime,
		EndTime:    m.EndTime,
	}
}

func (r *CampaignRepository) Find(campaign twomes.Campaign) (twomes.Campaign, error) {
	campaignModel := MakeCampaignModel(campaign)
	err := r.db.Preload("App").Where(&campaignModel).First(&campaignModel).Error
	return campaignModel.fromModel(), err
}

func (r *CampaignRepository) GetAll() ([]twomes.Campaign, error) {
	var campaigns []twomes.Campaign

	var campaignModels []CampaignModel
	err := r.db.Preload("App").Find(&campaignModels).Error
	if err != nil {
		return nil, err
	}

	for _, campaignModel := range campaignModels {
		campaigns = append(campaigns, campaignModel.fromModel())
	}

	return campaigns, nil
}

func (r *CampaignRepository) Create(campaign twomes.Campaign) (twomes.Campaign, error) {
	campaignModel := MakeCampaignModel(campaign)
	err := r.db.Create(&campaignModel).Error
	return campaignModel.fromModel(), err
}

func (r *CampaignRepository) Delete(campaign twomes.Campaign) error {
	CampaignModel := MakeCampaignModel(campaign)
	return r.db.Delete(&CampaignModel).Error
}
