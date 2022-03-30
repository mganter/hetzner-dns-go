package hetznerdns

type Zones struct {
	Zones []Zone `json:"zones"`
}

type Zone struct {
	ID              ZoneID   `json:"id"`
	Created         string   `json:"created"`
	Modified        string   `json:"modified"`
	LegacyDNSHost   string   `json:"legacy_dns_host"`
	LegacyNs        []string `json:"legacy_ns"`
	Name            string   `json:"name"`
	Ns              []string `json:"ns"`
	Owner           string   `json:"owner"`
	Paused          bool     `json:"paused"`
	Permission      string   `json:"permission"`
	Project         string   `json:"project"`
	Registrar       string   `json:"registrar"`
	Status          string   `json:"status"`
	TTL             int      `json:"ttl"`
	Verified        string   `json:"verified"`
	RecordsCount    int      `json:"records_count"`
	IsSecondaryDNS  bool     `json:"is_secondary_dns"`
	TxtVerification struct {
		Name  string `json:"name"`
		Token string `json:"token"`
	} `json:"txt_verification"`
}

type Record struct {
	Type     string `json:"type"`
	Id       string `json:"id"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
	ZoneId   string `json:"zone_id"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Ttl      int    `json:"ttl"`
}

type Records []Record
