package requests

import "strconv"

type Integer string

func NewInteger(n int) Integer {
	return Integer(strconv.Itoa(n))
}

func (integer Integer) hasValue() bool {
	return integer != ""
}

func (integer Integer) getValue() (int, error) {
	return strconv.Atoi(string(integer))
}

type Boolean string

func NewBoolean(bool bool) Boolean {
	return Boolean(strconv.FormatBool(bool))
}

func (boolean Boolean) hasValue() bool {
	return boolean != ""
}

func (boolean Boolean) getValue() (bool, error) {
	return strconv.ParseBool(string(boolean))
}

type Float string

func NewFloat(f float64) Float {
	return Float(strconv.FormatFloat(f, 'f', 6, 64))
}

func (float Float) hasValue() bool {
	return float != ""
}

func (float Float) getValue() (float64, error) {
	return strconv.ParseFloat(string(float), 64)
}
