// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package s6c composes S6c helpers with a caller-owned Route. It owns no timers,
// queues, storage, transports or recovery policy.
package s6c

import (
	"context"
	"errors"
	diameter "github.com/skill215/diameterkit"
	app "github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/validation"
)

type Client struct{ route *diameter.Route }
type RoutingReply struct {
	Exchange diameter.Result
	Routing  *app.RoutingResult
}

// New borrows a configured Route. Its frozen application/profile binding is
// enforced by Request/Submit. The client never changes destination or policy.
func New(route *diameter.Route) (*Client, error) {
	if route == nil {
		return nil, validation.ErrContext
	}
	return &Client{route: route}, nil
}
func (c *Client) QueryRouting(ctx context.Context, q app.RoutingQuery, opts diameter.RequestOptions) (RoutingReply, error) {
	r, err := app.NewSRR(q)
	if err != nil {
		return RoutingReply{}, err
	}
	if c == nil || c.route == nil {
		return RoutingReply{}, validation.ErrContext
	}
	result, err := c.route.Request(ctx, r, opts)
	if err != nil {
		return RoutingReply{Exchange: result}, err
	}
	reply, interpretationErr := c.InterpretRouting(app.RoutingContext{Request: r, Validation: opts.Validation}, result)
	return reply, errors.Join(err, interpretationErr)
}
func (c *Client) SubmitRouting(ctx context.Context, q app.RoutingQuery, opts diameter.RequestOptions, port *diameter.CompletionPort) (diameter.RequestID, error) {
	r, err := app.NewSRR(q)
	if err != nil {
		return diameter.RequestID{}, err
	}
	if c == nil || c.route == nil {
		return diameter.RequestID{}, validation.ErrContext
	}
	return c.route.Submit(ctx, r, opts, port)
}

// InterpretRouting is also the asynchronous path. It never consumes completions
// or publishes events. Exchange is preserved on every failure. Supply original
// request endpoints for named context, as required by app.InterpretSRA.
func (c *Client) InterpretRouting(query app.RoutingContext, completed diameter.Result) (RoutingReply, error) {
	reply := RoutingReply{Exchange: completed}
	if completed.Err != nil {
		return reply, completed.Err
	}
	if completed.Outcome != diameter.AnswerReceived || completed.Answer.Body == nil {
		return reply, validation.ErrInvalidAnswer
	}
	if !completed.Validation.Complete {
		return reply, completed.Validation.Err()
	}
	routing, err := app.InterpretSRA(query, completed.Answer.Body)
	if err != nil {
		return reply, err
	}
	reply.Routing = &routing
	return reply, nil
}
