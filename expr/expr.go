package expr

import (
	"github.com/tomo3110/gerbera"
)

type CallbackFunc func(gerbera.ConvertToMap) gerbera.ComponentFunc

func If(expr bool, trueCF gerbera.ComponentFunc, otherCF ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return func(parent *gerbera.Element) error {
		if expr {
			if err := trueCF(parent); err != nil {
				return err
			}
		} else {
			for _, ef := range otherCF {
				if err := ef(parent); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func Unless(expr bool, falseCF gerbera.ComponentFunc, otherCF ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return func(parent *gerbera.Element) error {
		if !expr {
			if err := falseCF(parent); err != nil {
				return err
			}
		} else {
			for _, ef := range otherCF {
				if err := ef(parent); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func Each(list []gerbera.ConvertToMap, callback CallbackFunc) gerbera.ComponentFunc {
	return func(parent *gerbera.Element) error {
		for _, item := range list {
			if err := callback(item)(parent); err != nil {
				return err
			}
		}
		return nil
	}
}
