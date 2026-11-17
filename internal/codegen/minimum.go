// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Compile explicitly reviewed minimum Grouped diagnostic grammar.
// SPDX-AI-Date: 2026-11-17

package codegen

import (
	"encoding/binary"
	"fmt"
)

// A plan is a bounded preorder sequence of 12-byte records: vendor, code,
// flags, scalar width, direct child count, primitive kind (0 for fixed/Grouped).
// V2 distinguishes a zero-width variable leaf from a Grouped record.
// Explicit opt-in asserts that required fields describe the complete minimum;
// custom rules with additional conditional presence must not opt in.
func minimumPlan(f Field, groups map[string]Group) (string, error) {
	var plan []byte
	var add func(Field, int) (bool, error)
	add = func(field Field, depth int) (bool, error) {
		if depth > 16 || len(plan) >= 64*12 {
			return false, fmt.Errorf("minimum diagnostic tree exceeds limit")
		}
		record := make([]byte, 12)
		binary.BigEndian.PutUint32(record, field.Vendor)
		binary.BigEndian.PutUint32(record[4:], field.Code)
		record[8] = field.Flags
		variable := false
		switch field.Type {
		case "u32", "i32", "time":
			record[9] = 4
		case "u64", "i64":
			record[9] = 8
		case "identity", "tbcd", "utf8", "octets":
			kind, width, ok := variableMinimum(field)
			if !ok || field.DiagnosticMinimumBytes == nil || *field.DiagnosticMinimumBytes != width {
				return false, fmt.Errorf("unproven variable minimum %s", field.Name)
			}
			record[9] = byte(width)
			record[11] = map[string]byte{"identity": 1, "tbcd": 2, "utf8": 3, "utf8-nonempty": 4, "octets": 5, "octets-nonempty": 6}[kind]
			variable = true
		default:
			g, ok := groups[field.Type]
			if !ok || !g.DiagnosticMinimum || g.CaptureMandatory {
				return false, fmt.Errorf("unproven minimum group %s", field.Type)
			}
			for _, child := range g.Fields {
				if child.Min > 0 {
					if child.Min != 1 {
						return false, fmt.Errorf("minimum cardinality must be one")
					}
					record[10]++
				}
			}
			if record[10] == 0 {
				return false, fmt.Errorf("empty minimum group %s", field.Type)
			}
		}
		plan = append(plan, record...)
		if record[10] > 0 {
			g := groups[field.Type]
			for _, child := range g.Fields {
				if child.Min > 0 {
					v, err := add(child, depth+1)
					if err != nil {
						return false, err
					}
					variable = variable || v
				}
			}
			if variable && g.Rule != "" && g.DiagnosticMinimumRule != g.Rule {
				return false, fmt.Errorf("variable minimum requires explicit rule audit for %s", g.Name)
			}
		}
		return variable, nil
	}
	if _, err := add(f, 1); err != nil {
		return "", err
	}
	return string(plan), nil
}
