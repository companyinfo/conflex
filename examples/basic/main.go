package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"go.companyinfo.dev/conflex"
	"go.companyinfo.dev/conflex/codec"
)

// Config is the configuration for the application.
type Config struct {
	Foo     string        `conflex:"foo"`
	Timeout time.Duration `conflex:"timeout"`
	Debug   bool          `conflex:"debug"`
	Worker  Worker        `conflex:"worker"`
	Date    time.Time     `conflex:"date"`
	Roles   []string      `conflex:"roles"`
	Types   []string      `conflex:"types"`
	Types2  string        `conflex:"types"`
}

// Worker is the worker configuration.
type Worker struct {
	Timeout time.Duration `conflex:"timeout"`
	Address *url.URL      `conflex:"address"`
}

// main is the main function.
func main() {

	var config Config

	c, err := conflex.New(
		conflex.WithFileSource("./config.yaml", codec.TypeYAML),
		conflex.WithBinding(&config),
	)
	if err != nil {
		log.Fatalf("failed to create conflex: %v", err)
	}

	err = c.Load(context.Background())
	if err != nil {
		log.Fatalf("failed to load conflex: %v", err)
	}

	fmt.Printf("%+v\n", config)

}
