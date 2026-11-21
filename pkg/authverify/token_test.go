package authverify

import (
	"testing"
	"time"

	"gitee.com/matkurban/openimsdk-tools/errs"
	"github.com/golang-jwt/jwt/v5"
)

func TestCreateAndParseToken(t *testing.T) {
	const secret = "test-secret-key"

	tok, err := CreateToken("user123", 5, 7, secret)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}

	claims, err := GetClaimFromToken(tok, secret)
	if err != nil {
		t.Fatalf("GetClaimFromToken failed: %v", err)
	}
	if claims.UserID != "user123" || claims.PlatformID != 5 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestGetClaimFromTokenErrors(t *testing.T) {
	const secret = "test-secret-key"

	if _, err := GetClaimFromToken("not-a-jwt", secret); !errs.ErrTokenMalformed.Is(err) {
		t.Fatalf("expected ErrTokenMalformed, got: %v", err)
	}

	expiredClaims := Claims{
		UserID:     "user123",
		PlatformID: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	expiredTok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign expired token failed: %v", err)
	}
	if _, err := GetClaimFromToken(expiredTok, secret); !errs.ErrTokenExpired.Is(err) {
		t.Fatalf("expected ErrTokenExpired, got: %v", err)
	}

	// Reject HS512 even though it is an HMAC algorithm, since only HS256 is accepted.
	hs512Tok, err := jwt.NewWithClaims(jwt.SigningMethodHS512, BuildClaims("user123", 1, 7)).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign HS512 token failed: %v", err)
	}
	if _, err := GetClaimFromToken(hs512Tok, secret); err == nil {
		t.Fatalf("expected HS512 token to be rejected")
	}
}
