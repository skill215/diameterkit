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
// flags, scalar width (0 for Grouped), direct child count, reserved zero.
// Explicit opt-in asserts that required fields describe the complete minimum;
// custom rules with additional conditional presence must not opt in.
func minimumPlan(f Field, groups map[string]Group) (string, error) {
	var plan []byte
	var add func(Field, int) error
	add = func(field Field, depth int) error {
		if depth > 16 || len(plan) >= 64*12 {
			return fmt.Errorf("minimum diagnostic tree exceeds limit")
		}
		record := make([]byte, 12)
		binary.BigEndian.PutUint32(record, field.Vendor)
		binary.BigEndian.PutUint32(record[4:], field.Code)
		record[8] = field.Flags
		switch field.Type {
		case "u32", "i32", "time":
			record[9] = 4
		case "u64", "i64":
			record[9] = 8
		default:
			g, ok := groups[field.Type]
			if !ok || !g.DiagnosticMinimum || g.CaptureMandatory {
				return fmt.Errorf("unproven minimum group %s", field.Type)
			}
			for _, child := range g.Fields {
				if child.Min > 0 {
					if child.Min != 1 {
						return fmt.Errorf("minimum cardinality must be one")
					}
					record[10]++
				}
			}
			if record[10] == 0 {
				return fmt.Errorf("empty minimum group %s", field.Type)
			}
		}
		plan = append(plan, record...)
		if record[9] == 0 {
			for _, child := range groups[field.Type].Fields {
				if child.Min > 0 {
					if err := add(child, depth+1); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	if err := add(f, 1); err != nil {
		return "", err
	}
	return string(plan), nil
}
