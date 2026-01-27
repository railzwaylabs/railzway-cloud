package deployment

import (
	"fmt"
	"strings"

	"github.com/railzwaylabs/railzway-cloud/internal/config"
	"github.com/railzwaylabs/railzway-cloud/internal/cryptoutils"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/instance"
)

func resolvePaymentProviderSecret(cfg *config.Config, inst *instance.Instance) (string, error) {
	key := strings.TrimSpace(cfg.InstanceSecretEncryptionKey)
	if key == "" {
		return "", fmt.Errorf("INSTANCE_SECRET_ENCRYPTION_KEY is required")
	}

	if strings.TrimSpace(inst.PaymentProviderConfigSecretEncrypted) == "" {
		secret, err := generatePaymentProviderSecret()
		if err != nil {
			return "", fmt.Errorf("generate payment provider secret: %w", err)
		}
		encrypted, err := cryptoutils.Encrypt(secret, key)
		if err != nil {
			return "", fmt.Errorf("encrypt payment provider secret: %w", err)
		}
		inst.PaymentProviderConfigSecretEncrypted = encrypted
		return secret, nil
	}

	secret, err := cryptoutils.Decrypt(inst.PaymentProviderConfigSecretEncrypted, key)
	if err != nil {
		return "", fmt.Errorf("decrypt payment provider secret: %w", err)
	}
	return secret, nil
}

func generatePaymentProviderSecret() (string, error) {
	return generatePassword()
}
