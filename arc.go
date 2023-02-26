package main

import (
    "archive/zip"
    "bufio"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/pterm/pterm"
    "io"
    "math/rand"
    "net/http"
    "os"
    "path/filepath"
    "runtime"
    "strconv"
    "strings"
    "time"
)

var appVersion = "0.1.0"
var archive = ""
var updateAvailable = ""
var manifest map[string]interface{}

func init() {
    rand.Seed(time.Now().UnixNano())
}

func randomString(length int) string {
    str := make([]rune, length)
    letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

    for i := range str {
        str[i] = letters[rand.Intn(len(letters))]
    }

    return string(str)
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func checkDirectory() {
    if _, err := os.Stat(filepath.Join("src", "scss", "style.scss")); errors.Is(err, os.ErrNotExist) {
        pterm.Error.Println("This directory doesn't contain an Arcadia theme")
        os.Exit(0)
    }
}

func groupName() string {
    if manifest["contents"] == "component" {
        return "group_5c903f684a8ae.json"
    }

    return "group_572229fc5045c.json"
}

func finalLayout() string {
    if manifest["contents"] == "component" {
        return "Widget"
    }

    return "Page Content (Layouts Only)"
}

func checkInstalled() {
    contentFile := filepath.Join("acf-json", groupName())
    content, err := os.ReadFile(contentFile)

    check(err)

    var data map[string]interface{}

    err = json.Unmarshal(content, &data)

    check(err)

    fields := data["fields"].([]interface{})
    firstField := fields[0].(map[string]interface{})
    layouts := firstField["layouts"].(map[string]interface{})

    for _, element := range layouts {
        layout := element.(map[string]interface{})

        if layout["name"] == manifest["key"] {
            pterm.Warning.Println(manifest["name"].(string) + " has already been installed")
            cleanUp()
            os.Exit(0)
        }
    }
}

func archivePath() string {
    return strings.Join([]string{archive, "zip"}, ".");
}

func extractManifest() {
    fileName := archivePath()

    archiveFile, err := zip.OpenReader(fileName)

    check(err)

    defer archiveFile.Close()

    for _, f := range archiveFile.File {
        filePath := f.Name

        if f.FileInfo().IsDir() || f.Name != "manifest.json" {
            continue
        }

        dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        check(err)

        fileInArchive, err := f.Open()
        check(err)

        if _, err := io.Copy(dstFile, fileInArchive); err != nil {
            panic(err)
        }

        dstFile.Close()
        fileInArchive.Close()
    }
}

func extractArchive() {
    fileName := archivePath()

    archiveFile, err := zip.OpenReader(fileName)

    check(err)

    defer archiveFile.Close()

    for _, f := range archiveFile.File {
        filePath := f.Name

        if f.Name == "manifest.json" {
            continue
        }

        if f.FileInfo().IsDir() {
            os.MkdirAll(filePath, os.ModePerm)
            continue
        }

        if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
            panic(err)
        }

        dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        check(err)

        fileInArchive, err := f.Open()
        check(err)

        if _, err := io.Copy(dstFile, fileInArchive); err != nil {
            panic(err)
        }

        dstFile.Close()
        fileInArchive.Close()
    }
}

func readManifest() {
    contentFile := filepath.Join("manifest.json")
    content, err := os.ReadFile(contentFile)

    check(err)

    err = json.Unmarshal(content, &manifest)

    check(err)
}

func blankLayout() []string {
    layoutKey := randomString(14)

    str := []string{
        "\"" + layoutKey + "\": {",
        "    \"key\": \"" + layoutKey + "\",",
        "    \"name\": \"" + manifest["key"].(string) + "\",",
        "    \"label\": \"" + manifest["name"].(string) + "\",",
        "    \"display\": \"block\",",
        "    \"sub_fields\": [",
        "        {",
        "            \"key\": \"field_" + randomString(8) + "\",",
        "            \"label\": \"Content\",",
        "            \"name\": \"ctn\",",
        "            \"aria-label\": \"\",",
        "            \"type\": \"clone\",",
        "            \"instructions\": \"\",",
        "            \"required\": 0,",
        "            \"conditional_logic\": 0,",
        "            \"wrapper\": {",
        "                \"width\": \"\",",
        "                \"class\": \"\",",
        "                \"id\": \"\"",
        "            },",
        "            \"clone\": [",
        "                \"" + manifest["acfgroup"].(string) + "\"",
        "            ],",
        "            \"display\": \"seamless\",",
        "            \"layout\": \"block\",",
        "            \"prefix_label\": 0,",
        "            \"prefix_name\": 0",
        "        }",
        "    ],",
        "    \"min\": \"\",",
        "    \"max\": \"\"",
        "},",
    }

    return str
}

func openFile(filePath string) ([]string, error) {
    file, err := os.Open(filePath)

    if err != nil {
        return nil, err
    }

    defer file.Close()

    var lines []string

    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return lines, nil
}

func writeFile(filePath string, fileContents []string) {
    newContent := ""

    for _, line := range fileContents {
        newContent += line
        newContent += "\n"
    }

    err := os.WriteFile(filePath, []byte(newContent), 0644)

    if err != nil {
        pterm.Warning.Println("There was an issue updating " + filePath)
    }
}

func injectLayout() {
    contentFile := filepath.Join("acf-json", groupName())
    content, err := openFile(contentFile)
    now := time.Now()
    currentTime := strconv.FormatInt(now.Unix(), 10)

    check(err)

    newLayout := blankLayout()
    finalLayout := finalLayout()

    startIndex := 0
    bracketCount := 0
    distanceFromBracket := 0
    matched := false
    prefix := strings.Repeat(" ", 16)

    for index, line := range content {
        if startIndex > 0 && !matched {
            if strings.HasSuffix(line, "{") {
                bracketCount = bracketCount + 1

                if bracketCount == 1 {
                    distanceFromBracket = 0
                }
            } else if bracketCount >= 1 {
                distanceFromBracket = distanceFromBracket + 1
            }

            if strings.Contains(line, "}") {
                bracketCount = bracketCount - 1
            }

            if bracketCount == 1 && strings.Contains(line, "\"label\"") {
                layoutName := line[strings.Index(line, ":") + 3:len(line) - 2]

                if strings.Compare(manifest["name"].(string), layoutName) < 0 || layoutName == finalLayout {
                    matched = true
                    start := index - distanceFromBracket

                    for distance, body := range newLayout {
                        content = append(content[:start+distance+1], content[start+distance:]...)
                        content[start+distance] = prefix + body
                    }
                }
            }
        } else if strings.Contains(line, "\"layouts\"") {
            startIndex = index;
        }

        if strings.Contains(line, "\"modified\"") {
            if strings.Contains(line, ",") {
                content[index] = "    \"modified\": " + currentTime + ","
            } else {
                content[index] = "    \"modified\": " + currentTime
            }
        }
    }

    writeFile(contentFile, content)
}

func modifyTheme() {
    appends, exists := manifest["modifications"].(map[string]interface{})

    if !exists {
        return
    }

    for file, code := range appends {
        fileContents, err := openFile(file)

        if err != nil {
            pterm.Warning.Println("File not available for modification: " + file)
            continue
        }

        if _, ok := code.(string); ok {
            fileContents = append(fileContents, code.(string))
            writeFile(file, fileContents)
            continue
        }

        lines := code.([]interface{})

        modificationMade := false

        OUTER:
        for _, line := range lines {
            if _, ok := line.(string); ok {
                fileContents = append(fileContents, line.(string))
                continue
            }

            config := line.(map[string]interface{})

            target, exists := config["target"].(string)

            if !exists {
                fileContents = append(fileContents, config["code"].(string))
                continue
            }

            mode, exists := config["mode"].(string)

            if !exists {
                mode = "append"
            }

            found := false

            for index, currentLine := range fileContents {
                // Code swap already happened
                if mode == "replace" && strings.Contains(currentLine, config["code"].(string)) {
                    found = true
                    continue OUTER
                } else if strings.Contains(currentLine, target) {
                    found = true
                    modificationMade = true

                    switch mode {
                    case "append":
                        fileContents = append(fileContents[:index+2], fileContents[index+1:]...)
                        fileContents[index+1] = config["code"].(string)
                    case "prepend":
                        fileContents = append(fileContents[:index+1], fileContents[index:]...)
                        fileContents[index] = config["code"].(string)
                    case "replace":
                        fileContents[index] = strings.Replace(fileContents[index], target, config["code"].(string), 1)
                    case "remove":
                        fileContents[index] = strings.Replace(fileContents[index], target, "", 1)
                    }

                    continue OUTER
                }
            }

            if !found && mode != "remove" && mode != "replace" {
                pterm.Warning.Println("Code could not be added to " + file)
            }

            if !found && mode == "replace" {
                pterm.Info.Println("A code replacement couldn't be made within " + file)
            }
        }

        if modificationMade {
            writeFile(file, fileContents)
        }
    }
}

func cleanUp() {
    os.Remove(archivePath())
    os.Remove("manifest.json")
}

func downloadBlock() {
    blockUrl := "https://arcadiadocs.com/download.php?target=" + archive
    resp, err := http.Get(blockUrl)
    if err != nil {
        pterm.Error.Println("Block not found")
        os.Exit(0)
    }

    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        pterm.Error.Println("Block not found")
        os.Exit(0)
    }

    out, err := os.Create(archivePath())
    check(err)

    defer out.Close()

    _, err = io.Copy(out, resp.Body)
    check(err)
}

func catalogArchives() []string {
    var archives []string

    directory, err := os.Open(".")
    check(err)

    files, err := directory.Readdir(0)
    check(err)

    for _, file := range files {
        if file.IsDir() {
            continue
        }

        if filepath.Ext(file.Name()) == ".zip" {
            valid := false
            archiveFile, err := zip.OpenReader(file.Name())

            check(err)

            defer archiveFile.Close()

            for _, f := range archiveFile.File {
                if f.Name == "manifest.json" {
                    valid = true
                    break
                }
            }

            if valid {
                archives = append(archives, file.Name())
            }
        }
    }

    return archives
}

func install() {
    // archives := make([]string, 3)
    // archives[0] = "faq"
    // archives[1] = "nav-bar"
    // archives[2] = "image-slider"
    if len(os.Args) <= 2 {
        pterm.Error.Println("Please specify a block")
        return
    }

    archives := os.Args[2:]

    // Step 1. Check if Arcadia theme
    checkDirectory()

    for _, file := range archives {
        archive = file

        // Step 2. Download Block
        downloadBlock()

        // Step 3. Extract & Read Manifest
        extractManifest()
        readManifest()

        // Step 4. Check if installed
        checkInstalled()

        // Step 5. Extract Archive
        extractArchive()

        // Step 6. Install
        injectLayout()
        modifyTheme()

        // Step 7. Clean up
        cleanUp()

        pterm.Success.Println(manifest["name"].(string) + " has been installed")
    }
}

func unpack() {
    // Step 1. Check if Arcadia theme
    checkDirectory()

    // Step 2: Catalog zips in the project root
    archives := catalogArchives()

    if len(archives) == 0 {
        pterm.Info.Println("No blocks found")
    }

    for _, file := range archives {
        archive = file[:len(file) - 4]

        // Step 3: Extract manifests one by one
        extractManifest()
        readManifest()

        // Step 4. Check if installed
        checkInstalled()

        // Step 5. Extract Archive
        extractArchive()

        // Step 6. Install
        injectLayout()
        modifyTheme()

        // Step 7. Clean up
        cleanUp()

        pterm.Success.Println(manifest["name"].(string) + " has been installed")
    }
}

func checkVersion() {
    checkUrl := "https://arcadiadocs.com/check.php?version=" + appVersion + "&os=" + runtime.GOOS + "&arch=" + runtime.GOARCH
    resp, err := http.Get(checkUrl)

    if err != nil {
        return
    }

    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return
    }

    body, err := io.ReadAll(resp.Body)
    updateAvailable = string(body)
}

func version() {
    pterm.DefaultBasicText.Println("You are currently running version: " + pterm.LightMagenta(appVersion))
}

func update() {
    if len(updateAvailable) == 0 {
        pterm.DefaultBasicText.Println("You are on the latest version")
        return
    }

    ext := ""

    if runtime.GOOS == "windows" {
        ext = ".exe"
    }

    pterm.DefaultBasicText.Println("Updating...")

    appPath, err := os.Executable()

    check(err)

    appDir := filepath.Dir(appPath)

    resp, err := http.Get(updateAvailable)

    if err != nil {
        pterm.Error.Println("Update failed")
        return
    }

    defer resp.Body.Close()

    tmpName := filepath.Join(appDir, "arc-new" + ext);
    currentName := filepath.Join(appDir, "arc" + ext);
    oldName := filepath.Join(appDir, "arc-old" + ext);

    out, err := os.Create(tmpName)
    check(err)

    _, err = io.Copy(out, resp.Body)
    check(err)

    out.Close()

    err = os.Rename(currentName, oldName)
    check(err)
    err = os.Rename(tmpName, currentName)
    check(err)

    pterm.Success.Println("Update complete")
}

func updateCleanup() {
    ext := ""

    if runtime.GOOS == "windows" {
        ext = ".exe"
    }

    appPath, err := os.Executable()

    if err != nil {
        return
    }

    appDir := filepath.Dir(appPath)

    os.Remove(filepath.Join(appDir, "arc-old" + ext))
}

func main() {
    command := "help"

    if len(os.Args) > 1 {
        command = os.Args[1]
    }

    checkVersion()
    updateCleanup()

    if command != "update" && len(updateAvailable) > 0 {
        pterm.Info.Println("Update Available\nRun " + pterm.LightMagenta("arc update") + " to install")
        fmt.Println()
    }

    switch command {
    case "install":
        install()
    case "unpack":
        unpack();
    case "version":
        version()
    case "update":
        update()
    case "help":
        pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
            {"Command", "Description"},
            {"install", "Install 1 or more blocks"},
            {"unpack", "Extract already downloaded blocks"},
            {"version", "Display the current app version"},
            {"update", "Perform an update on arc"},
        }).Render()
    default:
        pterm.Error.Println("Invalid command")
    }
}
