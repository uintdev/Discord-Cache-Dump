package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/h2non/filetype"
	"github.com/ricochet2200/go-disk-usage/du"
)

const (
	softVersion = 1.0
	dumpDir     = "dump"
)

var platform = runtime.GOOS

// Add extra line of space when the program exits
func exitNewLine() string {
	var endOfProgram string
	if platform == "windows" {
		endOfProgram = "\n"
	} else {
		endOfProgram = "\n\n"
	}
	return endOfProgram
}

// Check for root, as running the program with it can change the path
func rootCheck(tuid int) {
	if tuid == 0 {
		if platform != "darwin" {
			fmt.Print("[NOTICE] This program is running as root\n")
			fmt.Print("[...] The logged in user will be used.\n\n")
		}
	} else {
		if platform == "darwin" {
			fmt.Printf("[ERROR] Due to file permissions, this must be ran as root on macOS%s", exitNewLine())
			os.Exit(1)
		}
	}
}

// Set a date and time
func timeDate() string {
	timeDat := time.Now().Format("2006-01-02--15-04-05")
	return timeDat
}

// Initialisation of file stats variables
var unreadableRes int64
var overallSize int64
var spareStorage int64

// Copy source file to destination
func copyFile(from string, to string, permuid int) {
	input, err := ioutil.ReadFile(from)
	if err != nil {
		/*
			We simply assume that the file it got to cannot be 'opened' because Discord's process is using it at the time.
			Yes, as any other error could occur for whatever reason, it is not the best way to go but it's good enough.
		*/
		unreadableRes++
		return
	}

	err = ioutil.WriteFile(to, input, 0644)
	if err != nil {
		fmt.Printf("\n[ERROR] Write error: %s\n", err)
		return
	}
	// If ran as root as a sudoer on Linux or macOS, it's going to be root:root, so we change it to what it really should be
	if platform != "windows" {
		os.Chown(to, permuid, permuid)
	}
}

// Get file size and exclude unreadable files
func sizeStore(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	overallSize += int64(len(data))
	return
}

// Amount of free space on the partition the program is running off of (in bytes)
func freeStorage(path string) {
	usage := du.NewDiskUsage(path)
	free := usage.Free()
	spareStorage = int64(free)
	return
}

// Initialisation of system information variables
var uid int
var userName string
var homePath string
var sudoerUID int

func main() {
	fmt.Print("\n")
	// Banner
	fmt.Print("#####################################\n")
	fmt.Printf("# Discord Cache Dump :: Version %0.1f #\n", softVersion)
	fmt.Print("#####################################\n\n")

	user, err := user.Current()
	if err != nil {
		fmt.Printf("[ERROR] Failed to obtain user: %s%s", err, exitNewLine())
		os.Exit(1)
	} else {
		if platform != "windows" {
			uidConv, err := strconv.Atoi(user.Uid)
			if err != nil {
				fmt.Printf("[ERROR] Unable to obtain UID%s", exitNewLine())
				os.Exit(1)
			}
			uid = uidConv
		} else {
			uid = -1
		}
		userName = user.Username
		homePath = user.HomeDir
	}

	if platform != "windows" {
		// Get the sudoer user
		userNamei, ok := os.LookupEnv("SUDO_USER")
		if !ok {
			userName = user.Username
		} else {
			userName = userNamei
		}

		// Get the sudoer UID
		sudoerUIDi, ok := os.LookupEnv("SUDO_UID")
		if !ok {
			sudoerUID = uid
		} else {
			sudoerUID, err = strconv.Atoi(sudoerUIDi)
			if err != nil {
				fmt.Printf("[ERROR] Unable to convert UID to INT%s", exitNewLine())
				os.Exit(1)
			}
		}

		// Correct the path if logged in user isn't root
		if sudoerUID != 0 {
			if platform == "darwin" {
				homePath = "/Users/" + userName
			} else if platform == "linux" {
				// This assumes that the user was not assigned another 'home' directory
				homePath = "/home/" + userName
			}
		}
	}

	// Check the platform in use and adjust as appropriate
	if platform == "linux" || platform == "darwin" {
		rootCheck(uid)
	} else if platform == "windows" {
		// Windows does not return just the username, so we split and grab what we need
		userName = strings.Split(userName, "\\")[1]
	} else {
		fmt.Printf("[ERROR] Unsupported platform: %s%s", platform, exitNewLine())
		os.Exit(1)
	}

	fmt.Printf("Logged in as: %s\n\n", userName)

	fmt.Print("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
	fmt.Print("!! A few files (index, data_0-3) are in use by Discord while    !!\n")
	fmt.Print("!! it is running. To also dump those files, kill all instances  !!\n")
	fmt.Print("!! of Discord before continuing. There will be a notice         !!\n")
	fmt.Print("!! regarding this per installed Discord build where applicable  !!\n")
	fmt.Print("!! while copying the files over.                                !!\n")
	fmt.Print("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n\n")

	fmt.Print("Press enter to continue ...\n")
	fmt.Scanln()

	// Discord client build names
	discordBuildName := map[int]string{
		0: "Stable",
		1: "PTB",
		2: "Canary",
		3: "Development",
	}
	// Discord client build folder names
	discordBuildDir := map[int]string{
		0: "discord",
		1: "discordptb",
		2: "discordcanary",
		3: "discorddevelopment",
	}
	// Cache paths for each platform
	cachePath := map[string]string{
		"linux":   "%s/.config/%s/Cache/",
		"darwin":  "%s/Library/Application Support/%s/Cache/",
		"windows": "%s\\AppData\\Roaming\\%s\\Cache\\",
	}

	// Initialise a few maps
	pathStatus := make(map[int]bool)
	cachedFile := make(map[int]map[int]string)

	fmt.Print("Checking for existing cache directories ...\n\n")

	// Check if directories exist
	for i := 0; i < len(discordBuildDir); i++ {
		filePath := fmt.Sprintf(cachePath[platform], homePath, discordBuildDir[i])
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			fmt.Printf("Found: Discord %s\n", discordBuildName[i])
			pathStatus[i] = true
		} else {
			pathStatus[i] = false
		}
	}

	// Check if the directories are empty and store names of cached files
	for i := 0; i < len(discordBuildDir); i++ {
		if pathStatus[i] {
			filePath := fmt.Sprintf(cachePath[platform], homePath, discordBuildDir[i])
			cacheListing, err := ioutil.ReadDir(filePath)
			if err != nil {
				fmt.Printf("[ERROR] Unable to read directory for Discord %s%s", discordBuildName[i], exitNewLine())
				os.Exit(1)
			}
			// Go through the list of files present in a cache directory
			if len(cacheListing) > 0 {
				cachedFile[i] = make(map[int]string)
				for k, v := range cacheListing {
					cachedFile[i][k] = v.Name()
					sizeStore(filePath + v.Name())
				}
				fmt.Printf("Discord %s :: found %d cached files\n", discordBuildName[i], len(cacheListing))
			} else {
				fmt.Printf("Discord %s :: cache empty, skipping\n", discordBuildName[i])
				pathStatus[i] = false
			}
		}
	}

	// If there are still no results at all, there is no point in continuing
	pathStatusSuccessCount := 0
	for i := 0; i < len(discordBuildDir); i++ {
		if pathStatus[i] {
			pathStatusSuccessCount++
		}
	}
	if pathStatusSuccessCount == 0 {
		fmt.Printf("No cache found%s", exitNewLine())
		os.Exit(0)
	} else {
		fmt.Print("\n")
	}

	// Check storage requirements
	curDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("[ERROR] Unable to obtain current directory%s", exitNewLine())
		os.Exit(1)
	}
	freeStorage(curDir)
	if spareStorage <= overallSize {
		remainingStorage := spareStorage - overallSize
		requiredSpace := strings.Replace(fmt.Sprintf("%v", remainingStorage), "-", "", -1)
		fmt.Print("[ERROR] Insufficient storage where program is being ran\n")
		fmt.Printf("[...] %s bytes need sparing%s", requiredSpace, exitNewLine())
		os.Exit(1)
	} else {
		fmt.Print("Sufficient storage -- safe to go ahead\n\n")
	}

	// Check and create dump directory structure
	timeDateStamp := timeDate()
	if _, err := os.Stat(dumpDir + "/"); os.IsNotExist(err) {
		os.Mkdir(dumpDir, 0644)
		if platform != "windows" {
			os.Chown(dumpDir, sudoerUID, sudoerUID)
		}
	}
	if _, err := os.Stat(dumpDir + "/" + timeDateStamp + "/"); os.IsNotExist(err) {
		os.Mkdir(dumpDir+"/"+timeDateStamp, 0644)
		if platform != "windows" {
			os.Chown(dumpDir+"/"+timeDateStamp, sudoerUID, sudoerUID)
		}
	}

	// Create any Discord client directories that are required
	for i := 0; i < len(discordBuildDir); i++ {
		if len(cachedFile[i]) > 0 {
			if _, err := os.Stat(dumpDir + "/" + timeDateStamp + "/" + discordBuildName[i] + "/"); os.IsNotExist(err) {
				os.Mkdir(dumpDir+"/"+timeDateStamp+"/"+discordBuildName[i], 0644)
				if platform != "windows" {
					os.Chown(dumpDir+"/"+timeDateStamp+"/"+discordBuildName[i], sudoerUID, sudoerUID)
				}
			}

			// Copy the files over
			fmt.Printf("Copying %d files from Discord %s ...\n", len(cachedFile[i]), discordBuildName[i])
			for it := 0; it < len(cachedFile[i]); it++ {
				// Build the paths to use during the copy operation
				fromPath := fmt.Sprintf(cachePath[platform], homePath, discordBuildDir[i])
				toPath := dumpDir + "/" + timeDateStamp + "/" + discordBuildName[i] + "/" + cachedFile[i][it]
				// Copying the files one-by-one
				copyFile(fromPath+cachedFile[i][it], toPath, sudoerUID)
			}

			// Unable to copy client-critial cache
			if unreadableRes > 0 {
				fmt.Printf("[NOTICE] Cannot read client-critial cache while Discord %s is running\n", discordBuildName[i])
				fmt.Printf("[...] Unable to read %d client-critial cache file(s)\n", unreadableRes)
				unreadableResCount := int64(len(cachedFile[i])) - unreadableRes
				fmt.Printf("[...] Actually copied %d cache files from Discord %s\n", unreadableResCount, discordBuildName[i])
			}
		}
	}

	// Analyse and rename files
	fmt.Print("\n")
	fmt.Print("Analysing copied files ...\n")

	for i := 0; i < len(discordBuildDir); i++ {
		if len(cachedFile[i]) > 0 {
			fmt.Printf("Changing file extensions for Discord %s cache ...\n", discordBuildName[i])
			var identificationCount = 0
			for it := 0; it < len(cachedFile[i]); it++ {
				cachedFilePath := dumpDir + "/" + timeDateStamp + "/" + discordBuildName[i] + "/" + cachedFile[i][it]
				buf, err := ioutil.ReadFile(cachedFilePath)
				if err == nil {
					kind, _ := filetype.Match(buf)
					if kind != filetype.Unknown {
						// Rename files with an appended file extension if we know the file type
						os.Rename(cachedFilePath, cachedFilePath+"."+kind.Extension)
						identificationCount++
					}
				}
			}
			fmt.Printf("%d out of %d identified for Discord %s\n\n", identificationCount, len(cachedFile[i]), discordBuildName[i])
		}
	}

	// Let user know where to get to the files
	if platform == "windows" {
		fmt.Printf("Saved: %s\\%s\\%s%s", curDir, dumpDir, timeDateStamp, exitNewLine())
	} else {
		fmt.Printf("Saved: %s/%s/%s%s", curDir, dumpDir, timeDateStamp, exitNewLine())
	}

}
