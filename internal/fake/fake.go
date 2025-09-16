package fake

import (
	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
)

func New[T any](opt ...options.OptionFunc) T {
	var obj T
	opt = append([]options.OptionFunc{
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 2
			oo.RandomMaxSliceSize = 5
		},
	}, opt...)

	err := faker.FakeData(&obj, opt...)
	if err != nil {
		panic(err)
	}

	return obj
}
