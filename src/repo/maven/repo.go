package maven

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/joomcode/errorx"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

type Maven struct {
	MainDir  string
	RootDirs []string
}

type Repo struct {
}

type Registry struct {
	Dir string
}

// https://maven.apache.org/ref/3.8.4/maven-repository-metadata/repository-metadata.html
type Metadata struct {
	GroupId     string             `xml:"groupId"`
	ArtifactId  string             `xml:"artifactId"`
	Versioning  MetadataVersioning `xml:"versioning"`
	LastUpdated string             `xml:"lastUpdated"`
}

type MetadataVersioning struct {
	Latest   string   `xml:"latest"`
	Release  string   `xml:"release"` // https://stackoverflow.com/a/53856626 https://stackoverflow.com/questions/5901378/what-exactly-is-a-maven-snapshot-and-why-do-we-need-it https://semver.org/ https://viesure.io/automating-semantic-versioning-with-maven/
	Versions []string `xml:"version"`
}

func (r *Maven) Start(router fiber.Router) error {

	// Create mainDir
	err := os.MkdirAll(r.MainDir, 0700)
	if err != nil && !os.IsExist(err) {
		return errorx.IllegalState.Wrap(err, "Failed to create mainDir")
	}

	for _, dir := range r.RootDirs {

		// TODO: Validate path for security
		err := os.MkdirAll(path.Join(r.MainDir, dir), 0700)
		if err != nil && !os.IsExist(err) {
			return errorx.IllegalState.Wrap(err, "Failed to create rootDir %s", dir)
		}

		registry := &Registry{
			Dir: path.Join(r.MainDir, dir),
		}

		router.Put(fmt.Sprintf("%s+", dir), registry.updateFile)
		router.Get(fmt.Sprintf("%s+", dir), registry.retrieveMetadata)
	}

	return nil
}

func (r *Registry) updateMetadata(packageDir string) error {

	/*
		pkg := strings.TrimPrefix(packageDir, r.Dir)

		meta := Metadata{
			GroupId:    strings.ReplaceAll(pkg[:strings.LastIndex(pkg, "/")], "/", "."),
			ArtifactId: pkg[strings.LastIndex(pkg, "/")+1:],
			Versioning: MetadataVersioning{
				Latest:  "0.0.0",
				Release: "0.0.0",
			},
			LastUpdated: "",
		}

		// Encode xml and write to file named "Meow.xml"
		output, err := xml.MarshalIndent(meta, "  ", "    ")
		if err != nil {
			return errorx.IllegalState.Wrap(err, "Failed to marshal metadata")
		}

		err = os.WriteFile(path.Join(packageDir, "maven-metadata.xml"), output, 0700)

		//fmt.Println(meta)

		//fmt.Println("Updating metadata for", pkg)

	*/
	return nil
}

func (r *Registry) retrieveMetadata(c *fiber.Ctx) error {

	filePath := strings.TrimPrefix(c.Path(), strings.TrimSuffix(c.Route().Path, "+"))
	filePath = strings.TrimPrefix(filePath, "/")
	filePath = strings.ReplaceAll(filePath, "..", "")
	filePath = path.Join(r.Dir, filePath)

	// Make sure file name is maven-metadata.xml
	if !strings.HasSuffix(filePath, "/maven-metadata.xml") {
		return fiber.ErrNotFound
	}

	// Make sure file exists
	if _, err := os.Stat(filePath); err != nil {
		log.Println(errorx.IllegalState.Wrap(err, "Failed to stat file %s", filePath).Error())
		return fiber.ErrNotFound
	}

	err := c.SendFile(filePath, true)
	if err != nil {
		log.Println(errorx.IllegalState.Wrap(err, "Failed to send file %s", filePath).Error())
		return fiber.ErrInternalServerError
	}

	return nil
}
func (r *Registry) updateFile(c *fiber.Ctx) error {

	outputPath := strings.TrimPrefix(c.Path(), strings.TrimSuffix(c.Route().Path, "+"))
	outputPath = strings.TrimPrefix(outputPath, "/")
	outputPath = strings.ReplaceAll(outputPath, "..", "")
	outputPath = path.Join(r.Dir, outputPath)

	parentFolder := outputPath[:strings.LastIndex(outputPath, "/")]
	err := os.MkdirAll(parentFolder, 0700)
	if err != nil {
		return errorx.IllegalState.Wrap(err, "Failed to create dir %s", parentFolder)
	}

	writer, err := os.Create(outputPath)
	if err != nil {
		return errorx.IllegalState.Wrap(err, "Failed to create file %s", outputPath)
	}

	defer writer.Close()

	reader := c.Context().RequestBodyStream()
	buffer := make([]byte, 1024*1024)

	for {

		numBytes, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return errorx.IllegalState.Wrap(err, "Failed to read body")
		}

		if numBytes == 0 {
			break
		}

		_, err = writer.Write(buffer[:numBytes])
		if err != nil {
			return errorx.IllegalState.Wrap(err, "Failed to write file %s", outputPath)
		}

	}

	artifactDir := parentFolder[:strings.LastIndex(parentFolder, "/")]
	err = r.updateMetadata(artifactDir)
	if err != nil {
		return errorx.IllegalState.Wrap(err, "Failed to update metadata")
	}

	return c.SendStatus(200)
}
