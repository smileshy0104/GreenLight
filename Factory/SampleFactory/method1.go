package SampleFactory

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// fruitCreator 是一个函数类型，用于创建 Fruit 实例。
// 它接受一个字符串参数 name，并返回一个 Fruit 实例。
type fruitCreator func(name string) Fruit

// FruitFactory 是一个工厂结构体，用于根据类型创建不同的 Fruit 实例。
type FruitFactory struct {
	// creators 是一个映射，将水果名称映射到对应的创建函数。
	creators map[string]fruitCreator
}

// NewFruitFactory 创建并返回一个新的 FruitFactory 实例。
// 该实例预定义了如何创建某些类型水果的映射。
func NewFruitFactory() *FruitFactory {
	return &FruitFactory{
		creators: map[string]fruitCreator{
			"orange":     NewOrange,
			"strawberry": NewStrawberry,
			"cherry":     NewCherry,
		},
	}
}

// CreateFruit 根据给定的类型 typ 创建并返回一个 Fruit 实例。
// 如果给定的类型不受支持，则返回一个错误。
func (f *FruitFactory) CreateFruit(typ string) (Fruit, error) {
	// 尝试从映射中获取对应类型的水果创建函数。
	fruitCreator, ok := f.creators[typ]
	if !ok {
		// 如果类型不受支持，返回一个错误。
		return nil, fmt.Errorf("fruit typ: %s is not supported yet", typ)
	}

	// 使用当前时间生成一个随机源，以确保随机性。
	src := rand.NewSource(time.Now().UnixNano())
	// 基于随机源创建一个随机数生成器。
	rander := rand.New(src)
	// 生成一个随机名称，确保每个水果实例的唯一性。
	name := strconv.Itoa(rander.Int())
	// 使用获取到的创建函数生成水果实例，并返回。
	return fruitCreator(name), nil
}
