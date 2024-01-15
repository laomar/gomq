package cluster

type Config struct {
	NodeName         string
	RejoinAfterLeave bool
}

func (c *Config) Parse() {

}

func (c *Config) Validate() error {
	return nil
}

func Default() *Config {
	return &Config{}
}

var Cfg = Default()
