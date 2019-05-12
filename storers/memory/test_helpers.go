package memory

import (
	"context"

	"impractical.co/auth/clients"
)

type Factory struct{}

func (m Factory) NewStorer(ctx context.Context) (clients.Storer, error) {
	return NewStorer()
}

func (m Factory) TeardownStorers() error {
	return nil
}
