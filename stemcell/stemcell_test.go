package stemcell_test

import (
	. "github.com/cloudfoundry/bosh-cli/stemcell"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	"gopkg.in/yaml.v2"
	"path/filepath"
)

var _ = Describe("Stemcell", func() {
	var (
		stemcell      ExtractedStemcell
		manifest      Manifest
		fakefs        *fakesys.FakeFileSystem
		extractedPath string
	)

	BeforeEach(func() {
		manifest = Manifest{}

		extractedPath = "fake-path"

		fakefs = fakesys.NewFakeFileSystem()

		stemcell = NewExtractedStemcell(
			manifest,
			extractedPath,
			fakefs,
		)
	})

	Describe("Manifest", func() {
		It("returns the manifest", func() {
			Expect(stemcell.Manifest()).To(Equal(manifest))
		})
	})

	Describe("Delete", func() {
		var removeAllCalled bool

		BeforeEach(func() {
			fakefs.RemoveAllStub = func(_ string) error {
				removeAllCalled = true
				return nil
			}
		})

		It("removes the extrated path", func() {
			Expect(stemcell.Delete()).To(BeNil())
			Expect(removeAllCalled).To(BeTrue())
		})
	})

	Describe("String", func() {
		BeforeEach(func() {
			manifest = Manifest{
				Name:    "some-name",
				Version: "some-version",
			}

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				fakefs,
			)
		})

		It("returns the name and the version", func() {
			Expect(stemcell.String()).To(Equal("ExtractedStemcell{name=some-name version=some-version}"))
		})
	})

	Describe("OsAndVersion", func() {
		BeforeEach(func() {
			manifest = Manifest{
				OS:      "some-os",
				Version: "some-version",
			}

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				fakefs,
			)
		})

		It("returns the name and the version", func() {
			Expect(stemcell.OsAndVersion()).To(Equal("some-os/some-version"))
		})
	})

	Describe("SetName", func() {
		var newStemcellName string

		BeforeEach(func() {
			manifest = Manifest{
				Name:    "some-name",
				Version: "some-version",
			}

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				fakefs,
			)

			newStemcellName = "some-new-name"
		})

		It("sets the name", func() {
			stemcell.SetName(newStemcellName)
			Expect(stemcell.Manifest().Name).To(Equal(newStemcellName))
		})
	})

	Describe("SetVersion", func() {
		var newStemcellVersion string

		BeforeEach(func() {
			manifest = Manifest{
				Name:    "some-name",
				Version: "some-version",
			}

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				fakefs,
			)

			newStemcellVersion = "some-new-version"
		})

		It("sets the name", func() {
			stemcell.SetVersion(newStemcellVersion)
			Expect(stemcell.Manifest().Version).To(Equal(newStemcellVersion))
		})
	})

	Describe("SetCloudProperties", func() {
		var newStemcellCloudProperties string

		BeforeEach(func() {
			manifest = Manifest{CloudProperties: biproperty.Map{}}

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				fakefs,
			)

			newStemcellCloudProperties = `---
some_property: some_value
`

		})

		It("sets the properties", func() {
			stemcell.SetCloudProperties(newStemcellCloudProperties)
			Expect(stemcell.Manifest().CloudProperties["some_property"]).To(Equal("some_value"))
		})

		Context("there are existing properties in the MF file", func() {
			BeforeEach(func() {

				cloudProperties := biproperty.Map{
					"some_property":     "to be overwritten",
					"existing_property": "this should stick around",
				}

				manifest = Manifest{CloudProperties: cloudProperties}

				stemcell = NewExtractedStemcell(
					manifest,
					extractedPath,
					fakefs,
				)

				newStemcellCloudProperties = `---
some_property: totally overwritten, dude
new_property: didn't previously exist
`
			})

			It("overwrites existing properties", func() {
				stemcell.SetCloudProperties(newStemcellCloudProperties)
				Expect(stemcell.Manifest().CloudProperties["some_property"]).To(Equal("totally overwritten, dude"))
			})

			It("does not remove existing properties", func() {
				stemcell.SetCloudProperties(newStemcellCloudProperties)
				Expect(stemcell.Manifest().CloudProperties["existing_property"]).To(Equal("this should stick around"))
			})

			It("adds new properties", func() {
				stemcell.SetCloudProperties(newStemcellCloudProperties)
				Expect(stemcell.Manifest().CloudProperties["new_property"]).To(Equal("didn't previously exist"))
			})
		})
	})

	Describe("Save", func() {
		var (
			writtenManifest Manifest
			stemcellMfPath  string
		)

		BeforeEach(func() {
			manifest = Manifest{
				ImagePath: "fake-image-path",
				Name:      "fake-stemcell-name",
				Version:   "3312.12",
				OS:        "centos-7",
				SHA1:      "fake-sha",
				CloudProperties: biproperty.Map{
					"infrastructure": "vsphere",
				},
			}

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				fakefs,
			)
			stemcellMfPath = filepath.Join(extractedPath, "stemcell.MF")
		})

		It("writes the stemcell.MF in the extracted director", func() {
			err := stemcell.Save()
			Expect(err).To(Not(HaveOccurred()))
			Expect(fakefs.FileExists(stemcellMfPath)).To(BeTrue())

			manifestContents, err := fakefs.ReadFile(stemcellMfPath)
			yaml.Unmarshal(manifestContents, &writtenManifest)

			Expect(manifest).To(Equal(writtenManifest))
		})
		Context("when it can't write the stemcell.MF file", func() {
			It("returns an error", func() {
				fakefs.WriteFileError = errors.New("fake-write-file-error")
				err := stemcell.Save()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-write-file-error"))
			})
		})
	})
})
