package fake

import "github.com/go-faker/faker/v4"

func New[T any]() T {
	var t T
	if err := faker.FakeData(&t); err != nil {
		panic(err)
	}

	return t
}
