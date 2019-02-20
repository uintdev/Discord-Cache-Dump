# Discord Cache Dump
<img src="https://raw.githubusercontent.com/NodePoint/Discord-Cache-Dump/master/screenshot.png" style="width:50%;" alt="Demo">

## About

Discord Cache Dump is a tool that gathers the cache of all known Electron Discord client build types, copies it into its own directory, and gives them their appropriate file extensions.

## Features

- Detection of known Discord build types
- Discloses count of amount of files it is unable to gather at the time for that particular build
- Supports Windows, GNU/Linux, and macOS
- Checks storage available where the program is being ran before copying
- Dumps are timestamped along with the cache being in their own build type directories

## Known limitations

- The files that the Discord client process is utilising at the time cannot be copied over as it involves opening
  - It is advised to just kill the client you wish to copy files from that contains strings of *potentially sensitive* data
- macOS requires root due to how permissions are set
  - The tool does get the sudoer user as well as UID for permission changing purposes, so no need to worry about that

## Always opened files
The following files are known to be constantly used by Discord and so cannot be copied while that Discord client is running.

| File   | Contents                                                                                         |
| ------ | ------------------------------------------------------------------------------------------------ |
| index  | Unknown                                                                                          |
| data_0 | Unknown                                                                                          |
| data_1 | Full URLs to friendly URLs, API, avatars, emojis, embeds, attachments, uploads (self and others) |
| data_2 | Code, assets (png, svg)                                                                          |
| data_3 | Certificates, hostnames, IP addresses, image EXIF, reference to javascript assets (webpack)      |

## Prerequisites

In order to compile the tool, there are a few things required to get it set up.

- Go (compiling)
- [h2non/filetype](https://github.com/h2non/filetype) (recognition of file types and extensions): `go get github.com/h2non/filetype`
- [ricochet2200/go-disk-usage](https://github.com/ricochet2200/go-disk-usage) (disk information): `go get github.com/ricochet2200/go-disk-usage/du`

## Usage

| Platform  | Command             |
| --------- | ------------------- |
| Windows   | `dcd_windows.exe`   |
| GNU/Linux | `./dcd_linux`       |
| macOS     | `sudo ./dcd_darwin` |

## Credits
| User                                        | Contribution                                        |
| ------------------------------------------- | --------------------------------------------------- |
| [NodePoint](https://github.com/NodePoint)   | Development, Windows and GNU/Linux platform testing |
| [NotZoeyDev](https://github.com/NotZoeyDev) | macOS platform testing                              |

## Tested
- Windows 10 Pro
- Kali Linux (GNU/Linux)
- macOS
