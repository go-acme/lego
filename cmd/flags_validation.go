package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func runFlagsValidation(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	// we require either domains or csr, but not both
	hasDomains := len(cmd.StringSlice(flgDomains)) > 0

	hasCsr := cmd.String(flgCSR) != ""
	if hasDomains && hasCsr {
		return ctx, fmt.Errorf("please specify either '--%s'/'-%s' or '--%s', but not both",
			flgDomains, flgAliasDomains, flgCSR)
	}

	if !hasDomains && !hasCsr {
		return ctx, fmt.Errorf("please specify '--%s'/'-%s' (or '--%s' if you already have a CSR)",
			flgDomains, flgAliasDomains, flgCSR)
	}

	err := validateChallengeRequirements(cmd)
	if err != nil {
		return ctx, err
	}

	return ctx, validateNetworkStack(cmd)
}

func renewFlagsValidation(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	hasDomains := len(cmd.StringSlice(flgDomains)) > 0
	hasCsr := cmd.String(flgCSR) != ""
	hasCertID := cmd.String(flgCertName) != ""

	if hasDomains && hasCsr {
		return ctx, fmt.Errorf("please specify either '--%s'/'-%s' or '--%s', but not both",
			flgDomains, flgAliasDomains, flgCSR)
	}

	if !hasCertID && !hasDomains && !hasCsr {
		return ctx, fmt.Errorf("please specify '--%s' or '--%s'/'-%s' (or '--%s' if you already have a CSR)",
			flgCertName, flgDomains, flgAliasDomains, flgCSR)
	}

	if cmd.Bool(flgForceCertDomains) && hasCsr {
		return ctx, fmt.Errorf("'--%s' only works with '--%s'/'-%s', '--%s' doesn't support this option",
			flgForceCertDomains, flgDomains, flgAliasDomains, flgCSR)
	}

	err := validateChallengeRequirements(cmd)
	if err != nil {
		return ctx, err
	}

	return ctx, validateNetworkStack(cmd)
}

func validateNetworkStack(cmd *cli.Command) error {
	if cmd.Bool(flgIPv4Only) && cmd.Bool(flgIPv6Only) {
		return fmt.Errorf("cannot specify both '--%s' and '--%s'", flgIPv4Only, flgIPv6Only)
	}

	return nil
}

func validateChallengeRequirements(cmd *cli.Command) error {
	if !cmd.Bool(flgHTTP) && !cmd.Bool(flgTLS) && !cmd.IsSet(flgDNS) && !cmd.Bool(flgDNSPersist) {
		return fmt.Errorf("no challenge selected: you must specify at least one challenge: '--%s', '--%s', '--%s', '--%s'",
			flgHTTP, flgTLS, flgDNS, flgDNSPersist)
	}

	if isSetBool(cmd, flgDNS) {
		err := validatePropagationExclusiveOptions(cmd, flgDNSPropagationWait, flgDNSPropagationDisableANS, flgDNSPropagationDisableRNS)
		if err != nil {
			return err
		}
	}

	if isSetBool(cmd, flgDNSPersist) {
		err := validatePropagationExclusiveOptions(cmd, flgDNSPersistPropagationWait, flgDNSPersistPropagationDisableANS, flgDNSPersistIssuerDomainName)
		if err != nil {
			return err
		}
	}

	return nil
}

func validatePropagationExclusiveOptions(cmd *cli.Command, flgWait, flgANS, flgDNS string) error {
	if !cmd.IsSet(flgWait) {
		return nil
	}

	if isSetBool(cmd, flgANS) {
		return fmt.Errorf("'--%s' and '--%s' are mutually exclusive", flgWait, flgANS)
	}

	if isSetBool(cmd, flgDNS) {
		return fmt.Errorf("'--%s' and '--%s' are mutually exclusive", flgWait, flgDNS)
	}

	return nil
}

func isSetBool(cmd *cli.Command, name string) bool {
	return cmd.IsSet(name) && cmd.Bool(name)
}
