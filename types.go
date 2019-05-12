package main

// content-type consts
const txtutf8 = "text/plain; charset=utf8"
const htmlutf8 = "text/html; charset=utf8"

// config object definition
type configuration struct {
	port          int
	logfile       string
	twtxtfile     string
	stdoutLogging bool
}
