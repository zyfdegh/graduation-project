package entity

type Used struct {
	Cpus         float32 `json:"cpus"`
	Mems         int64   `json:"mems"`
	Disks        int     `json:"disks"`
	ContainerNum int     `json:"container_num"`
}

type Provider struct {
	Total  Used            `json:"total"`
	Status map[string]Used `json:"status"`
}

type ProviderMetrix struct {
	UserId    string              `json:"user_id"`
	TenantId  string              `json:"tenant_id"`
	Providers map[string]Provider `json:"providers"`
}

type SGMetrix struct {
	UserId    string           `json:"user_id"`
	TenantId  string           `json:"tenant_id"`
	TotalNum  int64            `json:"total_num"`
	StatusNum map[string]int64 `json:"status_num"`
}
