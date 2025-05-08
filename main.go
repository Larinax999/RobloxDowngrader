package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
	"golang.org/x/sys/windows/registry"
)

var files = map[string]string{
	"RobloxApp.zip": `\`,
	"redist.zip":    `\`,

	"shaders.zip": `shaders\`,
	"ssl.zip":     `ssl\`,

	"WebView2.zip":                 `\`,
	"WebView2RuntimeInstaller.zip": `WebView2RuntimeInstaller\`,

	"content-avatar.zip":    `content\avatar\`,
	"content-configs.zip":   `content\configs\`,
	"content-fonts.zip":     `content\fonts\`,
	"content-sky.zip":       `content\sky\`,
	"content-sounds.zip":    `content\sounds\`,
	"content-textures2.zip": `content\textures\`,
	"content-models.zip":    `content\models\`,

	"content-textures3.zip":      `PlatformContent\pc\textures\`,
	"content-terrain.zip":        `PlatformContent\pc\terrain\`,
	"content-platform-fonts.zip": `PlatformContent\pc\fonts\`,

	"extracontent-luapackages.zip":  `ExtraContent\LuaPackages\`,
	"extracontent-translations.zip": `ExtraContent\translations\`,
	"extracontent-models.zip":       `ExtraContent\models\`,
	"extracontent-textures.zip":     `ExtraContent\textures\`,
	"extracontent-places.zip":       `ExtraContent\places\`,
}

func GetRobloxPath() (string, error) {
	key, err := registry.OpenKey(registry.CLASSES_ROOT, `roblox-player\shell\open\command`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer key.Close()

	path, _, err := key.GetStringValue("")
	if err != nil {
		return "", err
	}
	paths := strings.Split(path, `\`)
	return strings.Join(paths[:len(paths)-1], `\`)[1:] + `\`, nil
}

func main() { // go build -ldflags "-s -w"
	fmt.Println("[<->] Roblox DownGrader Made By Larina")
	defer time.Sleep(time.Second * 5)

	path, err := GetRobloxPath()
	if err != nil {
		fmt.Println("[RBX] Roblox Not Found", err.Error())
		return
	}

	var Oversion string
	fmt.Println("[RBX] Roblox Location >", path)
	fmt.Print("[RBX] What? | version-")
	fmt.Scanln(&Oversion)

	// download rbxPkgManifest.txt
	r := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(r)
	defer fasthttp.ReleaseResponse(resp)

	r.SetRequestURI(fmt.Sprintf("https://setup.rbxcdn.com/version-%s-rbxPkgManifest.txt", Oversion))
	if err = fasthttp.Do(r, resp); err != nil {
		fmt.Println("[RBX] Fail to fetch rbxPkgManifest.txt", err.Error())
		return
	}

	if resp.StatusCode() != 200 {
		fmt.Println("[RBX] rbxPkgManifest.txt non 200 >", resp.StatusCode())
		return
	}

	// fmt.Println("[RBX] Removeing new roblox")
	// os.RemoveAll(path)
	// for n, v := range files {
	// 	os.MkdirAll(path+v, 0777)
	// 	fmt.Println("[RBX] Create folder for", n)
	// }

	var wg sync.WaitGroup
	for i, v := range strings.Split(string(resp.Body()), "\r\n")[1:] {
		if i%4 == 0 {
			go func(filename string) {
				defer wg.Done()
				subpath, ok := files[filename]
				if !ok {
					// fmt.Println("[RBX] No fmap for", filename)
					return
				}

				fmt.Printf("[RBX] downloading %s\n", filename)
				req := fasthttp.AcquireRequest()
				resp := fasthttp.AcquireResponse()
				defer fasthttp.ReleaseRequest(req)
				defer fasthttp.ReleaseResponse(resp)
				req.SetRequestURI(fmt.Sprintf("https://setup.rbxcdn.com/version-%s-%s", Oversion, filename))

				if err := fasthttp.Do(req, resp); err != nil {
					fmt.Println("[RBX] Fail to fetch", filename, err.Error())
					return
				}

				body, err := resp.BodyUncompressed()
				if err != nil {
					fmt.Println("[RBX] Fail to read body")
					return
				}

				zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
				if err != nil {
					fmt.Println("[RBX] Fail to unzip", filename, err.Error())
					return
				}

				for _, zf := range zipReader.File {
					if zf.FileInfo().IsDir() {
						// it \\ but it works :>
						os.MkdirAll(path+subpath+zf.Name[1:], 0777)
						continue
					}

					f, err := zf.Open()
					if err != nil {
						fmt.Println("[RBX] Fail to unzip", zf.Name, err.Error())
						continue
					}

					fmt.Println("[RBX] writing", zf.Name)
					ff, err := os.Create(path + subpath + zf.Name)
					if err != nil {
						f.Close()
						fmt.Println("[RBX] Fail to create", zf.Name, err.Error())
						continue
					}
					b, err := io.ReadAll(f)
					if err != nil {
						f.Close()
						ff.Close()
						fmt.Println("[RBX] Fail to read", zf.Name, err.Error())
						continue
					}
					ff.Write(b)

					f.Close()
					ff.Close()
				}
			}(v)
			wg.Add(1)
		}
	}
	wg.Wait()
	fmt.Println("[RBX] All Done.")
}
