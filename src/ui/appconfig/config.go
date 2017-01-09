package appconfig

import (
	"os"
	"strings"

	"github.com/astaxie/beego"
)

func init() {
	configPath := os.Getenv("CONFIG_PATH")
	if len(configPath) != 0 {
		beego.LoadAppConfig("ini", configPath)
	}
}

// MongoHosts ...
func MongoHosts() string {
	return beego.AppConfig.String("MONGO_HOSTS")
}

// MongoPort ...
func MongoPort() string {
	return beego.AppConfig.String("MONGO_PORT")
}

// MongoRSName ...
func MongoRSName() string {
	return beego.AppConfig.String("MONGO_RSNAME")
}

// MongoUsername ...
func MongoUsername() string {
	return beego.AppConfig.String("MONGO_USERNAME")
}

// MongoPassword ...
func MongoPassword() string {
	return beego.AppConfig.String("MONGO_PASSWORD")
}

// VerifyRemoteCert returns bool value.
func VerifyRemoteCert() bool {
	val, err := beego.AppConfig.Bool("VERIFY_REMOTE_CERT")
	if err != nil {
		panic(err)
	}
	return val
}

// ExtEndpoint ...
func ExtEndpoint() string {
	return beego.AppConfig.String("EXT_ENDPOINT")
}

// TokenEndpoint returns the endpoint string of token service, which can be accessed by internal service of Harbor.
func TokenEndpoint() string {
	return beego.AppConfig.String("TOKEN_ENDPOINT")
}

// LogLevel returns the log level in string format.
func LogLevel() string {
	return beego.AppConfig.String("LOG_LEVEL")
}

// AuthMode ...
func AuthMode() string {
	return beego.AppConfig.String("AUTH_MODE")
}

// AppOnly ...
func AppOnly() bool {
	val, err := beego.AppConfig.Bool("APP_ONLY")
	if err != nil {
		panic(err)
	}
	return val
}

// AppCheck ...
func AppCheck() bool {
	val, err := beego.AppConfig.Bool("APP_CHECK")
	if err != nil {
		panic(err)
	}
	return val
}

// TokenExpiration returns the token expiration time (in minute)
func TokenExpiration() int {
	val, err := beego.AppConfig.Int("TOKEN_EXPIRATION")
	if err != nil {
		panic(err)
	}
	return val
}

// ExtRegistryURL returns the registry URL to exposed to external client
func ExtRegistryURL() string {
	return beego.AppConfig.String("EXT_REG_URL")
}

// UISecret returns the value of UI secret cookie, used for communication between UI and JobService
func UISecret() string {
	return beego.AppConfig.String("UI_SECRET")
}

// SecretKey returns the secret key to encrypt the password of target
func SecretKey() string {
	return beego.AppConfig.String("SECRET_KEY")
}

// SelfRegistration returns the enablement of self registration
func SelfRegistration() bool {
	val, err := beego.AppConfig.Bool("SELF_REGISTRATION")
	if err != nil {
		panic(err)
	}
	return val
}

// InternalRegistryURL returns registry URL for internal communication between Harbor containers
func InternalRegistryURL() string {
	registryURL := beego.AppConfig.String("REGISTRY_URL")
	registryURL = strings.TrimRight(registryURL, "/")
	return registryURL
}

// InternalJobServiceURL returns jobservice URL for internal communication between Harbor containers
func InternalJobServiceURL() string {
	jobserviceURL := beego.AppConfig.String("JOB_SERVICE_URL")
	jobserviceURL = strings.TrimRight(jobserviceURL, "/")
	return jobserviceURL
}

// InitialAdminPassword returns the initial password for administrator
func InitialAdminPassword() string {
	return beego.AppConfig.String("HARBOR_ADMIN_PASSWORD")
}

// OnlyAdminCreateProject returns the flag to restrict that only sys admin can create project
func OnlyAdminCreateProject() bool {
	return strings.ToLower(beego.AppConfig.String("PROJECT_CREATION_RESTRICTION")) == "adminonly"
}
