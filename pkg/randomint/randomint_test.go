package randomint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/model"
)

func TestMaskingShouldReplaceSensitiveValueByRandomNumber(t *testing.T) {
	min := 7
	max := 77
	ageMask := NewMask(min, max, 0)

	result, err := ageMask.Mask(83)
	assert.Equal(t, nil, err, "error should be nil")

	assert.NotEqual(t, 83, result, "Should be masked")
	assert.True(t, result.(int) >= min, "Should be more than min")
	assert.True(t, result.(int) <= max, "Should be less than max")
}

func TestFactoryShouldCreateAMask(t *testing.T) {
	maskingConfig := model.Masking{Mask: model.MaskType{RandomInt: model.RandIntType{Min: 18, Max: 25}}}
	mask, present, err := Factory(maskingConfig, 0)
	assert.NotNil(t, mask, "shouldn't be nil")
	assert.True(t, present, "should be true")
	assert.Nil(t, err, "error should be nil")
}

func TestFactoryShouldNotCreateAMaskFromAnEmptyConfig(t *testing.T) {
	maskingConfig := model.Masking{Mask: model.MaskType{}}
	mask, present, err := Factory(maskingConfig, 0)
	assert.Nil(t, mask, "should be nil")
	assert.False(t, present, "should be false")
	assert.Nil(t, err, "error should be nil")
}
