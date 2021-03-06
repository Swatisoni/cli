/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package application_test

import (
	. "cf/commands/application"
	"cf/errors"
	"cf/models"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"time"
)

var _ = Describe("logs command", func() {
	It("fails with usage when called without one argument", func() {
		requirementsFactory, logsRepo := getLogsDependencies()

		ui := callLogs([]string{}, requirementsFactory, logsRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("fails requirements when not logged in", func() {
		requirementsFactory, logsRepo := getLogsDependencies()
		requirementsFactory.LoginSuccess = false

		callLogs([]string{"my-app"}, requirementsFactory, logsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("TestLogsOutputsRecentLogs", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		currentTime := time.Now()

		recentLogs := []*logmessage.LogMessage{
			NewLogMessage("Log Line 1", app.Guid, "DEA", currentTime),
			NewLogMessage("Log Line 2", app.Guid, "DEA", currentTime),
		}

		requirementsFactory, logsRepo := getLogsDependencies()
		requirementsFactory.Application = app
		logsRepo.RecentLogs = recentLogs

		ui := callLogs([]string{"--recent", "my-app"}, requirementsFactory, logsRepo)

		Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
		Expect(app.Guid).To(Equal(logsRepo.AppLoggedGuid))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Connected, dumping recent logs for app", "my-app", "my-org", "my-space", "my-user"},
			{"Log Line 1"},
			{"Log Line 2"},
		})
	})

	It("TestLogsEscapeFormattingVerbs", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		recentLogs := []*logmessage.LogMessage{
			NewLogMessage("hello%2Bworld%v", app.Guid, "DEA", time.Now()),
		}

		requirementsFactory, logsRepo := getLogsDependencies()
		requirementsFactory.Application = app
		logsRepo.RecentLogs = recentLogs

		ui := callLogs([]string{"--recent", "my-app"}, requirementsFactory, logsRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"hello%2Bworld%v"},
		})
	})

	It("TestLogsTailsTheAppLogs", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		logs := []*logmessage.LogMessage{
			NewLogMessage("Log Line 1", app.Guid, "DEA", time.Now()),
		}

		requirementsFactory, logsRepo := getLogsDependencies()
		requirementsFactory.Application = app
		logsRepo.TailLogMessages = logs

		ui := callLogs([]string{"my-app"}, requirementsFactory, logsRepo)

		Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
		Expect(app.Guid).To(Equal(logsRepo.AppLoggedGuid))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Connected, tailing logs for app", "my-app", "my-org", "my-space", "my-user"},
			{"Log Line 1"},
		})
	})

	Context("when the loggregator server has an invalid cert", func() {
		var (
			requirementsFactory *testreq.FakeReqFactory
			logsRepo            *testapi.FakeLogsRepository
		)

		BeforeEach(func() {
			requirementsFactory, logsRepo = getLogsDependencies()
		})

		Context("when the skip-ssl-validation flag is not set", func() {
			It("fails and informs the user about the skip-ssl-validation flag", func() {
				logsRepo.TailLogErr = errors.NewInvalidSSLCert("https://example.com", "it don't work good")
				ui := callLogs([]string{"my-app"}, requirementsFactory, logsRepo)

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Received invalid SSL certificate", "https://example.com"},
					{"TIP"},
				})
			})

			It("informs the user of the error when they include the --recent flag", func() {
				logsRepo.RecentLogErr = errors.NewInvalidSSLCert("https://example.com", "how does SSL work???")
				ui := callLogs([]string{"--recent", "my-app"}, requirementsFactory, logsRepo)

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Received invalid SSL certificate", "https://example.com"},
					{"TIP"},
				})
			})
		})
	})

	Context("when the loggregator server has a valid cert", func() {
		var (
			flags               []string
			requirementsFactory *testreq.FakeReqFactory
			logsRepo            *testapi.FakeLogsRepository
		)

		BeforeEach(func() {
			requirementsFactory, logsRepo = getLogsDependencies()
			flags = []string{"my-app"}
		})

		It("tails logs", func() {
			ui := callLogs(flags, requirementsFactory, logsRepo)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Connected, tailing logs for app", "my-org", "my-space", "my-user"},
			})
		})
	})
})

func getLogsDependencies() (requirementsFactory *testreq.FakeReqFactory, logsRepo *testapi.FakeLogsRepository) {
	logsRepo = &testapi.FakeLogsRepository{}
	requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	return
}

func callLogs(args []string, requirementsFactory *testreq.FakeReqFactory, logsRepo *testapi.FakeLogsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("logs", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewLogs(ui, configRepo, logsRepo)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}
