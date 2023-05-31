package twomes

// A CloudFeedAuth stores auth information about CloudFeeds authorized by an account.
type CloudFeedAuth struct {
	AccountID      uint   `json:"account_id"`
	CloudFeedID    uint   `json:"cloud_feed_id"`
	AccessToken    string `json:"-"`
	RefreshToken   string `json:"-"`
	AuthGrantToken string `json:"auth_grant_token"`
}

// Create a new CloudFeedAuth.
func MakeCloudFeedAuth(accountID, cloudFeedID uint, authGrantToken string) CloudFeedAuth {
	return CloudFeedAuth{
		AccountID:      accountID,
		CloudFeedID:    cloudFeedID,
		AuthGrantToken: authGrantToken,
	}
}
