package vo

type ClusterQueryVO struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Alias     string `json:"alias"`
	Describe  string `json:"describe"`
	Status    string `json:"status"`
	Version   string `json:"version"`
	ApiServer string `json:"apiServer"`
	Endpoint  string `json:"endpoint"`
}

type ClusterVO struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Alias     string `json:"alias"`
	Describe  string `json:"describe"`
	Status    string `json:"status"`
	Version   string `json:"version"`
	Config    string `json:"config"`
	ApiServer string `json:"apiServer"`
	Endpoint  string `json:"endpoint"`
}
