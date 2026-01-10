package proton

import (
	"context"

	"github.com/ProtonMail/go-proton-api"
)

func CreateClient(username string, password string) (*proton.Client, error) {
	// Create a new manager.
	m := proton.New()

	// All API operations must be run within a context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Clients are created via username/password if auth information isn't already known.
	c, _, err := m.NewClientWithLogin(ctx, "...user...", []byte("...pass..."))
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// // If 2FA is necessary, an additional request is required.
	// if auth.TwoFA.Enabled&proton.HasTOTP != 0 {
	// 	if err := c.Auth2FA(ctx, proton.Auth2FAReq{TwoFactorCode: "...TOTP..."}); err != nil {
	// 		panic(err)
	// 	}
	// }

	return c, nil
}
