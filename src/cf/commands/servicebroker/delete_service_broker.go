package servicebroker

import (
	"cf/api"
	"cf/configuration"
	"cf/errors"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteServiceBroker struct {
	ui     terminal.UI
	config configuration.Reader
	repo   api.ServiceBrokerRepository
}

func NewDeleteServiceBroker(ui terminal.UI, config configuration.Reader, repo api.ServiceBrokerRepository) (cmd DeleteServiceBroker) {
	cmd.ui = ui
	cmd.config = config
	cmd.repo = repo
	return
}

func (cmd DeleteServiceBroker) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-service-broker")
		return
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd DeleteServiceBroker) Run(c *cli.Context) {
	brokerName := c.Args()[0]
	if !c.Bool("f") && !cmd.ui.ConfirmDelete("service-broker", brokerName) {
		return
	}

	cmd.ui.Say("Deleting service broker %s as %s...",
		terminal.EntityNameColor(brokerName),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	broker, apiErr := cmd.repo.FindByName(brokerName)

	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn("Service Broker %s does not exist.", brokerName)
		return
	default:
		cmd.ui.Failed(apiErr.Error())
		return
	}

	apiErr = cmd.repo.Delete(broker.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	return
}
