package root

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/registration"
	"github.com/go-acme/lego/v5/registration/zerossl"
)

func registerAccount(ctx context.Context, client *lego.Client, accountConfig *configuration.Account) (*acme.ExtendedAccount, error) {
	accepted := handleTOS(client, accountConfig)
	if !accepted {
		return nil, errors.New("you did not accept the TOS: unable to proceed")
	}

	if accountConfig.ExternalAccountBinding != nil {
		return client.Registration.RegisterWithExternalAccountBinding(ctx, registration.RegisterEABOptions{
			TermsOfServiceAgreed: true,
			Kid:                  accountConfig.ExternalAccountBinding.KID,
			HmacEncoded:          accountConfig.ExternalAccountBinding.HmacKey,
		})
	} else if zerossl.IsZeroSSL(accountConfig.Server) {
		return registration.RegisterWithZeroSSL(ctx, client.Registration, accountConfig.Email)
	}

	return client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
}

func handleTOS(client *lego.Client, accountConfig *configuration.Account) bool {
	// metadata items are optional, and termsOfService too.
	urlTOS := client.GetServerMetadata().TermsOfService
	if urlTOS == "" {
		return true
	}

	// Check for a global acceptance override
	if accountConfig.AcceptsTermsOfService {
		return true
	}

	reader := bufio.NewReader(os.Stdin)

	log.Warn("Please review the TOS", slog.String("url", urlTOS))

	for {
		fmt.Println("Do you accept the TOS? Y/n")

		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Could not read from the console", log.ErrorAttr(err))
		}

		text = strings.Trim(text, "\r\n")
		switch text {
		case "", "y", "Y":
			return true
		case "n", "N":
			return false
		default:
			fmt.Println("Your input was invalid. Please answer with one of Y/y, n/N or by pressing enter.")
		}
	}
}
