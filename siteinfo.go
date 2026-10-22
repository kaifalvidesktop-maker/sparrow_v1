package main

// ---------------------------------------------------
// SITE INFO POPUP (clicking the shield icon shows this)
// ---------------------------------------------------

type SiteInfo struct {
	Domain              string            `json:"domain"`
	IsSecure            bool              `json:"isSecure"`
	AdsBlocked          bool              `json:"adsBlocked"`
	TrackersBlocked     bool              `json:"trackersBlocked"`
	Permissions         map[string]string `json:"permissions"`
	ThreatsBlockedTotal int               `json:"threatsBlockedTotal"`
}

// BuildSiteInfo assembles the full security/privacy snapshot for a domain
func BuildSiteInfo(rawURL string) SiteInfo {
	domain := extractDomain(rawURL)
	perms := permissionManager.GetFor(domain)

	return SiteInfo{
		Domain:          domain,
		IsSecure:        len(rawURL) >= 8 && rawURL[:8] == "https://",
		AdsBlocked:      shield.AdBlockOn,
		TrackersBlocked: shield.TrackerBlockOn,
		Permissions: map[string]string{
			"camera":        string(perms.Camera),
			"microphone":    string(perms.Microphone),
			"location":      string(perms.Location),
			"notifications": string(perms.Notifications),
		},
		ThreatsBlockedTotal: shield.BlockedCount,
	}
}
