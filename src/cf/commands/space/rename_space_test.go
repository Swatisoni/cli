package space_test

import (
	. "cf/commands/space"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("rename-space command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          configuration.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
		spaceRepo           *testapi.FakeSpaceRepository
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		spaceRepo = &testapi.FakeSpaceRepository{}
	})

	var callRenameSpace = func(args []string) {
		cmd := NewRenameSpace(ui, configRepo, spaceRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("create-space", args), requirementsFactory)
	}

	Describe("when the user is not logged in", func() {
		It("does not pass requirements", func() {
			requirementsFactory.LoginSuccess = false
			callRenameSpace([]string{"my-space", "my-new-space"})
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Describe("when the user has not targeted an org", func() {
		It("does not pass requirements", func() {
			requirementsFactory.TargetedOrgSuccess = false
			callRenameSpace([]string{"my-space", "my-new-space"})
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Describe("when the user provides fewer than two args", func() {
		It("fails with usage", func() {
			callRenameSpace([]string{"foo"})
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Describe("when the user is logged in and has provided an old and new space name", func() {
		BeforeEach(func() {
			space := models.Space{}
			space.Name = "the-old-space-name"
			space.Guid = "the-old-space-guid"
			requirementsFactory.Space = space
		})

		It("renames a space", func() {
			originalSpaceName := configRepo.SpaceFields().Name
			callRenameSpace([]string{"the-old-space-name", "my-new-space"})

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Renaming space", "the-old-space-name", "my-new-space", "my-org", "my-user"},
				{"OK"},
			})

			Expect(spaceRepo.RenameSpaceGuid).To(Equal("the-old-space-guid"))
			Expect(spaceRepo.RenameNewName).To(Equal("my-new-space"))
			Expect(configRepo.SpaceFields().Name).To(Equal(originalSpaceName))
		})

		Describe("renaming the space the user has targeted", func() {
			BeforeEach(func() {
				configRepo.SetSpaceFields(requirementsFactory.Space.SpaceFields)
			})

			It("renames the targeted space", func() {
				callRenameSpace([]string{"the-old-space-name", "my-new-space-name"})
				Expect(configRepo.SpaceFields().Name).To(Equal("my-new-space-name"))
			})
		})
	})
})
