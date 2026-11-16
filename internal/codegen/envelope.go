package codegen

// Envelope roles are validated against their actual Vendor/Code/type in validate.
func (w *writer) envelope(c Command) {
	w.p("func envelope%s(b pdu.Body)(pdu.Envelope,error){m,ok:=b.(*%s);if !ok||m==nil{return pdu.Envelope{},validation.ErrType};var out pdu.Envelope", c.Name, c.Name)
	var session, proxy, host, realm string
	for _, f := range c.Fields {
		switch f.Correlation {
		case "session":
			session = f.Name
			w.p("out.SessionID=avp.Some(m.%s)", f.Name)
		case "auth":
			w.p("out.AuthSessionState=avp.Some(m.%s)", f.Name)
		case "proxy":
			proxy = f.Name
			w.p("out.ProxyInfo=m.%s", f.Name)
		}
		if f.Envelope == "DestinationHost" || f.Envelope == "DestinationRealm" {
			if optional(f) {
				w.p("out.%s=m.%s", f.Envelope, f.Name)
			} else {
				w.p("out.%s=avp.Some(m.%s)", f.Envelope, f.Name)
			}
		}
		if f.Envelope == "OriginHost" {
			host = f.Name
			w.p("out.Origin.Host=m.%s", f.Name)
		}
		if f.Envelope == "OriginRealm" {
			realm = f.Name
			w.p("out.Origin.Realm=m.%s", f.Name)
		}
	}
	w.p("return out,nil\n}")
	if c.Request {
		return
	}
	w.p("func answer%s(b pdu.Body,env pdu.AnswerEnvelope)(pdu.Body,error){m,ok:=b.(*%s);if !ok||m==nil{return nil,validation.ErrType};out:=clone%s(m)", c.Name, c.Name, c.Name)
	w.p("var session avp.Optional[string];var host,realm string;var proxy []pdu.ProxyInfo")
	if session != "" {
		w.p("if out.%s!=\"\"{session=avp.Some(out.%s)}", session, session)
	}
	if host != "" {
		w.p("host=out.%s", host)
	} else {
		w.p("if env.Origin.Host!=\"\"{return nil,validation.ErrEnvelope}")
	}
	if realm != "" {
		w.p("realm=out.%s", realm)
	} else {
		w.p("if env.Origin.Realm!=\"\"{return nil,validation.ErrEnvelope}")
	}
	if proxy != "" {
		w.p("proxy=out.%s", proxy)
	}
	sp, pp := "nil", "nil"
	if session != "" {
		sp = "&session"
	}
	if proxy != "" {
		pp = "&proxy"
	}
	// Keep the generated locals usable even for a deliberately minimal extension.
	w.p("_ = session;_ = proxy")
	w.p("if err:=%sCompleteEnvelopeValues(%s,&host,&realm,%s,env);err!=nil{return nil,err}", w.prefix, sp, pp)
	if session != "" {
		w.p("out.%s=session.Value", session)
	}
	if host != "" {
		w.p("out.%s=host", host)
	}
	if realm != "" {
		w.p("out.%s=realm", realm)
	}
	if proxy != "" {
		w.p("out.%s=proxy", proxy)
	}
	w.p("return out,nil\n}")
}
