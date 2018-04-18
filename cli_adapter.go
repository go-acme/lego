package main

import (
	"github.com/urfave/cli"
	"github.com/xenolf/lego/cmd"
)

func run(c *cli.Context) error {
	if c.Bool("force-v1") {
		return runV1(c)
	}

	return cmd.Run(c)
}

func revoke(c *cli.Context) error {
	if c.Bool("force-v1") {
		return revokeV1(c)
	}

	return cmd.Revoke(c)
}

func renew(c *cli.Context) error {
	if c.Bool("force-v1") {
		return renewV1(c)
	}

	return cmd.Renew(c)
}
