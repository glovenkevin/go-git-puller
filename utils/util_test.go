package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Empty string Should not be passed
// because it test if the contained string is stringonly
func TestCheckStringOnlyEmpty(t *testing.T) {
	res := CheckIsStringOnly("")
	assert.False(t, res)
}

// Check for the normal string
func TestCheckStringOnly(t *testing.T) {
	res := CheckIsStringOnly("asdf")
	assert.True(t, res)
}

func TestCheckStringOnlyNumber(t *testing.T) {
	res := CheckIsStringOnly("asdf123")
	assert.False(t, res)
}

func TestCheckStringOnlyNumber2(t *testing.T) {
	res := CheckIsStringOnly("123asdf")
	assert.False(t, res)
}

func TestCheckStringOnlyNumber3(t *testing.T) {
	res := CheckIsStringOnly("12asd12")
	assert.False(t, res)
}

func TestCheckStringOnlySpecialChar(t *testing.T) {
	res := CheckIsStringOnly("asd_*")
	assert.False(t, res)
}

func TestCheckStringOnlySpecialChar2(t *testing.T) {
	res := CheckIsStringOnly("123_*")
	assert.False(t, res)
}

func TestCheckStringOnlySpecialChar3(t *testing.T) {
	res := CheckIsStringOnly("//_*_*")
	assert.False(t, res)
}
