package flags

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func RunFlagsValidation(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	hasDomains := len(cmd.StringSlice(FlgDomains)) > 0
	hasCsr := cmd.String(FlgCSR) != ""
	hasCertID := cmd.String(FlgCertName) != ""

	if hasDomains && hasCsr {
		return ctx, fmt.Errorf("please specify either '--%s'/'-%s' or '--%s', but not both",
			FlgDomains, flgAliasDomains, FlgCSR)
	}

	if !hasCertID && !hasDomains && !hasCsr {
		return ctx, fmt.Errorf("please specify '--%s' and/or '--%s'/'-%s' (or '--%s' if you already have a CSR)",
			FlgCertName, FlgDomains, flgAliasDomains, FlgCSR)
	}

	if cmd.Bool(FlgForceCertDomains) && hasCsr {
		return ctx, fmt.Errorf("'--%s' only works with '--%s'/'-%s', '--%s' doesn't support this option",
			FlgForceCertDomains, FlgDomains, flgAliasDomains, FlgCSR)
	}

	err := validateChallengeRequirements(cmd)
	if err != nil {
		return ctx, err
	}

	return ctx, validateNetworkStack(cmd)
}

func validateNetworkStack(cmd *cli.Command) error {
	if cmd.Bool(FlgIPv4Only) && cmd.Bool(FlgIPv6Only) {
		return fmt.Errorf("cannot specify both '--%s' and '--%s'", FlgIPv4Only, FlgIPv6Only)
	}

	return nil
}

func validateChallengeRequirements(cmd *cli.Command) error {
	if !cmd.Bool(FlgHTTP) && !cmd.Bool(FlgTLS) && !cmd.IsSet(FlgDNS) && !cmd.Bool(FlgDNSPersist) {
		return fmt.Errorf("no challenge selected: you must specify at least one challenge: '--%s', '--%s', '--%s', '--%s'",
			FlgHTTP, FlgTLS, FlgDNS, FlgDNSPersist)
	}

	if isSetBool(cmd, FlgDNS) {
		err := validatePropagationExclusiveOptions(cmd, FlgDNSPropagationWait, FlgDNSPropagationDisableANS, FlgDNSPropagationDisableRNS)
		if err != nil {
			return err
		}
	}

	if isSetBool(cmd, FlgDNSPersist) {
		err := validatePropagationExclusiveOptions(cmd, FlgDNSPersistPropagationWait, FlgDNSPersistPropagationDisableANS, FlgDNSPersistIssuerDomainName)
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
