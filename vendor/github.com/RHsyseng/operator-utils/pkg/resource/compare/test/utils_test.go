package test

import (
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmptyArray(t *testing.T) {
	builder := compare.NewMapBuilder()
	assert.Empty(t, builder.Map(), "Expected empty map")
}

//
//func TestSameTypeItems(t *testing.T) {
//	builder := compare.NewMapBuilder()
//
//}
