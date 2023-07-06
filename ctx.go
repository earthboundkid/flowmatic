package flowmatic

import "context"

type CancelGroup struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func WithContext(ctx context.Context) CancelGroup {
	var cg CancelGroup
	cg.ctx, cg.cancel = context.WithCancel(ctx)
	return cg
}

func (cg CancelGroup) Race(fn func(context.Context) error) func() error {
	return func() error {
		if cg.ctx.Err() != nil {
			return nil
		}
		if err := fn(cg.ctx); err != nil {
			return err
		}
		cg.cancel()
		return nil
	}
}

func (cg CancelGroup) All(fn func(context.Context) error) func() error {
	return func() error {
		if cg.ctx.Err() != nil {
			return cg.ctx.Err()
		}
		if err := fn(cg.ctx); err != nil {
			cg.cancel()
			return err
		}
		return nil
	}
}

func (cg CancelGroup) Done() {
	cg.cancel()
}
