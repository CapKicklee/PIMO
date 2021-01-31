package constant

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/model"
)

func TestMaskingShouldReplaceSensitiveValueByConstantValue(t *testing.T) {
	maskingConstant := "Toto"
	constMask := NewMask(maskingConstant)
	result, err := constMask.Mask("Benjamin")
	assert.Equal(t, nil, err, "error should be nil")
	waited := "Toto"
	assert.Equal(t, waited, result, "should be Toto")
}

func TestFactoryShouldCreateAMask(t *testing.T) {
	maskingConfig := model.Masking{Mask: model.MaskType{Constant: "Toto"}}
	mask, present, err := Factory(maskingConfig, 0)
	waitedMask := NewMask("Toto")
	assert.Equal(t, waitedMask, mask, "should be equal")
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
