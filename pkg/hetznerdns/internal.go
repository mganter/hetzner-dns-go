package hetznerdns

type zoneResponse struct {
	Zone Zone `json:"zone"`
}

type zonesResponse struct {
	Meta  Meta   `json:"meta"`
	Zones []Zone `json:"zones"`
}

type recordsResponse struct {
	Records Records `json:"records"`
}
