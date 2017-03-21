package controllers

// ResourceDirMap use
type ResourceDirMap struct {
	Local  string
	URI    string
	domain string
}

//FullURL use
func (rd *ResourceDirMap) FullURL() string {
	// return DomainName + rd.URI
	return rd.domain + rd.URI
}

func newResourceDirMap(local, uri, domain string) *ResourceDirMap {
	return &ResourceDirMap{
		Local:  local,
		URI:    uri,
		domain: domain,
	}
}
