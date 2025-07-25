package main

import "flag"

func main() {
	var confPath = flag.String("c", "./conf/media.toml", "config path")
	flag.Parse()
}
