package main

import (
	"fmt"

	"fyne.io/fyne/v2/app"
)

func main() {
	fmt.Println("Init the app")
	myApp := app.New()
	myWindow := myApp.NewWindow("Hello")
	myWindow.ShowAndRun()
}
