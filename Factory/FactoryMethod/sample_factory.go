package FactoryMethod

import (
	"fmt"
)

// Fruit 定义了一个接口，用于表示可以被食用的水果。
type Fruit interface {
	Eat()
}

// Orange 表示一个橙子实体，包括橙子的名称。
type Orange struct {
	name string
}

// NewOrange 创建一个新的橙子实例，并返回作为Fruit接口的实现。
func NewOrange(name string) Fruit {
	return &Orange{
		name: name,
	}
}

// Eat 实现了Fruit接口，打印出橙子被吃的信息。
func (o *Orange) Eat() {
	fmt.Printf("i am orange: %s, i am about to be eaten...\n", o.name)
}

// Strawberry 表示一个草莓实体，包括草莓的名称。
type Strawberry struct {
	name string
}

// NewStrawberry 创建一个新的草莓实例，并返回作为Fruit接口的实现。
func NewStrawberry(name string) Fruit {
	return &Strawberry{
		name: name,
	}
}

// Eat 实现了Fruit接口，打印出草莓被吃的信息。
func (s *Strawberry) Eat() {
	fmt.Printf("i am strawberry: %s, i am about to be eaten...\n", s.name)
}

// Cherry 表示一个樱桃实体，包括樱桃的名称。
type Cherry struct {
	name string
}

// NewCherry 创建一个新的樱桃实例，并返回作为Fruit接口的实现。
func NewCherry(name string) Fruit {
	return &Cherry{
		name: name,
	}
}

// Eat 实现了Fruit接口，打印出樱桃被吃的信息。
func (c *Cherry) Eat() {
	fmt.Printf("i am cherry: %s, i am about to be eaten...\n", c.name)
}

type Watermelon struct {
	name string
}

func NewWatermelon(name string) Fruit {
	return &Watermelon{
		name: name,
	}
}

func (c *Watermelon) Eat() {
	fmt.Printf("i am watermelon: %s, i am about to be eaten...\n", c.name)
}
