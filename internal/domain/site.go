package domain

type SiteInfo struct {
	Title                   string
	URL                     string
	IsRegistrationAvailable bool
}

type ReadSiteInfoReq struct {
	ID                      int
	Title                   *string
	URL                     string
	IsRegistrationAvailable *bool
}
