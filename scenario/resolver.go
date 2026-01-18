package scenario

import (
	"fmt"
	"regexp"
	"strings"
)

// CrossDomainRef represents a parsed cross-domain reference.
// References follow the syntax: ${domain.resource.outputs.field}
type CrossDomainRef struct {
	// Domain is the source domain name (e.g., "aws")
	Domain string

	// Resource is the resource identifier (e.g., "s3")
	Resource string

	// Field is the output field name (e.g., "bucket_name")
	Field string

	// Raw is the original reference string
	Raw string
}

// refPattern matches cross-domain references like ${domain.resource.outputs.field}
var refPattern = regexp.MustCompile(`\$\{([^.]+)\.([^.]+)\.outputs\.([^}]+)\}`)

// ParseRef parses a cross-domain reference string into a CrossDomainRef.
// The reference must follow the syntax: ${domain.resource.outputs.field}
//
// Examples:
//   - "${aws.s3.outputs.bucket_name}" -> {Domain: "aws", Resource: "s3", Field: "bucket_name"}
//   - "${gitlab.pipeline.outputs.project_id}" -> {Domain: "gitlab", Resource: "pipeline", Field: "project_id"}
//
// Returns an error if the reference format is invalid.
func ParseRef(ref string) (*CrossDomainRef, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("reference string is empty")
	}

	matches := refPattern.FindStringSubmatch(ref)
	if matches == nil || len(matches) != 4 {
		return nil, fmt.Errorf("invalid reference format: %s (expected: ${domain.resource.outputs.field})", ref)
	}

	return &CrossDomainRef{
		Domain:   matches[1],
		Resource: matches[2],
		Field:    matches[3],
		Raw:      ref,
	}, nil
}

// String returns a string representation of the reference.
func (r *CrossDomainRef) String() string {
	return r.Raw
}

// FindRefsInString finds all cross-domain references in a string.
// This can be used to extract references from configuration files, prompts, etc.
func FindRefsInString(s string) []*CrossDomainRef {
	var refs []*CrossDomainRef

	matches := refPattern.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if len(match) == 4 {
			refs = append(refs, &CrossDomainRef{
				Domain:   match[1],
				Resource: match[2],
				Field:    match[3],
				Raw:      match[0],
			})
		}
	}

	return refs
}
