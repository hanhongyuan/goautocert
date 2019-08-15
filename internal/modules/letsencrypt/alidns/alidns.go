package alidns

import (
	"fmt"
	a "github.com/go-acme/lego/v3/providers/dns/alidns"
	"github.com/ouqiang/gocron/internal/models"
)

func GetAliyunNewDefaultConfig(c models.AccessKey) (config *a.Config, err error) {

	config = a.NewDefaultConfig()
	config.APIKey = c.AccessKeyId
	config.SecretKey = c.AccessKeySecret
	// config.RegionID = d[1]
	if len(config.APIKey) == 0 || len(config.SecretKey) == 0 {
		return nil, fmt.Errorf("APIKey,SecretKey无效！")
	}
	return
}

// NewDNSProvider returns a DNSProvider instance configured for Alibaba Cloud DNS.
// Credentials must be passed in the environment variables: ALICLOUD_ACCESS_KEY and ALICLOUD_SECRET_KEY.
func NewDNSProvider(c models.AccessKey) (*a.DNSProvider, error) {
	if config, err := GetAliyunNewDefaultConfig(c); err != nil {
		return nil, err
	} else {
		return a.NewDNSProviderConfig(config)
	}
}
