package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Instances map[string]InstanceConfig
}

type InstanceConfig struct {
	Name      string `toml:"name"`
	Host      string `toml:"host"`
	Port      int `toml:"port"`
	Type      string `toml:"type"`
	Users     []UserConfig
	Params    map[string]interface{} `toml:"params"`
}

type UserConfig struct {
	Username string `toml:"username"`
	// Auth AuthConfig `toml:"auth"`
	AuthType string `toml:"auth_type"`
	AuthParams map[string]interface{} `toml:"params"`
}

func (c *Config) AddInstance(key string, instanceConfig InstanceConfig) {
	if c.Instances == nil {
		c.Instances = make(map[string]InstanceConfig)
	}
	c.Instances[key] = instanceConfig
}

func (c *Config) GetInstance(instanceName string) *InstanceConfig {
	instance := c.Instances[instanceName]

	return &instance
}

func (c *Config) AddInstanceUser(instanceName string, userConfig UserConfig) {
	instance := c.Instances[instanceName]
	instance.Users = append(instance.Users, userConfig)
	c.AddInstance(instanceName, instance)
}

func (c *InstanceConfig) GetUserConfig(userName string) (*UserConfig, error) {
	for _, user := range c.Users {
		if user.Username == userName {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("User Not Found")
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
