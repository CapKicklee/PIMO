package regex

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"makeit.imfr.cgi.com/makeit2/scm/lino/pimo/pkg/model"
)

func TestMaskingShouldReplaceSensitiveValueByRegex(t *testing.T) {
	regex := "0[1-7]( ([0-9]){2}){4}"
	regmask, err := NewMask(regex, 0)
	if err != nil {
		assert.Fail(t, "could not initialise regexep")
	}

	data := "00 00 00 00 00"
	result, err := regmask.Mask(data)
	assert.Equal(t, nil, err, "error should be nil")
	match, _ := regexp.MatchString(regex, result.(string))
	assert.NotEqual(t, data, result, "should be masked")
	assert.True(t, match, "should match the regexp")
}

func TestFactoryShouldCreateAMask(t *testing.T) {
	maskingConfig := model.Masking{Mask: model.MaskType{Regex: "[A-Z]oto([a-z]){3}"}}
	_, present, err := Factory(maskingConfig, 0)
	assert.Nil(t, err, "error should be nil")
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
