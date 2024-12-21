package auth

import (
	"github.com/pascaldekloe/jwt"
	"time"
)

// CreateJWT generates a JSON Web Token (JWT) for a patron using the provided patronID and jwtSecret.
func CreateJWT(patronID, jwtSecret, issuer, audience string) ([]byte, error) {
	var claims jwt.Claims

	claims.Subject = patronID
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	claims.Issuer = issuer
	claims.Audiences = []string{audience}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(jwtSecret))
	if err != nil {
		return jwtBytes, err
	}

	return jwtBytes, nil
}
