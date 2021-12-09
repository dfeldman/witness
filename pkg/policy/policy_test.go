package policy

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/testifysec/witness/pkg/attestation"
	"gitlab.com/testifysec/witness/pkg/attestation/commandrun"
	witcrypt "gitlab.com/testifysec/witness/pkg/crypto"
	"gitlab.com/testifysec/witness/pkg/dsse"
	"gitlab.com/testifysec/witness/pkg/intoto"
)

func createTestKey() (witcrypt.Signer, witcrypt.Verifier, []byte, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, err
	}

	signer := witcrypt.NewRSASigner(privKey, crypto.SHA256)
	verifier := witcrypt.NewRSAVerifier(&privKey.PublicKey, crypto.SHA256)
	keyBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, nil, nil, err
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: keyBytes})
	if err != nil {
		return nil, nil, nil, err
	}

	return signer, verifier, pemBytes, nil
}

func TestVerify(t *testing.T) {
	signer, verifier, pubKeyPem, err := createTestKey()
	require.NoError(t, err)
	keyID, err := verifier.KeyID()
	require.NoError(t, err)
	policy := Policy{
		Expires: time.Now().Add(1 * time.Hour),
		PublicKeys: map[string]PublicKey{
			keyID: {
				KeyID: keyID,
				Key:   pubKeyPem,
			},
		},
		Steps: map[string]Step{
			"step1": {
				Name: "step1",
				Functionaries: []Functionary{
					{
						Type:        "PublicKey",
						PublicKeyID: keyID,
					},
				},
				Attestations: []Attestation{
					{
						Type: commandrun.Type,
					},
				},
			},
		},
	}

	step1Collection := attestation.NewCollection("step1", []attestation.Attestor{commandrun.New()})
	step1CollectionJson, err := json.Marshal(&step1Collection)
	intotoStatement, err := intoto.NewStatement(attestation.CollectionType, step1CollectionJson, map[string]witcrypt.DigestSet{})
	require.NoError(t, err)
	statementJson, err := json.Marshal(&intotoStatement)
	require.NoError(t, err)
	env, err := dsse.Sign(intoto.PayloadType, bytes.NewReader(statementJson), signer)
	require.NoError(t, err)
	encodedEnv, err := json.Marshal(&env)
	require.NoError(t, err)
	envReader := bytes.NewReader(encodedEnv)
	assert.NoError(t, policy.Verify([]io.Reader{envReader}))
}
