package main

import "fmt"

func main() {
	api := Gobbler{ApiKey: "xxx", Secret: "xxx"}
	success, _ := api.Login("xxx", "xxx")
	fmt.Println(success)
}
