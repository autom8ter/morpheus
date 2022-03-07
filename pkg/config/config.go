package config

import (
	"fmt"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/spf13/viper"
	"os"
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
	viper.SetDefault("server.graphql_port", 8080)
	viper.SetDefault("server.raft_port", 7598)
	viper.SetDefault("database.storage_path", fmt.Sprintf("%s/.morpheus", homedir))
	viper.SetDefault("features.log_queries", false)
	viper.SetDefault("features.introspection", false)
	viper.SetDefault("features.apollo_tracing", false)
	viper.SetDefault("features.playground", true)
	viper.SetDefault("server.raft_cluster", "")
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
	GraphQLPort   int    `mapstructure:"graphql_port"`
	RaftPort      int    `mapstructure:"raft_port"`
	RaftCluster   string `mapstructure:"raft_cluster"`
	RaftSecret    string `mapstructure:"raft_secret"`
	RaftBroadcast string `mapstructure:"raft_broadcast"`
	TLSKey        string `mapstructure:"tls_key"`
	TLSCert       string `mapstructure:"tls_cert"`
}

type Auth struct {
	Users []*User `mapstructure:"users"`
}

type User struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	ReadOnly bool   `mapstructure:"read_only"`
}

type Database struct {
	StoragePath string `mapstructure:"storage_path"`
}

type Features struct {
	LogQueries    bool `mapstructure:"log_queries"`
	ApolloTracing bool `mapstructure:"apollo_tracing"`
	Introspection bool `mapstructure:"introspection"`
	Playground    bool `mapstructure:"playground"`
}
