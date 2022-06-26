package main

type (
	Check struct{}
)

func (c *Check) Run(cli *CLI) error {
	return nil
}
