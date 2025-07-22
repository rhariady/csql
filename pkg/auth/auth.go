package auth

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
	vault "github.com/hashicorp/vault/api"
)

type AuthConfig interface{
	GetCredential() string
}

type LocalAuthConfig struct {
	Password string `toml:"password"`
}

func (l LocalAuthConfig) GetCredential() string {
	return l.Password
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

func getPasswordFromVault(address string, mount_path string, secret_path string, secret_key string) (string, error) {
	config := vault.DefaultConfig()
	config.Address = address

	client, err := vault.NewClient(config)
	if err != nil {
		return "", err
	}

	secret, err := client.KVv2(mount_path).Get(context.Background(), secret_path)
	if err != nil {
		return "", err
	}
	if secret == nil {
		return "", fmt.Errorf("no secret found at path: %s", secret_path)
	}

	password, ok := secret.Data[secret_key].(string)
	if !ok {
		return "", fmt.Errorf("key '%s' not found in secret data", secret_key)
	}

	return password, nil
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

