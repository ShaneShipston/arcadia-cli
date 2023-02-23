package main

import (
    "archive/zip"
    "bufio"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "math/rand"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"
)

var archive = ""
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
        fmt.Println("This directory doesn't contain an Arcadia theme")
        os.Exit(0)
    }
}

func checkInstalled() {
    contentFile := filepath.Join("acf-json", "group_572229fc5045c.json")
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
            fmt.Println(manifest["name"].(string) + " has already been installed")
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

    archive, err := zip.OpenReader(fileName)

    check(err)

    defer archive.Close()

    for _, f := range archive.File {
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

    archive, err := zip.OpenReader(fileName)

    check(err)

    defer archive.Close()

    for _, f := range archive.File {
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

func blankLayout() map[string]interface{} {
    str := `{
        "key": "",
        "name": "",
        "label": "",
        "display": "block",
        "sub_fields": [
            {
                "key": "field_59eea71bac7ed",
                "label": "Content",
                "name": "ctn",
                "aria-label": "",
                "type": "clone",
                "instructions": "",
                "required": 0,
                "conditional_logic": 0,
                "wrapper": {
                    "width": "",
                    "class": "",
                    "id": ""
                },
                "clone": [],
                "display": "seamless",
                "layout": "block",
                "prefix_label": 0,
                "prefix_name": 0
            }
        ],
        "min": "",
        "max": ""
    }`

    var newLayout map[string]interface{}

    json.Unmarshal([]byte(str), &newLayout)

    return newLayout
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
        fmt.Println("There was an issue updating " + filePath)
    }
}

func injectLayout() {
    contentFile := filepath.Join("acf-json", "group_572229fc5045c.json")
    content, err := os.ReadFile(contentFile)
    now := time.Now()

    check(err)

    var data map[string]interface{}

    err = json.Unmarshal(content, &data)

    check(err)

    layoutKey := randomString(14)
    newLayout := blankLayout()

    newLayout["key"] = layoutKey
    newLayout["name"] = manifest["key"]
    newLayout["label"] = manifest["name"]

    subFields := newLayout["sub_fields"].([]interface{})
    cloneField := subFields[0].(map[string]interface{})

    cloneField["key"] = "field_" + randomString(8)
    cloneField["clone"] = [1]string{manifest["acfgroup"].(string)}

    data["modified"] = now.Unix()

    fields := data["fields"].([]interface{})
    firstField := fields[0].(map[string]interface{})
    layouts := firstField["layouts"].(map[string]interface{})

    layouts[layoutKey] = newLayout

    file, _ := json.MarshalIndent(data, "", "    ")

    err = os.WriteFile(contentFile, file, 0644)

    check(err)
}

func modifyTheme() {
    appends, exists := manifest["modifications"].(map[string]interface{})

    if !exists {
        return
    }

    for file, code := range appends {
        fileContents, err := openFile(file)

        if err != nil {
            fmt.Println("There was an issue updating " + file)
            continue
        }

        if _, ok := code.(string); ok {
            fileContents = append(fileContents, code.(string))
            writeFile(file, fileContents)
            continue
        }

        lines := code.([]interface{})

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
                if (currentLine == target) {
                    found = true

                    switch mode {
                    case "append":
                        fileContents = append(fileContents[:index+2], fileContents[index+1:]...)
                        fileContents[index+1] = config["code"].(string)
                    case "prepend":
                        fileContents = append(fileContents[:index+1], fileContents[index:]...)
                        fileContents[index] = config["code"].(string)
                    case "replace":
                        fileContents[index] = config["code"].(string)
                    case "remove":
                        fileContents = append(fileContents[:index], fileContents[index+1:]...)
                    }

                    continue OUTER
                }
            }

            if !found {
                fmt.Println("Modification target was missing for " + file)
            }
        }

        writeFile(file, fileContents)
    }
}

func installBlock() {
    injectLayout()
    modifyTheme()
}

func cleanUp() {
    os.Remove(archivePath())
    os.Remove("manifest.json")
}

func downloadBlock() {
    blockUrl := "https://arcadiadocs.com/packages/blocks/faqs/faq1/faq.zip"
    resp, err := http.Get(blockUrl)
    if err != nil {
        fmt.Printf("Block not found")
    }

    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        fmt.Printf("Block not found")
    }

    out, err := os.Create("faq.zip")
    check(err)

    defer out.Close()

    _, err = io.Copy(out, resp.Body)
    check(err)
}

func install() {
    archive = "faq"

    // Step 1. Check if Arcadia theme
    checkDirectory()

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
    installBlock()

    // Step 7. Clean up
    cleanUp()

    fmt.Println(manifest["name"].(string) + " has been installed")
}

func version() {
    fmt.Println("0.1.0-beta")
}

func main() {
    command := "install"

    switch command {
    case "install":
        install()
    case "version":
        version()
    }
}
