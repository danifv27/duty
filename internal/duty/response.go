package duty

// Response represents an http response of a status code and a given payload
type Response struct {
	Code    int    `yaml:"code"`
	Verb    string `yaml:"verb"`
	Payload string `yaml:"payload"`
	ID      string `yaml:"id"`
	Latency string `yaml:"latency"`
}
