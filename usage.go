package env

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// the width of the left column.
const colWidth = 25

func printTwoCols(w io.Writer, left, help, defaultVal string) {
	lhs := "  " + left
	fmt.Fprint(w, lhs)

	if help != "" {
		if len(lhs)+2 < colWidth {
			fmt.Fprint(w, strings.Repeat(" ", colWidth-len(lhs)))
		} else {
			fmt.Fprint(w, "\n"+strings.Repeat(" ", colWidth))
		}

		fmt.Fprint(w, help)
	}

	bracketsContent := []string{}

	if defaultVal != "" {
		bracketsContent = append(bracketsContent,
			fmt.Sprintf("default: %s", defaultVal),
		)
	}

	if len(bracketsContent) > 0 {
		fmt.Fprintf(w, " [%s]", strings.Join(bracketsContent, ", "))
	}

	fmt.Fprint(w, "\n")
}

// Help writes the usage string followed by the full help string for each option.
func (p *Parser) Help() string {
	var res bytes.Buffer

	p.writeHelp(&res, p.specs)

	return res.String()
}

// writeHelp writes the usage string for the given subcommand.
func (p *Parser) writeHelp(w io.Writer, specs []*spec) {
	options := make([]*spec, 0, len(specs))
	options = append(options, specs...)

	if p.description != "" {
		fmt.Fprintln(w, p.description)
	}

	// write the list of options
	if len(options) > 0 {
		fmt.Fprint(w, "Environments:\n")

		for _, spec := range options {
			p.printOption(w, spec)
		}
	}
}

func (p *Parser) printOption(w io.Writer, spec *spec) {
	left := synopsis(spec, spec.name)
	printTwoCols(w, left, spec.help, spec.defaultVal)
}

func synopsis(spec *spec, form string) string {
	return form
}
