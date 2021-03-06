package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/google/go-containerregistry/pkg/crane"
)

type opData struct {
	FinalRepo string
}

type resultData struct {
	Package Package
	Exists  bool
}

func downloadMeta(p Package, o opData) error {
	tmpdir, err := ioutil.TempDir(os.TempDir(), "ci")
	if err != nil {
		return err
	}
	unpackdir, err := ioutil.TempDir(os.TempDir(), "ci")
	if err != nil {
		return err
	}
	err = RunSH("unpack", fmt.Sprintf("TMPDIR=%s XDG_RUNTIME_DIR=%s luet util unpack %s %s", tmpdir, tmpdir, p.ImageMetadata(o.FinalRepo), unpackdir))
	if err != nil {
		return err
	}
	err = RunSH("move", fmt.Sprintf("mv %s/* build/", unpackdir))
	if err != nil {
		return err
	}
	os.RemoveAll(tmpdir)
	os.RemoveAll(unpackdir)
	return nil
}

func metaWorker(i int, wg *sync.WaitGroup, c <-chan Package, o opData) error {
	defer wg.Done()

	for p := range c {
		checkErr(downloadMeta(p, o))
	}
	return nil
}

func getResultData(p Package, o opData) resultData {
	fmt.Println("Checking", p, p.Image(o.FinalRepo))
	return resultData{Package: p, Exists: p.ImageAvailable(o.FinalRepo)}
}

func buildWorker(i int, wg *sync.WaitGroup, c <-chan Package, o opData, results chan<- resultData) error {
	defer wg.Done()

	for p := range c {
		results <- getResultData(p, o)
	}
	return nil
}

func main() {
	finalRepo := os.Getenv("FINAL_REPO")
	if finalRepo == "" {
		fmt.Println("A container repository must be specified with FINAL_REPO")
		os.Exit(1)
	}
	buildScript := os.Getenv("BUILD_SCRIPT")
	if buildScript == "" {
		buildScript = "./.github/build.sh"
	}
	packs, err := TreePackages("./packages")
	checkErr(err)
	missingPackages := []Package{}
	op := opData{FinalRepo: finalRepo}
	if os.Getenv("PARALLEL") == "true" {
		all := make(chan Package)
		results := make(chan resultData, len(packs.Packages))
		wg := new(sync.WaitGroup)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go buildWorker(i, wg, all, op, results)
		}

		for _, p := range packs.Packages {
			all <- p
		}
		close(all)

		for s := 1; s <= len(packs.Packages); s++ {
			a := <-results
			if !a.Exists {
				missingPackages = append(missingPackages, a.Package)
			}
		}

		wg.Wait()
	} else {
		for _, p := range packs.Packages {
			a := getResultData(p, op)
			if !a.Exists {
				missingPackages = append(missingPackages, a.Package)
			}
		}
	}

	fmt.Println("Missing packages: " + fmt.Sprint(len(missingPackages)))
	for _, m := range missingPackages {
		fmt.Println("-", m.String())
	}

	if os.Getenv("DOWNLOAD_ONLY") != "true" {
		for _, m := range missingPackages {
			fmt.Println("Building", m.String())
			checkErr(RunSH("build", fmt.Sprintf("%s %s", buildScript, m.String())))
		}
	}

	if os.Getenv("DOWNLOAD_METADATA") == "true" {
		fmt.Println("Populating build folder with metadata files")
		op := opData{FinalRepo: finalRepo}
		if os.Getenv("PARALLEL") == "true" {
			all := make(chan Package)
			wg := new(sync.WaitGroup)

			for i := 0; i < 1; i++ {
				wg.Add(1)
				go metaWorker(i, wg, all, op)
			}

			for _, p := range packs.Packages {
				if !contains(missingPackages, p) {
					all <- p
				}
			}
			close(all)
			wg.Wait()
		} else {
			for _, p := range packs.Packages {
				if !contains(missingPackages, p) {
					err := downloadMeta(p, op)
					checkErr(err)
				}
			}
		}
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func RunSHOUT(stepName, bashFragment string) ([]byte, error) {
	cmd := exec.Command("sh", "-s")
	cmd.Stdin = strings.NewReader(bashWrap(bashFragment))

	cmd.Env = os.Environ()
	//	log.Printf("Running in background: %v", stepName)

	return cmd.CombinedOutput()
}

func RunSH(stepName, bashFragment string) error {
	cmd := exec.Command("sh", "-s")
	cmd.Stdin = strings.NewReader(bashWrap(bashFragment))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	log.Printf("Running: %v (%v)", stepName, bashFragment)

	return cmd.Run()
}

func bashWrap(cmd string) string {
	return `
set -o errexit
set -o nounset
` + cmd + `
`
}

type SearchResult struct {
	Packages []Package
}

type Package struct {
	Name, Category, Version, Path string
}

func TreePackages(treedir string) (searchResult SearchResult, err error) {
	var res []byte
	res, err = RunSHOUT("tree", fmt.Sprintf("luet tree pkglist --tree %s --output json", treedir))
	if err != nil {
		fmt.Println(string(res))
		return
	}
	json.Unmarshal(res, &searchResult)
	return
}

func imageAvailable(image string) bool {
	_, err := crane.Digest(image)
	return err == nil
}

func contains(pp []Package, p Package) bool {
	for _, i := range pp {
		if i.Equal(p) {
			return true
		}
	}
	return false
}

func (p Package) String() string {
	return fmt.Sprintf("%s/%s@%s", p.Category, p.Name, p.Version)
}

func (p Package) Image(repository string) string {
	// ${name}-${category}-${version//+/-}
	return fmt.Sprintf("%s:%s-%s-%s", repository, p.Name, p.Category, strings.ReplaceAll(p.Version, "+", "-"))
}

func (p Package) ImageMetadata(repository string) string {
	// ${name}-${category}-${version//+/-}
	return fmt.Sprintf("%s.metadata.yaml", p.Image(repository))
}

func (p Package) ImageAvailable(repository string) bool {
	return imageAvailable(p.Image(repository))
}

func (p Package) Equal(pp Package) bool {
	if p.Name == pp.Name && p.Category == pp.Category && p.Version == pp.Version {
		return true
	}
	return false
}

func (p Package) EqualS(s string) bool {
	if s == fmt.Sprintf("%s/%s", p.Category, p.Name) {
		return true
	}
	return false
}

func (p Package) EqualNoV(pp Package) bool {
	if p.Name == pp.Name && p.Category == pp.Category {
		return true
	}
	return false
}

func (s SearchResult) FilterByCategory(cat string) SearchResult {
	new := SearchResult{Packages: []Package{}}

	for _, r := range s.Packages {
		if r.Category == cat {
			new.Packages = append(new.Packages, r)
		}
	}
	return new
}

func (s SearchResult) FilterByName(name string) SearchResult {
	new := SearchResult{Packages: []Package{}}

	for _, r := range s.Packages {
		if !strings.Contains(r.Name, name) {
			new.Packages = append(new.Packages, r)
		}
	}
	return new
}
