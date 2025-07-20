package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Instances map[string]InstanceConfig
}

type SourceType string

const (
	Manual SourceType = "Manual"
	GCP    SourceType = "GCP"
)

type InstanceConfig struct {
	Name      string `toml:"name"`
	ProjectID string `toml:"project_id"`
	Host      string `toml:"host"`
	Source    SourceType `toml:"source"`
	Users     map[string]UserConfig
}

type UserConfig struct {
	Username string `toml:"username"`
	// Auth AuthConfig `toml:"auth"`
	DefaultAuth string `toml:"default_auth"`
	Auth map[string]interface{} `toml:"Auth"`
}

type LocalAuthConfig struct {
	Password string `toml:"password"`
}

func (l LocalAuthConfig) GetCredential() string {
	return l.Password
}

func NewAuthConfig(authType string, authMap map[string]interface{}) (AuthConfig, error) {
	switch authType {
	case "vault":
		var buf bytes.Buffer
		if err := toml.NewEncoder(&buf).Encode(authMap["vault"]); err != nil {
			return nil, err
		}
		authConfigData := buf.String()
		var vaultAuthConfig VaultAuthConfig
		if _, err := toml.Decode(authConfigData, &vaultAuthConfig); err != nil {
			return nil, err
		}
		return vaultAuthConfig, nil
	case "local":
		var buf bytes.Buffer
		if err := toml.NewEncoder(&buf).Encode(authMap["local"]); err != nil {
			return nil, err
		}
		authConfigData := buf.String()
		var localAuthConfig LocalAuthConfig
		if _, err := toml.Decode(authConfigData, &localAuthConfig); err != nil {
			return nil, err
		}
		return localAuthConfig, nil
  default:
		return nil, errors.New("AuthTypeNotFound")
	}
}

type AuthConfig interface{
	GetCredential() string
}

type VaultAuthConfig struct {
	Address string `toml:"address"`
	MountPath string `toml:"mount_path"`
	SecretPath string `toml:"secret_path"`
	SecretKey string `toml:"secret_key"`
}

func (v VaultAuthConfig) GetCredential() string {
	vault_address := v.Address
	vault_mount_path := v.MountPath
	vault_secret_path := v.SecretPath
	vault_secret_key := v.SecretKey

	password, err := getPasswordFromVault(vault_address, vault_mount_path, vault_secret_path, vault_secret_key)
	if err != nil {
		panic(err)
	}

	return password
}

func (c *Config) AddInstance(key string, instanceConfig InstanceConfig) {
	if c.Instances == nil {
		c.Instances = make(map[string]InstanceConfig)
	}
	c.Instances[key] = instanceConfig
}

func GetConfigFile() (*string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(configDir, ".csql")

	return &configFile, nil
}

func CheckConfigFile() (error) {
	configFile, err := GetConfigFile()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(*configFile, os.O_RDWR|os.O_CREATE, 0644)
	defer file.Close()

	return nil
}

func GetConfig() (*Config, error) {

	configFile, err := GetConfigFile()
	if err != nil {
		return nil, err
	}

	var config Config
	if _, err := toml.DecodeFile(*configFile, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (newConfig *Config)  WriteConfig() (error) {
	var buffer bytes.Buffer
	if err := toml.NewEncoder(&buffer).Encode(newConfig); err != nil {
		return err
	}

	configFile, err := GetConfigFile()

	if err != nil {
		return nil
	}

	if err := os.WriteFile(*configFile, buffer.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}
