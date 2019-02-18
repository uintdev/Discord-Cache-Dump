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
	softVersion = 0.1
	dumpDir     = "dump"
)

var platform = runtime.GOOS

// Check for root, as running the program with it can change the path
func rootCheck(tuid int) {
	if tuid == 0 {
		if platform != "darwin" {
			fmt.Print("[WARN] This program is running as root\n")
			fmt.Print("[...] Unless you are logged in as root, this will not check with the actual user you have logged in as.\n\n")
		}
	} else {
		if platform == "darwin" {
			fmt.Print("[ERROR] Due to file permissions, this must be ran as root on macOS\n")
			os.Exit(1)
		}
	}
}

// Set a date and time
func timeDate() string {
	timeDat := time.Now().Format("2006-01-02--15-04-05")
	return timeDat
}

var unreadableResources = 0
var overallSize int64
var spareStorage int64

// Copy source file to destination
func copyFile(from string, to string) {
	input, err := ioutil.ReadFile(from)
	if err != nil {
		unreadableResources++
		return
	}

	err = ioutil.WriteFile(to, input, 0644)
	if err != nil {
		fmt.Printf("\n[ERROR] Write error: %s\n", err)
		return
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

var uid int
var userName string
var homePath string
var sudoerUID int

func main() {

	fmt.Print("\n")
	fmt.Printf("Discord Cache Dump :: Version %0.1f :: TESTER'S RELEASE 3\n\n", softVersion)

	user, err := user.Current()
	if err != nil {
		if platform != "darwin" {
			fmt.Printf("[ERROR] Failed to obtain user: %s\n", err)
			os.Exit(1)
		}
	} else {
		if platform != "windows" {
			uidconv, err := strconv.Atoi(user.Uid)
			if err != nil {
				fmt.Print("[ERROR] Unable to obtain UID\n")
				os.Exit(1)
			}
			uid = uidconv
		} else {
			uid = -1
		}
		userName = user.Username
		homePath = user.HomeDir
	}

	if platform != "windows" {
		if uid != 0 {
			userName = user.Username
		}

		userName, ok := os.LookupEnv("SUDO_USER")
		if !ok {
			fmt.Printf("[ERROR] Failed to obtain non-sudo user: %s\n", err)
			os.Exit(1)
		}

		sudoerUIDi, ok := os.LookupEnv("SUDO_UID")
		if !ok {
			fmt.Printf("[ERROR] Failed to obtain non-sudo UID: %s\n", err)
			os.Exit(1)
		}
		sudoerUID, err = strconv.Atoi(sudoerUIDi)
		if err != nil {
			fmt.Print("[ERROR] Unable to convert UID to INT\n")
			os.Exit(1)
		}

		if platform == "darwin" {
			homePath = "/Users/" + userName
		}
	}

	// Check the platform in use and adjust as appropriate
	if platform == "linux" || platform == "darwin" {
		rootCheck(uid)
	} else if platform == "windows" {
		userName = strings.Split(userName, "\\")[1]
	} else {
		fmt.Printf("[ERROR] Unsupported platform: %s\n", platform)
		os.Exit(1)
	}

	fmt.Printf("User used to execute this program: %s\n\n", userName)

	fmt.Print("[NOTICE] A few files (index, data_0-3) are in use by Discord while it is running.\n")
	fmt.Print("[...] To also dump those files, kill all instances of Discord before continuing.\n")
	fmt.Print("[...] There will be a notice regarding this per installed Discord build\n")
	fmt.Print("[...] where applicable while copying the files over.\n\n")

	fmt.Print("Press enter to continue ...\n")
	fmt.Scanln()

	// Discord client build names
	discordBuildName := map[int]string{
		0: "Normal",
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
	//cachedFileExt := make(map[int]map[int]string)

	fmt.Print("Checking for existing cache directories ...\n\n")

	// Check if directories exist
	for i := 0; i < len(discordBuildDir); i++ {
		filePath := fmt.Sprintf(cachePath[platform], homePath, discordBuildDir[i])
		fmt.Printf("DEBUG: %s\n", filePath)
		//fmt.Printf(cachePath[platform]+"\n", homePath, discordBuildDir[i])

		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			fmt.Printf("Found: Discord %s\n", discordBuildName[i])
			pathStatus[i] = true
		} else {
			pathStatus[i] = false
		}
	}

	fmt.Print("\n")

	// Check if the directories are empty and store names of cached files for later on
	for i := 0; i < len(discordBuildDir); i++ {
		if pathStatus[i] {
			filePath := fmt.Sprintf(cachePath[platform], homePath, discordBuildDir[i])
			cacheListing, err := ioutil.ReadDir(filePath)
			if err != nil {
				fmt.Printf("[ERROR] Unable to read directory for Discord %s\n", discordBuildName[i])
				os.Exit(1)
			}
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

	fmt.Print("\n")

	// If there are still no results at all, there is no point in continuing
	pathStatusSuccessCount := 0
	for i := 0; i < len(discordBuildDir); i++ {
		if pathStatus[i] {
			pathStatusSuccessCount++
		}
	}
	if pathStatusSuccessCount == 0 {
		fmt.Print("[ERROR] No cache found\n")
		os.Exit(1)
	}

	// Check space requirements
	curDir, err := os.Getwd()
	if err != nil {
		fmt.Print("[ERROR] Unable to obtain current directory\n")
		os.Exit(1)
	}
	freeStorage(curDir)
	remainingStorage := spareStorage - overallSize
	if remainingStorage <= 0 {
		requiredSpace := strings.Replace(fmt.Sprintf("%v", remainingStorage), "-", "", -1) // Remove '-' for a look that makes more sense
		fmt.Print("[ERROR] Insufficient storage where program is being ran\n")
		fmt.Printf("[...] %s bytes need sparing\n", requiredSpace)
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

	// Copy files over
	for i := 0; i < len(discordBuildDir); i++ {
		if len(cachedFile[i]) > 0 {
			if _, err := os.Stat(dumpDir + "/" + timeDateStamp + "/" + discordBuildName[i] + "/"); os.IsNotExist(err) {
				os.Mkdir(dumpDir+"/"+timeDateStamp+"/"+discordBuildName[i], 0644)
				if platform != "windows" {
					os.Chown(dumpDir+"/"+timeDateStamp+"/"+discordBuildName[i], sudoerUID, sudoerUID)
				}
			}

			fmt.Printf("Copying %d files from Discord %s ...\n", len(cachedFile[i]), discordBuildName[i])
			for it := 0; it < len(cachedFile[i]); it++ {
				// copy func
				fromPath := fmt.Sprintf(cachePath[platform], homePath, discordBuildDir[i])
				toPath := dumpDir + "/" + timeDateStamp + "/" + discordBuildName[i] + "/" + cachedFile[i][it]

				copyFile(fromPath+cachedFile[i][it], toPath) // Copy files
			}

			// Unable to copy client-critial cache
			if unreadableResources > 0 {
				fmt.Printf("[NOTICE] Cannot read client-critial cache while Discord %s is running\n", discordBuildName[i])
				fmt.Printf("[...] Unable to read %d client-critial cache file(s)\n", unreadableResources)
				fmt.Printf("[...] Actually copied %d cache files from Discord %s\n", len(cachedFile[i])-unreadableResources, discordBuildName[i])
			}
		}
	}

	// Analyse and rename files...
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
						os.Rename(cachedFilePath, cachedFilePath+"."+kind.Extension)
						identificationCount++
					}
				}
			}
			fmt.Printf("%d out of %d has been identified for Discord %s\n\n", identificationCount, len(cachedFile[i]), discordBuildName[i])
		}
	}

	// Let user know where to get to the files
	if platform == "windows" {
		fmt.Printf("Saved: %s\\%s\\%s\n", curDir, dumpDir, timeDateStamp)
	} else {
		fmt.Printf("Saved: %s/%s/%s\n", curDir, dumpDir, timeDateStamp)
	}

}
