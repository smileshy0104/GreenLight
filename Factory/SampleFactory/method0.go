package SampleFactory

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// FruitFactory 定义了一个水果工厂，用于创建各种类型的水果。
type FruitFactory0 struct {
}

// NewFruitFactory 创建一个新的水果工厂实例。
func NewFruitFactory0() *FruitFactory0 {
	return &FruitFactory0{}
}

// CreateFruit 根据给定的类型创建一个水果实例。
// 它支持创建"orange"、"strawberry"和"cherry"类型的水果。
// 如果给定的类型不支持，将返回一个错误。
func (f *FruitFactory0) CreateFruit0(typ string) (Fruit, error) {
	src := rand.NewSource(time.Now().UnixNano())
	rander := rand.New(src)
	name := strconv.Itoa(rander.Int())

	switch typ {
	case "orange":
		return NewOrange(name), nil
	case "strawberry":
		return NewStrawberry(name), nil
	case "cherry":
		return NewCherry(name), nil
	default:
		return nil, fmt.Errorf("fruit typ: %s is not supported yet", typ)
	}
}
