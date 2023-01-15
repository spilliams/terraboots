package scopedata

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/sirupsen/logrus"
	"github.com/spilliams/terraboots/internal/hclhelp"
)

// Generator objects can work with scope data, taking input from the user and
// saving the data to files.
type Generator interface {
	Create(io.Reader, io.Writer) ([]byte, error)
}

// generator stores an ordered list of scope types, a filename to store data to,
// and composes a Logger for debugging
type generator struct {
	scopeTypes []string
	*logrus.Entry
}

// NewGenerator builds a new Generator with the given scope types, destination
// filename, and logger.
func NewGenerator(scopeTypes []string, logger *logrus.Logger) Generator {
	return &generator{
		scopeTypes: scopeTypes,
		Entry:      logger.WithField("prefix", "scopedata"),
	}
}

// Create prompts the user for input using `input` and `output`, and returns
// bytes representing an hcl file
func (g *generator) Create(input io.Reader, output io.Writer) ([]byte, error) {
	rootScopes, err := g.promptForScopeValues(input, output)
	if err != nil {
		return nil, err
	}
	if len(rootScopes) == 0 {
		g.Warn("No scopes were generated, exiting.")
		return nil, nil
	}

	hclfile := g.generateScopeDataFile(rootScopes)
	return hclfile.Bytes(), nil
}

// promptForScopeValues uses the receiver's scopeTypes to ask the user for all
// the different values of the scopes.
// Returns a list of the top-level scope values (the scope values for the first
// scope type)
func (g *generator) promptForScopeValues(input io.Reader, output io.Writer) ([]*NestedScope, error) {
	// TODO: adopt survey instead of doing it myself
	fmt.Fprintln(output, "Scope types in this projct, in order, are:")
	fmt.Fprintln(output, strings.Join(g.scopeTypes, ", "))
	fmt.Fprintln(output, "")

	fmt.Fprintln(output, "Answers must be space-separated, and may consist of the characters")
	// TODO: use this charset to validate the input...
	answerCharacterSet := "0-9a-zA-Z-_"
	fmt.Fprintln(output, answerCharacterSet)
	fmt.Fprintln(output, "")

	fmt.Fprintln(output, "Leave any answer blank to mark the current scope as complete with no children")
	fmt.Fprintln(output, "")

	fmt.Fprintln(output, "Press Ctrl+C at any time to cancel.")
	fmt.Fprintln(output, "")

	scanner := bufio.NewScanner(input)

	// First one's free
	fmt.Fprintf(output, "What are the allowable values for `%s`?\n", g.scopeTypes[0])
	scanner.Scan()
	err := scanner.Err()
	if err != nil {
		return nil, err
	}
	if len(scanner.Text()) == 0 {
		g.Debugf("user entered blank line, exiting")
		return nil, nil
	}
	// TODO: validate input against list of blocklisted words, and the above
	// charset, and each other (no dupes)...
	firstValues := strings.Split(scanner.Text(), " ")
	g.Debugf("read new scope values %v", firstValues)

	roots := make([]*NestedScope, len(firstValues))
	prompts := make([]*NestedScope, len(firstValues))
	for i, el := range firstValues {
		value := &NestedScope{
			Name:           el,
			Type:           g.scopeTypes[0],
			scopeTypeIndex: 0,
			Children:       make([]*NestedScope, 0),
		}
		value.Address = fmt.Sprintf("%s.%s", value.Type, value.Name)
		roots[i] = value
		prompts[i] = value
	}

	for len(prompts) > 0 {
		prompt := prompts[0]
		prompts = prompts[1:]

		if prompt.scopeTypeIndex+1 == len(g.scopeTypes) {
			// this scope value cannot have children
			continue
		}

		fmt.Fprintf(output, "Within %s, what are the allowable values for `%s`?\n", prompt.Address, g.scopeTypes[prompt.scopeTypeIndex+1])

		scanner.Scan()
		err := scanner.Err()
		if err != nil {
			return nil, err
		}
		if len(scanner.Text()) == 0 {
			// user entered none
			g.Debugf("user entered no values for prompt, closing this scope")
			continue
		}
		// TODO: validate input against list of blocklisted words, and the above
		// charset, and each other (no dupes)...

		values := strings.Split(scanner.Text(), " ")
		g.Debugf("read new scope values %v", values)
		for _, el := range values {
			value := &NestedScope{
				Name:           el,
				Type:           g.scopeTypes[prompt.scopeTypeIndex+1],
				scopeTypeIndex: prompt.scopeTypeIndex + 1,
				Children:       make([]*NestedScope, 0),
			}
			value.Address = strings.Join([]string{prompt.Address, string(value.Type), value.Name}, ".")
			prompt.Children = append(prompt.Children, value)
			prompts = append(prompts, value)
		}
	}

	g.Debugf("%+v", roots)

	return roots, nil
}

// generateScopeDataFile reads the given scopes and produces an `hclwrite.File`
// object that is ready to be written to disk.
func (g *generator) generateScopeDataFile(rootScopes []*NestedScope) *hclwrite.File {
	// TODO: now that Scopes are gohcl structs, can we write the file more simply?
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	rootBody.AppendUnstructuredTokens(hclhelp.CommentTokens("This file was generated by terraboots"))
	rootBody.AppendNewline()

	for _, root := range rootScopes {
		rootBody = addScopeValueToBody(root, rootBody)
	}

	return f
}

// addScopeValueToBody writes a new block representing the scope value to the
// given body. This is especially useful for writing nested scope values.
func addScopeValueToBody(scope *NestedScope, body *hclwrite.Body) *hclwrite.Body {
	childBlock := body.AppendNewBlock("scope", []string{string(scope.Type), scope.Name})
	childBody := childBlock.Body()
	for _, grandchild := range scope.Children {
		childBody = addScopeValueToBody(grandchild, childBody)
	}
	return body
}
