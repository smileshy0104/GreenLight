package FactoryMethod

import "testing"

func Test_factory(t *testing.T) {
	// 尝个橘子
	orangeFactory := NewOrangeFactory()
	orange := orangeFactory.CreateFruit()
	orange.Eat()

	// 来颗樱桃
	cherryFactory := NewCherryFactory()
	cherry := cherryFactory.CreateFruit()
	cherry.Eat()

	// 来颗西瓜
	watermelonFactory := NewWatermelonFactory()
	watermelon := watermelonFactory.CreateFruit()
	watermelon.Eat()
}
