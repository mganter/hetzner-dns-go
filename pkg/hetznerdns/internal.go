package hetznerdns

type zoneResponse struct {
	Zone Zone `json:"zone"`
}

type zonesResponse struct {
	Meta struct {
		Pagination struct {
			Page         int `json:"page"`
			PerPage      int `json:"per_page"`
			LastPage     int `json:"last_page"`
			TotalEntries int `json:"total_entries"`
		} `json:"pagination"`
	} `json:"meta"`

	Zones []Zone `json:"zones"`
}

type recordsResponse struct {
	Records Records `json:"records"`
}
