package assets

import "net/http"

// Mode is used to specify the asset pipelines run mode i.e. production, development...
type Mode int

// Application run modes
const (
	ProductionMode Mode = iota
	DevelopmentMode
)

const (
	defaultAssetsURL = "/assets/"
	defaultAssetPath = "./assets"
)

var fileServer http.Handler

// Config houses the configuration information for running the asset pipeline
type Config struct {
	RunMode   Mode
	AssetURL  string
	AssetPath string
	ServeMux  *http.ServeMux
}

// Init initializes the asset pipeline using the configurations passed
func Init(config *Config) {

	if config.ServeMux == nil {
		config.ServeMux = http.DefaultServeMux
	}

	if len(config.AssetURL) == 0 {
		config.AssetURL = defaultAssetsURL
	}

	if len(config.AssetPath) == 0 {
		config.AssetPath = defaultAssetPath
	}

	fileServer = http.FileServer(http.Dir(config.AssetPath) + "/..")

	config.ServeMux.Handle(config.AssetURL, http.HandlerFunc(serveAssets))
}

func serveAssets(w http.ResponseWriter, r *http.Request) {
	fileServer.ServeHTTP(w, r)
}
