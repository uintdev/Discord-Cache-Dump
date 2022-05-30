package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"

	DCDUtils "discordcachedump/utils"

	"github.com/h2non/filetype"
	"github.com/jessevdk/go-flags"
)

const (
	softVersion = "1.2.3"
	dumpDir     = "dump"
)

var platform = runtime.GOOS

// Initialisation of file stats variables
var unreadableRes int64
var overallSize int64
var spareStorage int64

// Initialisation of system information variables
var uid int
var userName string
var homePath string
var sudoerUID int
var unreadableResCount int64

func main() {

	fmt.Println()

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

	// Initialise flags
	var discordBuildListSlice []string
	var discordBuildListOption string
	var discordBuildListEntryName string
	var discordBuildListEntryDir string

	for i := 0; i < len(discordBuildName); i++ {
		discordBuildListSlice = append(discordBuildListSlice, strings.ToLower(discordBuildName[i]))
	}

	var opts struct {
		Build          string `short:"b" long:"build" description:"Select build type: stable, ptb, canary, development"`
		Noninteractive bool   `short:"n" long:"noninteractive" description:"Non-interactive -- no 'enter' key required"`
	}

	_, err := flags.Parse(&opts)

	if err != nil {
		os.Exit(0)
	}

	// Banner
	fmt.Print("#######################################\n")
	fmt.Printf("# Discord Cache Dump :: Version %s #\n", softVersion)
	fmt.Print("#######################################\n\n")

	user, err := user.Current()
	if err != nil {
		fmt.Printf("[ERROR] Failed to obtain user: %s%s", err, DCDUtils.ExitNewLine())
		os.Exit(6)
	} else {
		if platform != "windows" {
			uidConv, err := strconv.Atoi(user.Uid)
			if err != nil {
				fmt.Printf("[ERROR] Unable to obtain UID%s", DCDUtils.ExitNewLine())
				os.Exit(5)
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
				fmt.Printf("[ERROR] Unable to convert UID to INT%s", DCDUtils.ExitNewLine())
				os.Exit(4)
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
		DCDUtils.RootCheck(uid)
	} else if platform == "windows" {
		// Windows does not return just the username, so we split and grab what we need
		userName = strings.Split(userName, "\\")[1]
	} else {
		fmt.Printf("[ERROR] Unsupported platform: %s%s", platform, DCDUtils.ExitNewLine())
		os.Exit(3)
	}

	fmt.Printf("Logged in as: %s\n\n", userName)
	// Build flag
	if opts.Build != "" {
		discordBuildListOption = strings.ToLower(opts.Build)
		if !DCDUtils.SearchSlice(discordBuildListSlice, discordBuildListOption) {
			fmt.Printf("[ERROR] Build type does not exist %s", DCDUtils.ExitNewLine())
			os.Exit(1)
		} else {
			fmt.Printf("Build selected: %s\n\n", discordBuildListOption)
			for i := 0; i < len(discordBuildDir); i++ {
				if discordBuildListOption == strings.ToLower(discordBuildName[i]) {
					discordBuildListEntryName = discordBuildName[i]
					discordBuildListEntryDir = discordBuildDir[i]
					break
				}
			}
			discordBuildName = map[int]string{
				0: discordBuildListEntryName,
			}
			discordBuildDir = map[int]string{
				0: discordBuildListEntryDir,
			}
		}
	}

	fmt.Print("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
	fmt.Print("!! A few files (index, data_0-3) are in use by Discord while    !!\n")
	fmt.Print("!! it is running. To also dump those files, kill all instances  !!\n")
	fmt.Print("!! of Discord before continuing. There will be a notice         !!\n")
	fmt.Print("!! regarding this per installed Discord build where applicable  !!\n")
	fmt.Print("!! while copying the files over.                                !!\n")
	fmt.Print("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n\n")

	if !opts.Noninteractive {
		fmt.Print("Press enter to continue ...\n")
		fmt.Scanln()
	}

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
				fmt.Printf("[ERROR] Unable to read directory for Discord %s%s", discordBuildName[i], DCDUtils.ExitNewLine())
				os.Exit(2)
			}
			// Go through the list of files present in a cache directory
			if len(cacheListing) > 0 {
				cachedFile[i] = make(map[int]string)
				for k, v := range cacheListing {
					cachedFile[i][k] = v.Name()
					overallSize = DCDUtils.SizeStore(filePath + v.Name(), overallSize)
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
		fmt.Printf("No cache found -- ensure Discord is installed and had been ran at least once%s", DCDUtils.ExitNewLine())
		os.Exit(0)
	} else {
		fmt.Print("\n")
	}

	// Check storage requirements
	curDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("[ERROR] Unable to obtain current directory%s", DCDUtils.ExitNewLine())
		os.Exit(1)
	}
	spareStorage = DCDUtils.FreeStorage(curDir)
	if spareStorage <= overallSize {
		remainingStorage := spareStorage - overallSize
		requiredSpace := strings.Replace(fmt.Sprintf("%v", remainingStorage), "-", "", -1)
		fmt.Print("[ERROR] Insufficient storage where program is being ran\n")
		fmt.Printf("[...] %s bytes need sparing%s", requiredSpace, DCDUtils.ExitNewLine())
		os.Exit(1)
	} else {
		fmt.Print("Sufficient storage -- safe to go ahead\n\n")
	}

	// Check and create dump directory structure
	timeDateStamp := DCDUtils.TimeDate()
	if _, err := os.Stat(dumpDir + "/"); os.IsNotExist(err) {
		os.Mkdir(dumpDir, 0755)
		if platform != "windows" {
			os.Chown(dumpDir, sudoerUID, sudoerUID)
		}
	}
	if _, err := os.Stat(dumpDir + "/" + timeDateStamp + "/"); os.IsNotExist(err) {
		os.Mkdir(dumpDir+"/"+timeDateStamp, 0755)
		if platform != "windows" {
			os.Chown(dumpDir+"/"+timeDateStamp, sudoerUID, sudoerUID)
		}
	}

	// Create any Discord client directories that are required
	for i := 0; i < len(discordBuildDir); i++ {
		if len(cachedFile[i]) > 0 {
			if _, err := os.Stat(dumpDir + "/" + timeDateStamp + "/" + discordBuildName[i] + "/"); os.IsNotExist(err) {
				os.Mkdir(dumpDir+"/"+timeDateStamp+"/"+discordBuildName[i], 0755)
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
				unreadableRes = DCDUtils.CopyFile(fromPath+cachedFile[i][it], toPath, sudoerUID, unreadableRes)
			}

			// Unable to copy client-critical cache
			if unreadableRes > 0 {
				fmt.Printf("[NOTICE] Cannot read client-critical cache while Discord %s is running\n", discordBuildName[i])
				fmt.Printf("[...] Unable to read %d client-critical cache file(s)\n", unreadableRes)
				unreadableResCount = int64(len(cachedFile[i])) - unreadableRes
				if unreadableResCount < 0 {
					unreadableResCount = 0
				}
				fmt.Printf("[...] Actually copied %d cache files from Discord %s\n", unreadableResCount, discordBuildName[i])
			} else {
				unreadableResCount = int64(len(cachedFile[i]))
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
			fmt.Printf("%d out of %d identified for Discord %s\n\n", identificationCount, unreadableResCount, discordBuildName[i])
		}
	}

	// Let user know where to get to the files
	if platform == "windows" {
		fmt.Printf("Saved: %s\\%s\\%s%s", curDir, dumpDir, timeDateStamp, DCDUtils.ExitNewLine())
	} else {
		fmt.Printf("Saved: %s/%s/%s%s", curDir, dumpDir, timeDateStamp, DCDUtils.ExitNewLine())
	}
}
