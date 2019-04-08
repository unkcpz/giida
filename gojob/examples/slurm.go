package main

import (
	"log"
	"os"
	// "fmt"
	"bytes"

	"github.com/scut-ccmp/flowmat/gojob"
	"github.com/spf13/viper"
)


func main() {

	var tomlExample = []byte(`
{
	"server": {
		"host": "202.38.220.15",
		"port": "22",
		"user": "unkcpz",
		"password": "sunshine"
	},
	"file": {
		"tempDir": "/home/unkcpz/giida",
		"dirPrefix": "tmp"
	}
}
	`)

	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.config/gojob/")
	viper.AddConfigPath(".")
	viper.SetConfigType("json")
	// err := viper.ReadInConfig()
	viper.ReadConfig(bytes.NewBuffer(tomlExample))
	// if err != nil { // Handle errors reading the config file
	// 	panic(fmt.Errorf("Fatal error config file: %s \n", err))
	// }

	user := viper.GetString("server.user")
	pass := viper.GetString("server.password")
	host := viper.GetString("server.host")
	port := viper.GetString("server.port")

	dir := viper.GetString("file.tempDir")
	prefix := viper.GetString("file.dirPrefix")

	// get host public key
	// hostKey := getHostKey(host)

	conn, err := gojob.NewConnect(user, pass, host, port)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	pathname, err := gojob.TempDir(conn.Client, dir, prefix)
	if err != nil {
		log.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// send files
	// bugs need execute mod
	err = gojob.SendFiles(conn.Client, wd, pathname)
	if err != nil {
		log.Fatal(err)
	}

	jobMgt := gojob.JobManager("slurm")
	// submit job
	jobID, err := jobMgt.SubmitJob(conn, pathname)
	if err != nil {
		log.Fatal(err)
	}

	gojob.Check(jobMgt.CheckDoneFunc, conn, jobID)

	// recive files
	err = gojob.ReciveFiles(conn.Client, pathname, wd)
	if err != nil {
		log.Fatal(err)
	}
}
