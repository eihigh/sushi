package main

import "github.com/eihigh/sushi"

func main() {
	if err := sushi.Main(); err != nil {
		panic(err)
	}
}
