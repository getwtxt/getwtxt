package main

// content-type consts
const txtutf8 = "text/plain; charset=utf-8"
const htmlutf8 = "text/html; charset=utf-8"
const cssutf8 = "text/css; charset=utf-8"

// config object definition
type configuration struct {
	port          int
	logfile       string
	stdoutLogging bool
	instance
}

// refers to this specific instance of getwtxt
type instance struct {
	name  string
	url   string
	owner string
	mail  string
}
