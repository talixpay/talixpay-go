package talixpay

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifyWebhook(t *testing.T) {
	body := []byte(`{"event":"payment.paid"}`)
	secret := "whsec_test"
	sum := sha256.Sum256(append(append([]byte{}, body...), []byte(secret)...))
	hdr := "sha256=" + hex.EncodeToString(sum[:])
	if !VerifyWebhook(body, secret, hdr) {
		t.Fatal("expected valid signature")
	}
	if VerifyWebhook(body, secret, "sha256=deadbeef") {
		t.Fatal("expected invalid signature")
	}
	if VerifyWebhook(body, secret, "md5=abc") {
		t.Fatal("expected bad prefix")
	}
}
