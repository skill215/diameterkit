// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Explicit variable-field diagnostic minimum validation.
// SPDX-AI-Date: 2026-11-17

package codegen

// Diagnostic widths describe existing generated type/Nonempty constraints.
// They do not add a value constraint or infer a handwritten domain rule.
func variableMinimum(f Field) (string, int, bool) {
	if f.Min != 1 {
		return "", 0, false
	}
	switch f.Type {
	case "identity", "tbcd":
		return f.Type, 1, true
	case "utf8", "octets":
		if f.Nonempty {
			return f.Type + "-nonempty", 1, true
		}
		return f.Type, 0, true
	}
	return "", 0, false
}
