package util

import "github.com/fatih/color"


var CreateColor func(a ...interface{}) string = color.New(color.FgGreen, color.Bold).SprintFunc()
var RenameColor func(a ...interface{}) string = color.New(color.FgHiYellow, color.Bold).SprintFunc()
var RemoveColor func(a ...interface{}) string = color.New(color.FgRed, color.Bold).SprintFunc()

