package memory

import (
	"context"

	"lockbox.dev/clients"
)

type Factory struct{}

func (m Factory) NewStorer(ctx context.Context) (clients.Storer, error) {
	return NewStorer()
}

func (m Factory) TeardownStorers() error {
	return nil
}
