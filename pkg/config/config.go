package config

import (
	"fmt"
	"github.com/autom8ter/morpheus/pkg/constants"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/spf13/viper"
	"os"
	"time"
)

type Config struct {
	Server   *Server   `mapstructure:"server"`
	Auth     *Auth     `mapstructure:"auth"`
	Features *Features `mapstructure:"features"`
	Database *Database `mapstructure:"database"`
}

func LoadConfig(configFile string) (*Config, error) {
	homedir, _ := os.UserHomeDir()
	viper.AutomaticEnv()
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("database.storage_path", fmt.Sprintf("%s/.morpheus", homedir))
	viper.SetDefault("features.log_queries", false)
	viper.SetDefault("features.introspection", false)
	viper.SetDefault("features.apollo_tracing", false)
	viper.SetDefault("features.playground", true)
	viper.SetDefault("server.raft_cluster", "")

	viper.SetDefault("auth.signing_secret", "default_secret")
	viper.SetDefault("auth.token_ttl", 24*time.Hour)
	viper.SetDefault("auth.users", []interface{}{
		User{
			Username: constants.ProjectName,
			Password: constants.ProjectName,
			Roles: []Role{
				ADMIN,
			},
		},
	})
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		logger.L.Info("missing config file", map[string]interface{}{
			"expected_path": configFile,
		})
	}
	c := &Config{}
	if err := viper.Unmarshal(c); err != nil {
		return nil, err
	}
	return c, nil
}

type Server struct {
	Port          int    `mapstructure:"port"`
	RaftCluster   string `mapstructure:"raft_cluster"`
	RaftSecret    string `mapstructure:"raft_secret"`
	RaftBroadcast string `mapstructure:"raft_broadcast"`
	TLSKey        string `mapstructure:"tls_key"`
	TLSCert       string `mapstructure:"tls_cert"`
}

type Database struct {
	StoragePath string `mapstructure:"storage_path"`
}

type Features struct {
	GraphqlConsole string `mapstructure:"graphqlConsole"`
	LogQueries     bool   `mapstructure:"log_queries"`
	ApolloTracing  bool   `mapstructure:"apollo_tracing"`
	Introspection  bool   `mapstructure:"introspection"`
	Playground     bool   `mapstructure:"playground"`
}

type Auth struct {
	Disabled      bool          `mapstructure:"disabled"`
	SigningSecret string        `mapstructure:"signing_secret"`
	TokenTTL      time.Duration `mapstructure:"token_ttl"`
	Users         []User        `mapstructure:"users"`
}

type Role string

const (
	READER Role = "reader"
	WRITER Role = "writer"
	ADMIN  Role = "admin"
)

var Roles = []Role{READER, WRITER, ADMIN}

type User struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Roles    []Role `mapstructure:"roles"`
}
