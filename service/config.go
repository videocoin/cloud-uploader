package service

type Config struct {
	Name    string `envconfig:"-"`
	Version string `envconfig:"-"`

	Addr            string `envconfig:"ADDR" default:"0.0.0.0:8090"`
	StreamsRPCAddr  string `envconfig:"STREAMS_RPC_ADDR" default:"127.0.0.1:5102"`
	SplitterRPCAddr string `envconfig:"SPLITTER_RPC_ADDR" default:"127.0.0.1:5103"`
	RedisURI        string `envconfig:"REDISURI" default:"redis://:@127.0.0.1:6379/1"`
	DownloadDir     string `envconfig:"DOWNLOAD_DIR" default:"/tmp"`
	EnableCORS      bool   `default:"true"`
	GDriveKey       string `envconfig:"GDRIVE_KEY" required:"true"`
	AuthTokenSecret string `envconfig:"AUTH_TOKEN_SECRET" default:"secret"`
}
