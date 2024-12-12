package FactoryMethod

type FruitFactory interface {
	CreateFruit() Fruit
}

type OrangeFactory struct {
}

func NewOrangeFactory() FruitFactory {
	return &OrangeFactory{}
}

func (o *OrangeFactory) CreateFruit() Fruit {
	return NewOrange("")
}

type StrawberryFactory struct {
}

func NewStrawberryFactory() FruitFactory {
	return &StrawberryFactory{}
}

func (s *StrawberryFactory) CreateFruit() Fruit {
	return NewStrawberry("")
}

type CherryFactory struct {
}

func NewCherryFactory() FruitFactory {
	return &CherryFactory{}
}

func (c *CherryFactory) CreateFruit() Fruit {
	return NewCherry("")
}

type WatermelonFactory struct {
}

func NewWatermelonFactory() FruitFactory {
	return &WatermelonFactory{}
}

func (w *WatermelonFactory) CreateFruit() Fruit {
	return NewWatermelon("")
}
