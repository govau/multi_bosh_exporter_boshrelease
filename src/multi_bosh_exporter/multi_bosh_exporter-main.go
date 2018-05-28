package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type config struct {
	Commands []*ourCmd
}

type ourCmd struct {
	Command string
	Args    []string
	Files   map[string]string
}

func (oc *ourCmd) RunForever(idx int) {
	for n, v := range oc.Files {
		err := ioutil.WriteFile(n, []byte(v), 0600)
		if err != nil {
			log.Fatalf("cannot start %d because cannot write to file %s because of %s", idx, n, err.Error())
		}
	}

	timeToSleep := time.Second
	for {
		cmd := exec.Command(oc.Command, oc.Args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		log.Printf("Attempting start %d...\n", idx)
		beforeTime := time.Now()
		err := cmd.Run()
		log.Printf("Completed %d with status %s\n", idx, err.Error())
		if time.Now().Sub(beforeTime).Seconds() > 60.0 {
			timeToSleep = time.Second
		} else {
			timeToSleep *= 2
		}
		log.Printf("Sleeping for %d for %d seconds...\n", idx, int(timeToSleep.Seconds()))
		time.Sleep(timeToSleep)
	}
}

func main() {
	var configPath string

	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.Parse()

	if configPath == "" {
		log.Fatal("must specify config path")
	}

	dd, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}

	var conf config
	err = yaml.UnmarshalStrict(dd, &conf)
	if err != nil {
		log.Fatal(err)
	}

	for idx, cmd := range conf.Commands {
		go cmd.RunForever(idx)
	}

	// Wait for ctrl-C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	<-c
	log.Println("Got signal, stopping")
}
