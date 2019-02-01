package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"strings"
)

const workSettingsFilePath string = "/home/dev/.m2/settings-work.xml"
const homeSettingsFilePath string = "/home/dev/.m2/settings-home.xml"

func main() {

	multiline :=
		`Use environment variables true or false (defaults to false)
	 Environment variables supported:
	'default_maven_setting' for maven settings file path
	'home_maven_setting' for maven home settings file path
	'work_maven_setting' for maven work settings file path`

	boolPtr := flag.Bool("env", false, multiline)
	workSettingsFilePathPtr := flag.String("workPath", "", " -workPath=/home/dev/.m2/settings-work.xml")
	homeSettingsFilePathPtr := flag.String("homePath", "", "-homePath=/home/dev/.m2/settings-home.xml")
	defaultSettingsFilePathPtr := flag.String("settingsPath", "", "-settingsPath=/home/dev/.m2/settings.xml")
	workIPStartsWith := flag.String("workIPRange", "10.30", "-workIPRange=10.30")
	homeIPStartsWith := flag.String("homeIPRange", "192.168.1", "-homeIPRange=192.168.1")

	flag.Parse()

	ip, err := externalIP()
	CheckIfError(err)
	fmt.Printf("IP is %s \n", ip)

	// Assign Defaults
	workSettingsFilename := workSettingsFilePath
	homeSettingsFilename := homeSettingsFilePath

	usr, err := user.Current()
	CheckIfError(err)
	fmt.Printf("User's Home Directory : %s \n", usr.HomeDir)

	defaultSettingsFilename := usr.HomeDir + "/.m2/settings.xml"

	// Show variables used
	fmt.Println("Using system environment variables :", *boolPtr)
	if *boolPtr == false {
		if *workSettingsFilePathPtr != "" {
			workSettingsFilename = *workSettingsFilePathPtr
		}

		if *homeSettingsFilePathPtr != "" {
			homeSettingsFilename = *homeSettingsFilePathPtr
		}

		if *defaultSettingsFilePathPtr != "" {
			defaultSettingsFilename = *defaultSettingsFilePathPtr
		}

	} else {

		defaultSettingsFilename = os.Getenv("default_maven_setting")
		homeSettingsFilename = os.Getenv("home_maven_setting")
		workSettingsFilename = os.Getenv("work_maven_setting")
	}

	fmt.Println("Using settings.xml path =", defaultSettingsFilename)
	fmt.Println("Using work settings.xml path =", workSettingsFilename)
	fmt.Println("Using home settings.xml path =", homeSettingsFilename)

	// Check home settings file
	if _, err := os.Stat(homeSettingsFilename); os.IsNotExist(err) {
		msg := fmt.Sprintf("home settings file does not exist %s \n", homeSettingsFilename)
		fmt.Println(msg)
		CheckIfError(errors.New(msg))
	} else {
		fmt.Printf("home settings file exists - %s \n", homeSettingsFilename)
	}

	// Check work settings file
	if _, err := os.Stat(workSettingsFilename); os.IsNotExist(err) {
		msg := fmt.Sprintf("Work settings file does not exist %s \n", workSettingsFilename)
		fmt.Println(msg)
		CheckIfError(errors.New(msg))

	} else {
		fmt.Printf("Work settings file exists - %s \n", workSettingsFilename)
	}

	if strings.Contains(ip, *homeIPStartsWith) {
		fmt.Printf("IP contains %s - Home settings\n", *homeIPStartsWith)
		err := copyFiles(homeSettingsFilename, defaultSettingsFilename)
		if err != nil {
			fmt.Printf("File copying failed: %q\n", err)
			CheckIfError(err)
		}
	}

	if strings.Contains(ip, *workIPStartsWith) {
		fmt.Printf("IP contains %s - Work settings\n", *workIPStartsWith)
		err := copyFiles(workSettingsFilename, defaultSettingsFilename)
		if err != nil {
			fmt.Printf("File copying failed: %q\n", err)
			CheckIfError(err)
		}
	}

}

// source: https://opensource.com/article/18/6/copying-files-go
func copyFiles(source string, destination string) error {
	fmt.Printf("Copying from [%s] to [%s] \n", source, destination)
	sourceFile := source
	destinationFile := destination

	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		fmt.Println("File cannot be read", sourceFile)
		CheckIfError(err)
	}

	err = ioutil.WriteFile(destinationFile, input, 0644)
	if err != nil {
		fmt.Println("Error creating", destinationFile)
		CheckIfError(err)
	}

	return err
}

// source: https://play.golang.org/p/BDt3qEQ_2H
func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("Are you connected to the network?")
}

// CheckIfError - check for errors then exit
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}
