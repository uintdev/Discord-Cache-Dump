package DCDUtils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Formulate the end of the flatpak package name to then use as a normal path
func FlatpakPath(buildType string) string {
	resultingBuildSegment := ""
	splitBuildName := strings.Split(buildType, "discord")
	if len(splitBuildName) > 1 {
		if splitBuildName[1] != "" {
			resultingBuildSegment = strings.Join([]string{"discord", splitBuildName[1]}, " ")
		} else {
			resultingBuildSegment = buildType
		}
	} else {
		resultingBuildSegment = buildType
	}

	titleCaser := cases.Title(language.English)
	casedBuildName := titleCaser.String(resultingBuildSegment)

	packageNameSuffix := strings.Replace(casedBuildName, " ", "", -1)

	path := "%s/.var/app/com.discordapp.%s/config/%s/Cache/Cache_Data/"
	path = fmt.Sprintf(path, "%s", packageNameSuffix, "%s")

	return path
}


// Extract cache files (for GNU/Linux and macOS)
func FileExtractor(contents []byte) []byte {
	var magicNumber string
	formatLock := false
	sanitiseLock := false
	re := regexp.MustCompile(`^[a-zA-Z0-9_\-:/.%?&=]*$`)

	var extractDat []byte
	extractDat = contents

	// Clean ending
	magicNumber = "\xd8\x41\x0d\x97\x45\x6f\xfa\xf4\x01\x00"
	extractRequired := bytes.Contains(extractDat, []byte(magicNumber))
	if extractRequired {
		// Remove end bytes
		extractDat = bytes.SplitN(extractDat, []byte(magicNumber), 2)[0]
	}

	// There are a few file types that are missing data we would need
	// in order to extract them the way we intend to do so.
	// In that case, we rely on their magic numbers.

	// Clean JPG/JPEG
	if !formatLock {
		magicNumber = "\xff\xd8\xff"
		magicNumberIndex := bytes.Index(contents, []byte(magicNumber))
		extractRequired = false
		if magicNumberIndex > -1 {
			if bytes.Contains(contents[magicNumberIndex:magicNumberIndex+3], []byte(magicNumber)) {
				if len(contents[magicNumberIndex+6:magicNumberIndex+10]) == 4 {
					if bytes.Equal(contents[magicNumberIndex+6:magicNumberIndex+10], []byte("\x4a\x46\x49\x46")) {
						extractRequired = true
					}
				}
			}
		}
		if extractRequired {
			formatLock = true
			extractDat = bytes.SplitN(extractDat, []byte(magicNumber), 2)[1]
			extractDat = append([]byte(magicNumber), extractDat...)
		}
	}

	// Clean WEBP
	if !formatLock {
		magicNumber = "\x00\x00\x57\x45\x42\x50\x56\x50\x38"
		extractRequired = bytes.Contains(contents, []byte(magicNumber))
		if !extractRequired {
			magicNumber = "\x01\x00\x57\x45\x42\x50\x56\x50\x38"
			extractRequired = bytes.Contains(contents, []byte(magicNumber))
		}
		if extractRequired {
			formatLock = true
			magicNumber = "\x52\x49\x46\x46"
			extractDat = bytes.SplitN(extractDat, []byte(magicNumber), 2)[1]
			extractDat = append([]byte(magicNumber), extractDat...)
		}
	}

	// Extract the rest
	if !formatLock && len(extractDat[12:13]) == 1 {
		uriLengthInt := extractDat[12:13]
		uriLengthConv := fmt.Sprintf("%d", uriLengthInt[0])
		uriLength, err := strconv.Atoi(uriLengthConv)
		if err == nil {
			if uriLength > 0 && len(extractDat[24:24+uriLength]) == uriLength && re.MatchString(string(extractDat[24 : 24+uriLength][0])) {
				fileContent := extractDat[24+uriLength:]
				extractDat = fileContent
			} else {
				extractDat = contents
				sanitiseLock = true
			}
		} else {
			extractDat = contents
			sanitiseLock = true
		}
	}

	// Clean extra data
	magicNumber = "\x6b\x67\x53\x65\x01\xbf\x97\xeb"
	extractRequired = bytes.Contains(extractDat, []byte(magicNumber))
	if !sanitiseLock && extractRequired {
		extractDatTmp := bytes.Split(extractDat, []byte(magicNumber))
		var extractDatBuilder []byte
		for i := 0; i < len(extractDatTmp); i++ {
			if i > 0 {
				extractDatBuilder = append(extractDatBuilder, extractDatTmp[i][24:]...)
			}
		}
		extractDat = extractDatBuilder
	}

	extractedConvByte := extractDat

	return extractedConvByte
}

// Copy source file to destination
func CopyFile(from string, to string, permuid int, unreadableCount int64) int64 {
	var platform string = runtime.GOOS
	var resData []byte
	input, err := ioutil.ReadFile(from)
	if err != nil {
		/*
			We simply assume that the file it got to cannot be 'opened' because Discord's process is using it at the time.
			Yes, as any other error could occur for whatever reason, it is not the best way to go but it's good enough.
		*/
		return unreadableCount + 1
	}

	if platform != "windows" {
		resData = FileExtractor(input)
	} else {
		resData = input
	}

	err = ioutil.WriteFile(to, resData, 0644)
	if err != nil {
		fmt.Printf("\n[ERROR] Write error: %s\n", err)
		os.Exit(7)
	}
	// If ran as root as a sudoer on GNU/Linux or macOS, it's going to be root:root, so we change it to what it really should be
	if platform != "windows" {
		os.Chown(to, permuid, permuid)
	}

	// Preserve file modified timestamp
	info, _ := os.Stat(from)
	_ = os.Chtimes(to, info.ModTime(), info.ModTime())

	return unreadableCount
}

// Slice search
func SearchSlice(y []string, z string) bool {
	for _, n := range y {
		if z == n {
			return true
		}
	}
	return false
}
