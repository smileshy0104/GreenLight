package SampleFactory

import "testing"

func Test_factory0(t *testing.T) {
	// 构造工厂
	fruitFactory := NewFruitFactory0()

	// 尝个橘子
	orange, _ := fruitFactory.CreateFruit0("orange")
	orange.Eat()

	// 来颗樱桃
	cherry, _ := fruitFactory.CreateFruit0("cherry")
	cherry.Eat()

	// 来个西瓜，因为未实现会报错
	watermelon, err := fruitFactory.CreateFruit0("watermelon")
	if err != nil {
		t.Error(err)
		return
	}
	watermelon.Eat()
}
