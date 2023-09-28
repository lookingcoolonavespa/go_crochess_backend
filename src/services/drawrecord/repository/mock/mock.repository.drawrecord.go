package repository_drawrecord_mock

import "github.com/stretchr/testify/mock"

type DrawRecordMockRepo struct {
	mock.Mock
}

func (c *DrawRecordMockRepo) Update(id int, version int, changes map[string]interface{}) (bool, error) {
	args := c.Called(id, version, changes)
	result := args.Get(0)

	return result.(bool), args.Error(1)
}
