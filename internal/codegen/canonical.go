package codegen

func (w *writer) canonicalFields(d Definition, fields []Field) {
	w.p("[]wire.CanonicalField{")
	for _, f := range fields {
		w.p("{Key:%s,Identity:%t,", key(f), f.Type == "identity")
		switch f.Type {
		case "proxy":
			w.p("Group:[]wire.CanonicalField{{Key:avp.Key{Code:280},Identity:true},{Key:avp.Key{Code:33}}},")
		case "vsai":
			w.p("Group:[]wire.CanonicalField{{Key:avp.Key{Code:266}},{Key:avp.Key{Code:258}},{Key:avp.Key{Code:259}}},")
		default:
			for _, g := range d.Groups {
				if g.Name == f.Type {
					if g.CaptureMandatory {
						w.p("PreserveMandatory:true,")
					}
					w.p("Group:")
					w.canonicalFields(d, g.Fields)
				}
			}
		}
		w.p("},")
	}
	w.p("},")
}
