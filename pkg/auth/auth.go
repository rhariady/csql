package auth

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
	vault "github.com/hashicorp/vault/api"
	"github.com/rivo/tview"
)

type AuthType = string

const (
	Local AuthType = "Local"
	Vault  AuthType = "Vault"
)

var AuthList = map[AuthType]IAuth{
	Local: &LocalAuth{},
	Vault: &VaultAuth{},
}

type IAuth interface{
	GetCredential() (string, error)
	GetFormInput(form *tview.Form)
	ParseFormInput(form *tview.Form) map[string]interface{}
}

func GetAuth(authType string, authParams map[string]interface{}) (IAuth, error) {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(authParams); err != nil {
		return nil, err
	}
	authConfigData := buf.String()

	switch authType {
	case Vault:
		var vaultAuth VaultAuth
		if _, err := toml.Decode(authConfigData, &vaultAuth); err != nil {
			return nil, err
		}
		return vaultAuth, nil
	case Local:
		var localAuth LocalAuth
		if _, err := toml.Decode(authConfigData, &localAuth); err != nil {
			return nil, err
		}
		return localAuth, nil
  default:
		return nil, errors.New("AuthTypeNotFound")
	}
}

type LocalAuth struct {
	Password string `toml:"password"`
}

func (l LocalAuth) GetCredential() (string, error) {
	return l.Password, nil
}

func (l LocalAuth) GetFormInput(form *tview.Form) {
	form.AddInputField("Password", "", 0, nil, nil)
}

func (l LocalAuth) ParseFormInput(form *tview.Form) map[string]interface{} {
	password := form.GetFormItemByLabel("Password").(*tview.InputField).GetText()
	return map[string]interface{}{
		"password": password,
	}
}

type VaultAuth struct {
	Address string `toml:"address"`
	MountPath string `toml:"mount_path"`
	SecretPath string `toml:"secret_path"`
	SecretKey string `toml:"secret_key"`
}

func (v VaultAuth) GetCredential() (string, error) {
	vault_address := v.Address
	vault_mount_path := v.MountPath
	vault_secret_path := v.SecretPath
	vault_secret_key := v.SecretKey

	password, err := getPasswordFromVault(vault_address, vault_mount_path, vault_secret_path, vault_secret_key)
	if err != nil {
		return "", err
	}

	return password, nil
}

func (l VaultAuth) GetFormInput(form *tview.Form) {
	form.
		AddInputField("Vault Address", "", 0, nil, nil).
		AddInputField("Vault Mount Path", "", 0, nil, nil).
		AddInputField("Vault Secret Path", "", 0, nil, nil).
		AddInputField("Vault Secret Key", "", 0, nil, nil)
}

func (l VaultAuth) ParseFormInput(form *tview.Form) map[string]interface{} {
	vaultAddress := form.GetFormItemByLabel("Vault Address").(*tview.InputField).GetText()
	vaultMountPath := form.GetFormItemByLabel("Vault Mount Path").(*tview.InputField).GetText()
	vaultSecretPath := form.GetFormItemByLabel("Vault Secret Path").(*tview.InputField).GetText()
	vaultSecretKey := form.GetFormItemByLabel("Vault Secret Key").(*tview.InputField).GetText()

	return map[string]interface{}{
		"address":     vaultAddress,
		"mount_path":  vaultMountPath,
		"secret_path": vaultSecretPath,
		"secret_key":  vaultSecretKey,
	}
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

