package expr

import (
	"github.com/tomo3110/gerbera"
)

type CallbackFunc func(gerbera.ConvertToMap) gerbera.ComponentFunc

func If(expr bool, trueCF gerbera.ComponentFunc, otherCF ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return func(parent gerbera.Node) {
		if expr {
			trueCF(parent)
		} else {
			for _, ef := range otherCF {
				ef(parent)
			}
		}
	}
}

func Unless(expr bool, falseCF gerbera.ComponentFunc, otherCF ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return func(parent gerbera.Node) {
		if !expr {
			falseCF(parent)
		} else {
			for _, ef := range otherCF {
				ef(parent)
			}
		}
	}
}

func Each(list []gerbera.ConvertToMap, callback CallbackFunc) gerbera.ComponentFunc {
	return func(parent gerbera.Node) {
		for _, item := range list {
			callback(item)(parent)
		}
	}
}
